#!/bin/bash

# Setup Cloud Monitoring metrics for Ligain API
# This script creates log-based distribution metrics for tracking API performance
# Safe to run multiple times (idempotent)

set -e

# Get environment from argument (default: dev)
ENV=${1:-dev}

# Determine project ID based on environment
if [ "$ENV" = "prd" ]; then
    PROJECT_ID="prd-ligain"
    echo "📊 Setting up metrics for PRODUCTION environment"
else
    PROJECT_ID="woven-century-307314"
    echo "📊 Setting up metrics for DEV environment"
fi

# Metric name
METRIC_NAME="http_request_duration"

echo "Project: $PROJECT_ID"
echo "Metric: $METRIC_NAME"
echo ""

# Check if metric already exists
echo "Checking if metric already exists..."
if gcloud logging metrics describe "$METRIC_NAME" --project="$PROJECT_ID" &> /dev/null; then
    echo "✅ Metric '$METRIC_NAME' already exists - skipping creation"
    echo ""
    echo "To update the metric, delete it first:"
    echo "  gcloud logging metrics delete $METRIC_NAME --project=$PROJECT_ID"
    exit 0
fi

echo "Creating new metric '$METRIC_NAME'..."

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="$SCRIPT_DIR/metrics-config.yaml"

# Create log-based metric for request duration using config file
gcloud logging metrics create "$METRIC_NAME" \
    --project="$PROJECT_ID" \
    --config-from-file="$CONFIG_FILE"

echo "✅ Metric created successfully!"
echo ""
echo "Note: Labels (route, method, status) are automatically extracted from log fields"
echo ""
echo "Next steps:"
echo "1. Wait 2-3 minutes for data to appear"
echo "2. View in Cloud Monitoring: https://console.cloud.google.com/monitoring/metrics-explorer?project=$PROJECT_ID"
echo "3. Search for: logging.googleapis.com/user/$METRIC_NAME"
echo "4. See backend/METRICS.md for dashboard setup instructions"
echo ""
