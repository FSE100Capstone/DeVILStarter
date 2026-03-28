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
	"sync"
	"time"

	"github.com/hashicorp/terraform-exec/tfexec"
)

type orchestrationState struct {
	appDataDir string
	infraPath  string
	tfExecPath string
	nodeRoot   string
	npmCmd     string
	tf         *tfexec.Terraform
}

var (
	initOnce  sync.Once
	initErr   error
	initState orchestrationState
)

func Initialize(ctx context.Context) error {
	initOnce.Do(func() {
		tools.ProgressLogOrchestration(ctx, "Creating local app data folder...", 0)
		appDataDir := appdata.CreateAppDataFolders(ctx)

		tools.ProgressLogOrchestration(ctx, "Downloading Terraform...", 20)
		tfFolderDir := filepath.Join(appDataDir, "terraform")
		if err := os.MkdirAll(tfFolderDir, 0755); err != nil {
			initErr = err
			return
		}

		tfExecPath := tools.DownloadTerraform(ctx, tfFolderDir)
		tools.ProgressLogOrchestration(ctx, fmt.Sprintf("Terraform downloaded to %s!", tfExecPath), 30)

		infraPath := filepath.Join(appDataDir, "infra")
		tools.ProgressLogOrchestration(ctx, "Cloning/updating DeVILSona infrastructure code...", 40)
		if err := tools.CloneOrUpdateRepo(infraPath, "https://github.com/FSE100Capstone/DeVILSona-infra"); err != nil {
			initErr = err
			return
		}

		tools.ProgressLogOrchestration(ctx, "Initializing Terraform...", 55)
		tf, err := tfexec.NewTerraform(infraPath, tfExecPath)
		if err != nil {
			initErr = err
			return
		}

		if err := tf.Init(ctx, tfexec.Upgrade(true)); err != nil {
			initErr = err
			return
		}

		tools.ProgressLogOrchestration(ctx, "Downloading portable Node.js...", 70)
		nodeRoot, npmCmd := tools.EnsurePortableNode(ctx, filepath.Join(appDataDir, "node"))
		tools.ProgressLogOrchestration(ctx, fmt.Sprintf("Portable Node.js downloaded to %s!", nodeRoot), 80)

		tools.ProgressLogOrchestration(ctx, "Executing 'npm install' for each Lambda function...", 90)
		if err := tools.RunNpmInstallForLambdaFolders(ctx, npmCmd, filepath.Join(infraPath, "lambda")); err != nil {
			initErr = err
			return
		}

		initState = orchestrationState{
			appDataDir: appDataDir,
			infraPath:  infraPath,
			tfExecPath: tfExecPath,
			nodeRoot:   nodeRoot,
			npmCmd:     npmCmd,
			tf:         tf,
		}

		tools.ProgressLogOrchestration(ctx, "Initialization complete.", 100)
	})

	if initErr != nil {
		tools.ProgressLogOrchestration(ctx, fmt.Sprintf("Initialization failed: %v", initErr), 0)
	}

	return initErr
}

func getOrchestrationState(ctx context.Context) (*orchestrationState, error) {
	if err := Initialize(ctx); err != nil {
		return nil, err
	}

	return &initState, nil
}

func CreateInfrastructure(ctx context.Context) string {
	tools.ProgressLogOrchestration(ctx, "Waiting for user to authenticate with AWS SSO...", 0)
	creds := aws.GetAWSCredentials(ctx)
	fmt.Println("\n--- AWS Credentials Retrieved ---")
	fmt.Printf("Access Key ID: %s\n", *creds.AccessKeyId)
	fmt.Printf("Secret Access Key: %s\n", *creds.SecretAccessKey)
	fmt.Printf("Session Token: %s\n", *creds.SessionToken)
	fmt.Printf("Expiration: %v\n", time.UnixMilli(creds.Expiration))

	state, err := getOrchestrationState(ctx)
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
	tools.ProgressLogOrchestration(ctx, "Setting secure environment variables for Terraform...", 20)
	err = state.tf.SetEnv(envVars)
	if err != nil {
		log.Fatalf("Failed to set secure environment variables for Terraform: %v", err)
	}

	tools.ProgressLogOrchestration(ctx, "Planning Terraform...", 40)
	planOutputPath := filepath.Join(state.appDataDir, "plan.out")
	_, err = state.tf.Plan(ctx, tfexec.Out(planOutputPath))
	if err != nil {
		log.Fatal(err)
	}

	tools.ProgressLogOrchestration(ctx, "Applying Terraform (this may take a minute)...", 70)
	err = state.tf.Apply(ctx, tfexec.DirOrPlan(planOutputPath))
	if err != nil {
		log.Fatal(err)
	}

	tfOutputs, err := state.tf.Output(ctx)
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

	tools.ProgressLogOrchestration(ctx, "Infrastructure deployed successfully!", 100)
	return outputURL
}

func DestroyInfrastructure(ctx context.Context) {
	creds := aws.GetAWSCredentials(ctx)

	state, err := getOrchestrationState(ctx)
	if err != nil {
		log.Fatal(err)
	}

	envVars := map[string]string{
		"AWS_ACCESS_KEY_ID":     *creds.AccessKeyId,
		"AWS_SECRET_ACCESS_KEY": *creds.SecretAccessKey,
		"AWS_SESSION_TOKEN":     *creds.SessionToken,
	}

	tools.ProgressLogOrchestration(ctx, "Setting secure environment variables for Terraform...", 20)
	err = state.tf.SetEnv(envVars)
	if err != nil {
		log.Fatalf("Failed to set secure environment variables for Terraform: %v", err)
	}

	tools.ProgressLogOrchestration(ctx, "Destroying Terraform...", 70)
	err = state.tf.Destroy(ctx)
	if err != nil {
		log.Fatal(err)
	}

	tools.ProgressLogOrchestration(ctx, "Infrastructure destroyed successfully!", 100)
}

func IsInfrastructureDeployed(ctx context.Context) bool {
	state, err := getOrchestrationState(ctx)
	if err != nil {
		return false
	}

	tools.ProgressLogOrchestration(ctx, "Checking existing infrastructure...", 95)

	outputs, err := state.tf.Output(ctx)
	if err != nil {
		tools.ProgressLogOrchestration(ctx, fmt.Sprintf("Failed to read terraform outputs: %v", err), 0)
		return false
	}

	outputMeta, ok := outputs["session_api_base_url"]
	if !ok {
		return false
	}

	var outputURL string
	if err := json.Unmarshal(outputMeta.Value, &outputURL); err != nil {
		tools.ProgressLogOrchestration(ctx, fmt.Sprintf("Failed to parse output 'session_api_base_url': %v", err), 0)
		return false
	}

	tools.ProgressLogOrchestration(ctx, "Infrastructure check complete.", 100)

	return outputURL != ""
}
