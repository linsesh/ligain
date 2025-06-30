#!/bin/bash

PROJECT_ID="woven-century-307314"
REGION="europe-west1"
SERVICE_NAME="server-dev"

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
  *)
    echo "Usage: $0 {start|stop|status}"
    exit 1
    ;;
esac