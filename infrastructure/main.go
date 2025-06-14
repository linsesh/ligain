package main

import (
	"encoding/base64"
	"fmt"

	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/cloudrun"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/cloudscheduler"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Get the current stack name (dev or prod)
		stack := ctx.Stack()

		// Configuration
		cfg := config.New(ctx, "gcp")
		projectID := cfg.Require("project")
		region := cfg.Require("region")
		serviceName := fmt.Sprintf("server-%s", stack)

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
									ContainerPort: pulumi.Int(3000),
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
									Name:  pulumi.String("NODE_ENV"),
									Value: pulumi.String(stack),
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

		// Only create scheduler jobs for dev environment
		if stack == "dev" {
			// Create Cloud Scheduler job to start the service at 4 PM
			startBody := `{"spec":{"template":{"metadata":{"annotations":{"autoscaling.knative.dev/minScale":"1"}}}}}`
			startBodyBase64 := base64.StdEncoding.EncodeToString([]byte(startBody))

			_, err = cloudscheduler.NewJob(ctx, fmt.Sprintf("%s-start", serviceName), &cloudscheduler.JobArgs{
				Schedule:    pulumi.String("0 16 * * *"), // 4 PM every day
				TimeZone:    pulumi.String("Europe/Paris"),
				Description: pulumi.String("Start the Cloud Run service"),
				HttpTarget: &cloudscheduler.JobHttpTargetArgs{
					Uri:        pulumi.Sprintf("https://%s-run.googleapis.com/apis/serving.knative.dev/v1/namespaces/%s/services/%s", region, projectID, serviceName),
					HttpMethod: pulumi.String("PATCH"),
					Headers: pulumi.StringMap{
						"Content-Type": pulumi.String("application/json"),
					},
					Body: pulumi.String(startBodyBase64),
					OauthToken: &cloudscheduler.JobHttpTargetOauthTokenArgs{
						ServiceAccountEmail: pulumi.Sprintf("cloud-scheduler@%s.iam.gserviceaccount.com", projectID),
					},
				},
			})
			if err != nil {
				return err
			}

			// Create Cloud Scheduler job to stop the service at 10 PM
			stopBody := `{"spec":{"template":{"metadata":{"annotations":{"autoscaling.knative.dev/minScale":"0"}}}}}`
			stopBodyBase64 := base64.StdEncoding.EncodeToString([]byte(stopBody))

			_, err = cloudscheduler.NewJob(ctx, fmt.Sprintf("%s-stop", serviceName), &cloudscheduler.JobArgs{
				Schedule:    pulumi.String("0 22 * * *"), // 10 PM every day
				TimeZone:    pulumi.String("Europe/Paris"),
				Description: pulumi.String("Stop the Cloud Run service"),
				HttpTarget: &cloudscheduler.JobHttpTargetArgs{
					Uri:        pulumi.Sprintf("https://%s-run.googleapis.com/apis/serving.knative.dev/v1/namespaces/%s/services/%s", region, projectID, serviceName),
					HttpMethod: pulumi.String("PATCH"),
					Headers: pulumi.StringMap{
						"Content-Type": pulumi.String("application/json"),
					},
					Body: pulumi.String(stopBodyBase64),
					OauthToken: &cloudscheduler.JobHttpTargetOauthTokenArgs{
						ServiceAccountEmail: pulumi.Sprintf("cloud-scheduler@%s.iam.gserviceaccount.com", projectID),
					},
				},
			})
			if err != nil {
				return err
			}
		}

		// Allow unauthenticated access to the service
		_, err = cloudrun.NewIamMember(ctx, fmt.Sprintf("%s-public", serviceName), &cloudrun.IamMemberArgs{
			Location: pulumi.String(region),
			Service:  service.Name,
			Role:     pulumi.String("roles/run.invoker"),
			Member:   pulumi.String("allUsers"),
		})
		if err != nil {
			return err
		}

		// Export the service URL
		ctx.Export("serviceUrl", service.Statuses.Index(pulumi.Int(0)).Url())

		return nil
	})
}

func getMemoryLimit(stack string) string {
	// Same minimal resources for both environments
	return "64Mi"
}

func getCPULimit(stack string) string {
	// Same minimal resources for both environments
	return "100m"
}

func getMinScale(stack string) string {
	if stack == "dev" {
		return "0" // Dev starts with 0 instances and is scheduled
	}
	return "1" // Prod always runs
}

func getMaxScale(stack string) string {
	if stack == "dev" {
		return "2" // Dev environment only needs 2 instances max
	}
	return "10" // Prod can scale higher if needed
}
