package orchestration

import (
	"DeVILStarter/internal/appdata"
	"DeVILStarter/internal/aws"
	"DeVILStarter/internal/tools"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func logOrchestration(ctx context.Context, message string) {
	log.Println(message)
	runtime.EventsEmit(ctx, "orchestrationLog", message)
}

func CreateInfrastructure(ctx context.Context) string {
	logOrchestration(ctx, "Waiting for user to authenticate with AWS SSO...")
	creds := aws.GetAWSCredentials(ctx)
	fmt.Println("\n--- AWS Credentials Retrieved ---")
	fmt.Printf("Access Key ID: %s\n", *creds.AccessKeyId)
	fmt.Printf("Secret Access Key: %s\n", *creds.SecretAccessKey)
	fmt.Printf("Session Token: %s\n", *creds.SessionToken)
	fmt.Printf("Expiration: %v\n", time.UnixMilli(creds.Expiration))

	logOrchestration(ctx, "Creating local app data folder...")
	appDataDir := appdata.CreateAppDataFolders(ctx)

	logOrchestration(ctx, "Downloading Terraform...")
	tfFolderDir := filepath.Join(appDataDir, "terraform")
	err := os.MkdirAll(tfFolderDir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	tfExecPath := tools.DownloadTerraform(ctx, tfFolderDir)

	logOrchestration(ctx, fmt.Sprintf("Terraform downloaded to %s!\n", tfExecPath))

	infraPath := filepath.Join(appDataDir, "infra")
	logOrchestration(ctx, "Cloning/updating DeVILSona infrastructure code...")
	err = tools.CloneOrUpdateRepo(infraPath, "https://github.com/FSE100Capstone/DeVILSona-infra")
	if err != nil {
		log.Fatal(err)
	}

	// Terraform init using terraform-exec
	logOrchestration(ctx, "Initializing Terraform...")
	tf, err := tfexec.NewTerraform(infraPath, tfExecPath)
	if err != nil {
		log.Fatal(err)
	}

	err = tf.Init(ctx, tfexec.Upgrade(true))
	if err != nil {
		log.Fatal(err)
	}

	// Download a portable Node.js (includes npm) and run npm install
	logOrchestration(ctx, "Downloading portable Node.js...")
	nodeRoot, npmCmd := tools.EnsurePortableNode(ctx, filepath.Join(appDataDir, "node"))
	logOrchestration(ctx, fmt.Sprintf("Portable Node.js downloaded to %s!", nodeRoot))

	// For every folder in infra/lambda, run npm install if package.json exists
	logOrchestration(ctx, "Executing 'npm install' for each Lambda function...")
	err = tools.RunNpmInstallForLambdaFolders(ctx, npmCmd, filepath.Join(infraPath, "lambda"))
	if err != nil {
		log.Fatal(err)
	}

	// 1. Create a map of the required AWS environment variables
	// We use the temporary STS credentials retrieved from the SSO flow.
	envVars := map[string]string{
		"AWS_ACCESS_KEY_ID":     *creds.AccessKeyId,
		"AWS_SECRET_ACCESS_KEY": *creds.SecretAccessKey,
		"AWS_SESSION_TOKEN":     *creds.SessionToken,
	}

	// 2. Inject them into the Terraform execution context.
	// This is highly secure because it DOES NOT set these globally on your host OS.
	// It only makes them available to the specific child process running Terraform.
	logOrchestration(ctx, "Setting secure environment variables for Terraform...")
	err = tf.SetEnv(envVars)
	if err != nil {
		log.Fatalf("Failed to set secure environment variables for Terraform: %v", err)
	}

	logOrchestration(ctx, "Planning Terraform...")
	planOutputPath := filepath.Join(appDataDir, "plan.out")
	_, err = tf.Plan(ctx, tfexec.Out(planOutputPath))
	if err != nil {
		log.Fatal(err)
	}

	logOrchestration(ctx, "Applying Terraform (this may take a minute)...")
	err = tf.Apply(ctx, tfexec.DirOrPlan(planOutputPath))
	if err != nil {
		log.Fatal(err)
	}

	tfOutputs, err := tf.Output(ctx)
	if err != nil {
		log.Fatal(err)
	}

	outputMeta, ok := tfOutputs["session_api_base_url"]
	if !ok {
		log.Fatal("Expected output 'session_api_base_url' not found")
	}

	var outputURL string
	err = json.Unmarshal(outputMeta.Value, &outputURL)
	if err != nil {
		log.Fatal("Failed to unmarshal output 'session_api_base_url':", err)
	}

	logOrchestration(ctx, "Infrastructure deployed successfully!")
	return outputURL
}

func DestroyInfrastructure(ctx context.Context) {
	creds := aws.GetAWSCredentials(ctx)

	logOrchestration(ctx, "Creating local app data folder")
	appDataDir := appdata.CreateAppDataFolders(ctx)

	logOrchestration(ctx, "Downloading/fetching Terraform...")
	tfFolderDir := filepath.Join(appDataDir, "terraform")
	err := os.MkdirAll(tfFolderDir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	tfExecPath := tools.DownloadTerraform(ctx, tfFolderDir)

	infraPath := filepath.Join(appDataDir, "infra")
	logOrchestration(ctx, "Cloning/updating GitHub repo...")
	err = tools.CloneOrUpdateRepo(infraPath, "https://github.com/FSE100Capstone/DeVILSona-infra")
	if err != nil {
		log.Fatal(err)
	}

	logOrchestration(ctx, "Initializing Terraform...")
	tf, err := tfexec.NewTerraform(infraPath, tfExecPath)
	if err != nil {
		log.Fatal(err)
	}

	err = tf.Init(ctx, tfexec.Upgrade(true))
	if err != nil {
		log.Fatal(err)
	}

	envVars := map[string]string{
		"AWS_ACCESS_KEY_ID":     *creds.AccessKeyId,
		"AWS_SECRET_ACCESS_KEY": *creds.SecretAccessKey,
		"AWS_SESSION_TOKEN":     *creds.SessionToken,
	}

	err = tf.SetEnv(envVars)
	if err != nil {
		log.Fatalf("Failed to set secure environment variables for Terraform: %v", err)
	}

	logOrchestration(ctx, "Destroying Terraform...")
	err = tf.Destroy(ctx)
	if err != nil {
		log.Fatal(err)
	}

	logOrchestration(ctx, "Infrastructure destroyed successfully!")
}
