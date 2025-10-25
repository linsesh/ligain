#!/bin/bash

# Setup Cloud Monitoring Dashboard for Ligain API
# This script creates a comprehensive dashboard for API performance metrics
# Safe to run multiple times (idempotent)

set -e

# Get environment from argument (default: dev)
ENV=${1:-dev}

# Determine project ID based on environment
if [ "$ENV" = "prd" ]; then
    PROJECT_ID="prd-ligain"
    echo "📊 Setting up dashboard for PRODUCTION environment"
else
    PROJECT_ID="woven-century-307314"
    echo "📊 Setting up dashboard for DEV environment"
fi

# Dashboard name
DASHBOARD_NAME="ligain-api-performance"

echo "Project: $PROJECT_ID"
echo "Dashboard: $DASHBOARD_NAME"
echo ""

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DASHBOARD_FILE="$SCRIPT_DIR/dashboard.json"

# Check if dashboard already exists
echo "Checking if dashboard already exists..."
if gcloud monitoring dashboards list --project="$PROJECT_ID" --filter="displayName:$DASHBOARD_NAME" --format="value(name)" | grep -q "dashboards"; then
    echo "✅ Dashboard '$DASHBOARD_NAME' already exists"
    echo ""
    echo "To update the dashboard, delete it first:"
    echo "  gcloud monitoring dashboards delete \$(gcloud monitoring dashboards list --project=$PROJECT_ID --filter=\"displayName:$DASHBOARD_NAME\" --format=\"value(name)\")"
    echo "  Then run this script again"
    exit 0
fi

echo "Creating dashboard '$DASHBOARD_NAME'..."

# Create the dashboard
gcloud monitoring dashboards create \
    --project="$PROJECT_ID" \
    --config-from-file="$DASHBOARD_FILE"

echo "✅ Dashboard created successfully!"
echo ""
echo "View your dashboard:"
echo "https://console.cloud.google.com/monitoring/dashboards?project=$PROJECT_ID"
echo ""
echo "Dashboard includes:"
echo "- P50, P90, P99 response times by route"
echo "- Mean response time by route"
echo "- Request count by status code"
echo "- Response time distribution heatmap"
echo ""
