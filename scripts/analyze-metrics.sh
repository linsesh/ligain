#!/bin/bash

# Analyze API Performance Metrics from Cloud Logging
# This script extracts logs and computes performance metrics (mean, p50, p90, p99, etc.)
# Usage: ./analyze-metrics.sh [ENV] [START_TIME] [END_TIME]
# Example: ./analyze-metrics.sh prd "2025-10-24T10:00:00Z" "2025-10-24T11:00:00Z"

set -e

# Get parameters
ENV=${1:-prd}
START_TIME=${2:-$(date -u -v-1H +%Y-%m-%dT%H:%M:%SZ)}
END_TIME=${3:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}

# Determine project ID based on environment
if [ "$ENV" = "prd" ]; then
    PROJECT_ID="prd-ligain"
    echo "📊 Analyzing metrics for PRODUCTION environment"
else
    PROJECT_ID="woven-century-307314"
    echo "📊 Analyzing metrics for DEV environment"
fi

echo "Project: $PROJECT_ID"
echo "Time range: $START_TIME to $END_TIME"
echo ""

# Create temporary files
TEMP_LOGS="/tmp/ligain_logs_$$.json"
TEMP_ANALYSIS="/tmp/ligain_analysis_$$.json"

# Function to cleanup temp files
cleanup() {
    rm -f "$TEMP_LOGS" "$TEMP_ANALYSIS"
}
trap cleanup EXIT

echo "🔍 Extracting logs from Cloud Logging..."

# Extract logs with the specified time range
gcloud logging read \
    --project="$PROJECT_ID" \
    --format=json \
    --limit=10000 \
    'jsonPayload.metric_type="http_request"' \
    > "$TEMP_LOGS"

# Check if we got any logs
LOG_COUNT=$(jq length "$TEMP_LOGS")
if [ "$LOG_COUNT" -eq 0 ]; then
    echo "❌ No logs found in the specified time range"
    exit 1
fi

echo "✅ Found $LOG_COUNT log entries"
echo ""

echo "📊 Computing performance metrics..."

# Process logs and compute metrics
jq -r --arg start_time "$START_TIME" --arg end_time "$END_TIME" '
# Filter by time range first
map(select(.timestamp >= $start_time and .timestamp <= $end_time)) |
# Group by route and compute statistics
group_by(.jsonPayload.route) | 
map({
  route: .[0].jsonPayload.route,
  method: .[0].jsonPayload.method,
  count: length,
  durations: [.[] | .jsonPayload.duration_ms],
  status_codes: [.[] | .jsonPayload.status]
}) |
map({
  route: .route,
  method: .method,
  count: .count,
  durations: .durations,
  status_codes: .status_codes,
  # Compute statistics
  mean: (.durations | add / length),
  min: (.durations | min),
  max: (.durations | max),
  # Percentiles (simple implementation)
  p50: (.durations | sort | .[length/2 | floor]),
  p90: (.durations | sort | .[length * 0.9 | floor]),
  p95: (.durations | sort | .[length * 0.95 | floor]),
  p99: (.durations | sort | .[length * 0.99 | floor]),
  # Status code distribution
  status_200: (.status_codes | map(select(. == 200)) | length),
  status_4xx: (.status_codes | map(select(. >= 400 and . < 500)) | length),
  status_5xx: (.status_codes | map(select(. >= 500)) | length)
}) |
sort_by(.count) | reverse
' "$TEMP_LOGS" > "$TEMP_ANALYSIS"

echo "📈 Performance Analysis Results:"
echo "=================================="
echo ""

# Display results in a nice format
jq -r '
.[] | 
"Route: \(.route) (\(.method))" +
"\n  Requests: \(.count)" +
"\n  Response Time (ms):" +
"\n    Mean: \(.mean | round)" +
"\n    Min:  \(.min)" +
"\n    Max:  \(.max)" +
"\n    P50:  \(.p50)" +
"\n    P90:  \(.p90)" +
"\n    P95:  \(.p95)" +
"\n    P99:  \(.p99)" +
"\n  Status Codes:" +
"\n    200: \(.status_200) (\(.status_200 / .count * 100 | round)% success)" +
"\n    4xx: \(.status_4xx) (\(.status_4xx / .count * 100 | round)% client errors)" +
"\n    5xx: \(.status_5xx) (\(.status_5xx / .count * 100 | round)% server errors)" +
"\n"
' "$TEMP_ANALYSIS"

echo ""
echo "📊 Summary Statistics:"
echo "======================"

# Overall statistics
TOTAL_REQUESTS=$(jq '[.[] | .count] | add' "$TEMP_ANALYSIS")
TOTAL_SUCCESS=$(jq '[.[] | .status_200] | add' "$TEMP_ANALYSIS")
TOTAL_4XX=$(jq '[.[] | .status_4xx] | add' "$TEMP_ANALYSIS")
TOTAL_5XX=$(jq '[.[] | .status_5xx] | add' "$TEMP_ANALYSIS")

echo "Total Requests: $TOTAL_REQUESTS"
echo "Success Rate: $(( TOTAL_SUCCESS * 100 / TOTAL_REQUESTS ))%"
echo "4xx Errors: $TOTAL_4XX ($(( TOTAL_4XX * 100 / TOTAL_REQUESTS ))%)"
echo "5xx Errors: $TOTAL_5XX ($(( TOTAL_5XX * 100 / TOTAL_REQUESTS ))%)"

# Overall response time statistics
ALL_DURATIONS=$(jq '[.[] | .durations[]] | sort' "$TEMP_ANALYSIS")
TOTAL_MEAN=$(echo "$ALL_DURATIONS" | jq 'add / length | round')
TOTAL_P50=$(echo "$ALL_DURATIONS" | jq '.[length/2 | floor]')
TOTAL_P90=$(echo "$ALL_DURATIONS" | jq '.[length * 0.9 | floor]')
TOTAL_P95=$(echo "$ALL_DURATIONS" | jq '.[length * 0.95 | floor]')
TOTAL_P99=$(echo "$ALL_DURATIONS" | jq '.[length * 0.99 | floor]')

echo ""
echo "Overall Response Times (ms):"
echo "  Mean: $TOTAL_MEAN"
echo "  P50:  $TOTAL_P50"
echo "  P90:  $TOTAL_P90"
echo "  P95:  $TOTAL_P95"
echo "  P99:  $TOTAL_P99"

echo ""
echo "💡 Recommendations:"
echo "==================="

# Check for slow routes
SLOW_ROUTES=$(jq -r '.[] | select(.p95 > 1000) | "  - \(.route): P95 = \(.p95)ms"' "$TEMP_ANALYSIS")
if [ -n "$SLOW_ROUTES" ]; then
    echo "🐌 Slow routes (P95 > 1000ms):"
    echo "$SLOW_ROUTES"
    echo ""
fi

# Check for high error rates
ERROR_ROUTES=$(jq -r '.[] | select((.status_4xx + .status_5xx) / .count > 0.05) | "  - \(.route): \(.status_4xx + .status_5xx)/\(.count) errors"' "$TEMP_ANALYSIS")
if [ -n "$ERROR_ROUTES" ]; then
    echo "⚠️  High error rate routes (>5%):"
    echo "$ERROR_ROUTES"
    echo ""
fi

echo "✅ Analysis complete!"
echo ""
echo "To analyze a different time range:"
echo "  ./analyze-metrics.sh $ENV \"2025-10-24T10:00:00Z\" \"2025-10-24T11:00:00Z\""
echo ""
echo "To analyze the last hour:"
echo "  ./analyze-metrics.sh $ENV"
