package aws

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	sso "github.com/aws/aws-sdk-go-v2/service/sso"
	ssotypes "github.com/aws/aws-sdk-go-v2/service/sso/types"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	ssooidctypes "github.com/aws/aws-sdk-go-v2/service/ssooidc/types"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	ssoStartURL = "https://arizonastateuniversity.awsapps.com/start"
	ssoRegion   = "us-west-2"
	accountID   = "011422532823"
	roleName    = "saco-asafse-prod-adm"
)

func GetAWSCredentials(ctx context.Context) *ssotypes.RoleCredentials {
	// 1. Initialize standard AWS config for the SSO region
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(ssoRegion))
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	oidcClient := ssooidc.NewFromConfig(cfg)

	// 2. Register this Go app as an OIDC public client
	regRes, err := oidcClient.RegisterClient(ctx, &ssooidc.RegisterClientInput{
		ClientName: aws.String("devilstarter"),
		ClientType: aws.String("public"),
	})
	if err != nil {
		log.Fatalf("Failed to register client: %v", err)
	}

	// 3. Start the Device Authorization Flow
	authRes, err := oidcClient.StartDeviceAuthorization(ctx, &ssooidc.StartDeviceAuthorizationInput{
		ClientId:     regRes.ClientId,
		ClientSecret: regRes.ClientSecret,
		StartUrl:     aws.String(ssoStartURL),
	})
	if err != nil {
		log.Fatalf("Failed to start device auth: %v", err)
	}

	// 4. Prompt the user to log in via their browser
	fmt.Printf("\nAction Required: Please open the following URL in your browser:\n")
	fmt.Printf("URL: %s\n", *authRes.VerificationUriComplete)
	fmt.Println("\nWaiting for browser authentication...")

	runtime.BrowserOpenURL(ctx, *authRes.VerificationUriComplete)

	// 5. Poll AWS to see if the user has completed the login
	var tokenRes *ssooidc.CreateTokenOutput
	interval := time.Duration(authRes.Interval) * time.Second

	for {
		time.Sleep(interval)
		tokenRes, err = oidcClient.CreateToken(ctx, &ssooidc.CreateTokenInput{
			ClientId:     regRes.ClientId,
			ClientSecret: regRes.ClientSecret,
			DeviceCode:   authRes.DeviceCode,
			GrantType:    aws.String("urn:ietf:params:oauth:grant-type:device_code"),
		})

		if err == nil {
			// Success! User clicked 'Allow' in the browser.
			break
		}

		// If the error is simply "AuthorizationPendingException", we keep waiting.
		var pendingErr *ssooidctypes.AuthorizationPendingException
		if errors.As(err, &pendingErr) {
			continue
		}

		// If it's a different error (e.g., timeout, denied), fail out.
		log.Fatalf("\nFailed during token polling: %v", err)
	}

	fmt.Println("Browser authentication successful!")

	// 6. Exchange the resulting SSO Access Token for actual AWS Credentials
	ssoClient := sso.NewFromConfig(cfg)
	credsRes, err := ssoClient.GetRoleCredentials(ctx, &sso.GetRoleCredentialsInput{
		AccessToken: tokenRes.AccessToken,
		AccountId:   aws.String(accountID),
		RoleName:    aws.String(roleName),
	})
	if err != nil {
		log.Fatalf("Failed to get role credentials: %v", err)
	}

	return credsRes.RoleCredentials
}
