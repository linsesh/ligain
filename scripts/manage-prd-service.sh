#!/bin/bash

PROJECT_ID="prd-ligain"
REGION="europe-west1"
SERVICE_NAME="server-prd"

case "$1" in
  "start")
    echo "Starting $SERVICE_NAME..."
    gcloud run services update $SERVICE_NAME \
      --region=$REGION \
      --project=$PROJECT_ID \
      --min-instances=1
    ;;
  "stop")
    echo "Stopping $SERVICE_NAME..."
    gcloud run services update $SERVICE_NAME \
      --region=$REGION \
      --project=$PROJECT_ID \
      --min-instances=0
    ;;
  "status")
    echo "Checking status of $SERVICE_NAME..."
    gcloud run services describe $SERVICE_NAME \
      --region=$REGION \
      --project=$PROJECT_ID \
      --format="value(status.conditions[0].status,status.url)"
    ;;
  "is-running")
    echo "Checking if $SERVICE_NAME is running..."
    MIN_INSTANCES=$(gcloud run services describe $SERVICE_NAME \
      --region=$REGION \
      --project=$PROJECT_ID \
      --format="value(spec.template.metadata.annotations['autoscaling.knative.dev/minScale'])" 2>/dev/null)
    
    if [ "$MIN_INSTANCES" = "1" ]; then
      echo "✅ Server is RUNNING (min instances: $MIN_INSTANCES)"
      exit 0
    else
      echo "⏸️  Server is STOPPED (min instances: $MIN_INSTANCES)"
      exit 1
    fi
    ;;
  "url")
    echo "Getting URL for $SERVICE_NAME..."
    gcloud run services describe $SERVICE_NAME \
      --region=$REGION \
      --project=$PROJECT_ID \
      --format="value(status.url)"
    ;;
  "info")
    echo "Detailed info for $SERVICE_NAME..."
    echo "=== Service Status ==="
    gcloud run services describe $SERVICE_NAME \
      --region=$REGION \
      --project=$PROJECT_ID \
      --format="table(
        metadata.name,
        status.conditions[0].status,
        status.url,
        spec.template.metadata.annotations['autoscaling.knative.dev/minScale'],
        spec.template.metadata.annotations['autoscaling.knative.dev/maxScale']
      )"
    ;;
  "allow-public")
    echo "Allowing public access to $SERVICE_NAME..."
    gcloud run services add-iam-policy-binding $SERVICE_NAME \
      --region=$REGION \
      --project=$PROJECT_ID \
      --member="allUsers" \
      --role="roles/run.invoker"
    ;;
  "deny-public")
    echo "Removing public access from $SERVICE_NAME..."
    gcloud run services remove-iam-policy-binding $SERVICE_NAME \
      --region=$REGION \
      --project=$PROJECT_ID \
      --member="allUsers" \
      --role="roles/run.invoker"
    ;;
  "test")
    echo "Testing access to $SERVICE_NAME..."
    URL=$(gcloud run services describe $SERVICE_NAME \
      --region=$REGION \
      --project=$PROJECT_ID \
      --format="value(status.url)" 2>/dev/null)
    
    if [ -n "$URL" ]; then
      echo "Testing URL: $URL"
      curl -s -o /dev/null -w "HTTP Status: %{http_code}\n" "$URL" || echo "Failed to connect"
    else
      echo "Could not get service URL"
    fi
    ;;
  "destroy")
    echo "⚠️  WARNING: This will completely destroy the PRODUCTION Cloud Run service via Pulumi!"
    echo "Service: $SERVICE_NAME"
    echo "Region: $REGION"
    echo "Project: $PROJECT_ID"
    echo ""
    read -p "Are you sure you want to destroy the PRODUCTION service? (yes/no): " confirm
    if [ "$confirm" = "yes" ]; then
      echo "Destroying $SERVICE_NAME via Pulumi..."
      cd infrastructure && pulumi stack select prd && pulumi destroy --yes
      echo "✅ Production service destroyed successfully!"
    else
      echo "❌ Destruction cancelled."
    fi
    ;;
  "deploy")
    echo "Deploying $SERVICE_NAME to PRODUCTION..."
    cd infrastructure && pulumi stack select prd && pulumi up --yes
    ;;
  "build-and-deploy")
    echo "Building and deploying $SERVICE_NAME to PRODUCTION..."
    
    # Build the Docker image from the root directory
    echo "Building Docker image..."
    docker build -t gcr.io/$PROJECT_ID/$SERVICE_NAME:latest -f backend/Dockerfile .
    
    # Push to Google Container Registry
    echo "Pushing to Google Container Registry..."
    docker push gcr.io/$PROJECT_ID/$SERVICE_NAME:latest
    
    # Deploy with Pulumi
    echo "Deploying with Pulumi..."
    cd infrastructure && pulumi stack select prd && pulumi up --yes
    ;;
  *)
    echo "Usage: $0 {start|stop|status|is-running|url|info|allow-public|deny-public|test|destroy|deploy|build-and-deploy}"
    echo ""
    echo "Commands:"
    echo "  start           - Start the production server (set min instances to 1)"
    echo "  stop            - Stop the production server (set min instances to 0)"
    echo "  status          - Show basic status and URL"
    echo "  is-running      - Check if server is running (returns 0 if running, 1 if stopped)"
    echo "  url             - Get just the service URL"
    echo "  info            - Show detailed service information"
    echo "  allow-public    - Allow public access to the service"
    echo "  deny-public     - Remove public access from the service"
    echo "  test            - Test access to the service"
    echo "  destroy         - Destroy the service via Pulumi (complete removal)"
    echo "  deploy          - Deploy the service via Pulumi"
    echo "  build-and-deploy - Build Docker image, push to GCR, and deploy"
    exit 1
    ;;
esac 