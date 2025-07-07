package main

import (
	"fmt"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/cloudrun"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Get the current stack name (dev or prod)
		stack := ctx.Stack()

		// Configuration
		cfg := config.New(ctx, "gcp")
		ligainCfg := config.New(ctx, "ligain")
		projectID := cfg.Require("project")
		region := cfg.Require("region")
		serviceName := fmt.Sprintf("server-%s", stack)

		// Get environment variables from config
		databaseURL := ligainCfg.Require("database_url")
		apiKey := ligainCfg.Require("api_key")
		allowedOrigins := ligainCfg.Require("allowed_origins")
		sportsmonkToken := ligainCfg.Require("sportsmonk_api_token")

		// Create a Cloud Run service
		service, err := cloudrun.NewService(ctx, serviceName, &cloudrun.ServiceArgs{
			Location: pulumi.String(region),
			Template: &cloudrun.ServiceTemplateArgs{
				Metadata: &cloudrun.ServiceTemplateMetadataArgs{
					Annotations: pulumi.StringMap{
						"autoscaling.knative.dev/minScale": pulumi.String(getMinScale(stack)),
						"autoscaling.knative.dev/maxScale": pulumi.String(getMaxScale(stack)),
					},
				},
				Spec: &cloudrun.ServiceTemplateSpecArgs{
					Containers: cloudrun.ServiceTemplateSpecContainerArray{
						&cloudrun.ServiceTemplateSpecContainerArgs{
							Image: pulumi.Sprintf("gcr.io/%s/%s:latest", projectID, serviceName),
							Ports: cloudrun.ServiceTemplateSpecContainerPortArray{
								&cloudrun.ServiceTemplateSpecContainerPortArgs{
									ContainerPort: pulumi.Int(8080), // Fixed: match Dockerfile port
								},
							},
							Resources: &cloudrun.ServiceTemplateSpecContainerResourcesArgs{
								Limits: pulumi.StringMap{
									"memory": pulumi.String(getMemoryLimit(stack)),
									"cpu":    pulumi.String(getCPULimit(stack)),
								},
							},
							Envs: cloudrun.ServiceTemplateSpecContainerEnvArray{
								&cloudrun.ServiceTemplateSpecContainerEnvArgs{
									Name:  pulumi.String("ENV"),
									Value: pulumi.String(stack),
								},
								&cloudrun.ServiceTemplateSpecContainerEnvArgs{
									Name:  pulumi.String("DATABASE_URL"),
									Value: pulumi.String(databaseURL),
								},
								&cloudrun.ServiceTemplateSpecContainerEnvArgs{
									Name:  pulumi.String("API_KEY"),
									Value: pulumi.String(apiKey),
								},
								&cloudrun.ServiceTemplateSpecContainerEnvArgs{
									Name:  pulumi.String("ALLOWED_ORIGINS"),
									Value: pulumi.String(allowedOrigins),
								},
								&cloudrun.ServiceTemplateSpecContainerEnvArgs{
									Name:  pulumi.String("SPORTSMONK_API_TOKEN"),
									Value: pulumi.String(sportsmonkToken),
								},
							},
						},
					},
				},
			},
		})
		if err != nil {
			return err
		}

		// Allow unauthenticated access to the service
		// Commented out to make service private - only authenticated users can access
		/*
			_, err = cloudrun.NewIamMember(ctx, fmt.Sprintf("%s-public", serviceName), &cloudrun.IamMemberArgs{
				Location: pulumi.String(region),
				Service:  service.Name,
				Role:     pulumi.String("roles/run.invoker"),
				Member:   pulumi.String("allUsers"),
			})
			if err != nil {
				return err
			}
		*/

		// Export the service URL
		ctx.Export("serviceUrl", service.Statuses.ApplyT(func(statuses []cloudrun.ServiceStatus) (string, error) {
			if len(statuses) == 0 {
				return "", nil // Return empty string instead of error
			}
			if statuses[0].Url == nil {
				return "", nil // Return empty string instead of error
			}
			return *statuses[0].Url, nil
		}))

		return nil
	})
}

func getMemoryLimit(stack string) string {
	// Minimum memory required for 100m CPU in Cloud Run
	return "128Mi"
}

func getCPULimit(stack string) string {
	// Same minimal resources for both environments
	return "100m"
}

func getMinScale(stack string) string {
	if stack == "dev" {
		return "0" // Dev starts with 0 instances and is managed manually
	}
	return "1" // Prod always runs
}

func getMaxScale(stack string) string {
	if stack == "dev" {
		return "2" // Dev environment only needs 2 instances max
	}
	return "10" // Prod can scale higher if needed
}
