# Performance Metrics Setup for Cloud Run

This guide explains how to set up and view performance metrics (mean, p50, p75, p90, p99) for your API routes on GCP Cloud Run.

## How It Works

1. **MetricsMiddleware** logs structured request data to Cloud Logging
2. **Cloud Monitoring** automatically processes these logs
3. **Distribution metrics** provide automatic percentile calculations

## Setup Steps

### 1. Deploy Your Code

The metrics middleware and metric creation are **fully automated**. Just deploy:

```bash
make deploy  # Dev environment - automatically sets up metrics
# or
make deploy-prd  # Production environment - automatically sets up metrics
```

The deployment will:
1. Build and push your Docker image
2. Deploy to Cloud Run via Pulumi
3. **Automatically create the log-based metric** (idempotent - safe to run multiple times)

### 2. Manual Metric Creation (Optional)

If you need to create the metric separately, you can run:

```bash
# Dev environment
make setup-metrics

# Production environment
make setup-metrics ENV=prd
```

Or use gcloud directly:

```bash
# Create distribution metric for request duration
gcloud logging metrics create http_request_duration \
  --description="HTTP request duration by route" \
  --value-extractor='EXTRACT(jsonPayload.duration_ms)' \
  --metric-kind=DELTA \
  --value-type=DISTRIBUTION \
  --log-filter='jsonPayload.metric_type="http_request"'
```

### 3. View Metrics in Cloud Monitoring

After deploying and generating some traffic:

1. Go to **Cloud Monitoring** → **Metrics Explorer**
2. Search for `logging.googleapis.com/user/http_request_duration`
3. Select the metric
4. Configure visualization:
   - **Group by**: `route`, `method`
   - **Aggregation**:
     - Mean: Select "mean"
     - P50: Select "50th percentile"
     - P75: Select "75th percentile"
     - P90: Select "90th percentile"
     - P99: Select "99th percentile"
   - **Aligner**: Sum or Delta (depending on your needs)

### 4. Create Dashboard

Create a custom dashboard to view all percentiles at once:

```bash
# Save this as dashboard.json
{
  "displayName": "API Performance Metrics",
  "mosaicLayout": {
    "columns": 12,
    "tiles": [
      {
        "width": 6,
        "height": 4,
        "widget": {
          "title": "Request Duration by Route (p50, p90, p99)",
          "xyChart": {
            "dataSets": [
              {
                "timeSeriesQuery": {
                  "timeSeriesFilter": {
                    "filter": "resource.type=\"cloud_run_revision\" metric.type=\"logging.googleapis.com/user/http_request_duration\"",
                    "aggregation": {
                      "alignmentPeriod": "60s",
                      "perSeriesAligner": "ALIGN_DELTA",
                      "crossSeriesReducer": "REDUCE_PERCENTILE_50",
                      "groupByFields": ["metric.label.route"]
                    }
                  }
                },
                "plotType": "LINE",
                "targetAxis": "Y1"
              }
            ]
          }
        }
      }
    ]
  }
}

# Import dashboard
gcloud monitoring dashboards create --config-from-file=dashboard.json
```

Or create manually in Cloud Console:
1. **Cloud Monitoring** → **Dashboards** → **Create Dashboard**
2. Add charts for each percentile
3. Group by route/method

## Querying Metrics

### Using MQL (Monitoring Query Language)

```sql
# P99 latency by route
fetch cloud_run_revision
| metric 'logging.googleapis.com/user/http_request_duration'
| group_by 1m, [value_http_request_duration_percentile: percentile(value.http_request_duration, 99)]
| every 1m
| group_by [resource.service_name, metric.route],
    [value_http_request_duration_percentile_aggregate: aggregate(value_http_request_duration_percentile)]

# Mean latency by route
fetch cloud_run_revision
| metric 'logging.googleapis.com/user/http_request_duration'
| group_by 1m, [value_http_request_duration_mean: mean(value.http_request_duration)]
| every 1m
| group_by [resource.service_name, metric.route],
    [value_http_request_duration_mean_aggregate: aggregate(value_http_request_duration_mean)]
```

### Using Logs Explorer (for debugging)

```
jsonPayload.metric_type="http_request"
jsonPayload.route="/api/games/:id"
```

Then use **Log fields explorer** to see distribution of `duration_ms`.

## Setting Up Alerts

Create alerts for high latency:

```bash
gcloud alpha monitoring policies create \
  --notification-channels=CHANNEL_ID \
  --display-name="High P99 Latency" \
  --condition-display-name="P99 > 1000ms" \
  --condition-expression='
    metric.type="logging.googleapis.com/user/http_request_duration" AND
    resource.type="cloud_run_revision"
  ' \
  --condition-threshold-value=1000 \
  --condition-threshold-duration=60s \
  --condition-comparison=COMPARISON_GT \
  --condition-aggregation-alignment-period=60s \
  --condition-aggregation-per-series-aligner=ALIGN_DELTA \
  --condition-aggregation-cross-series-reducer=REDUCE_PERCENTILE_99 \
  --condition-aggregation-group-by-fields=metric.route
```

Or create in Cloud Console:
1. **Cloud Monitoring** → **Alerting** → **Create Policy**
2. Select `http_request_duration` metric
3. Set condition: P99 > threshold
4. Add notification channels

## Testing Locally

Run the backend and make requests:

```bash
# Start backend
make docker-up

# Generate traffic
for i in {1..100}; do
  curl -H "X-API-Key: $API_KEY" http://localhost:8080/api/matches
  sleep 0.1
done
```

Check logs for structured output:

```bash
docker logs ligain-api-1 2>&1 | grep "request completed" | jq
```

You should see JSON output like:

```json
{
  "level": "info",
  "msg": "request completed",
  "route": "/api/matches",
  "method": "GET",
  "status": 200,
  "duration_ms": 45.2,
  "metric_type": "http_request",
  "time": "2025-10-24T10:30:00Z"
}
```

## Cost Considerations

- **Log ingestion**: ~$0.50 per GB
- **Log storage**: First 50 GB/month free
- **Metrics**: First 150 MB/month free

For typical API usage, expect:
- ~500 bytes per request log
- 1M requests/month = ~500 MB logs = ~$0.25/month

## Alternative: OpenTelemetry (Advanced)

If you need more advanced metrics or distributed tracing, consider OpenTelemetry:

```go
// See: https://cloud.google.com/trace/docs/setup/go-ot
// Requires additional setup but provides richer metrics
```

## Troubleshooting

### Metrics not appearing?

1. **Check logs are being written**:
   ```bash
   gcloud logging read 'jsonPayload.metric_type="http_request"' --limit 10 --format json
   ```

2. **Verify log-based metric exists**:
   ```bash
   gcloud logging metrics list
   ```

3. **Check metric has data**:
   - Go to Metrics Explorer
   - Search for `logging.googleapis.com/user/http_request_duration`
   - If no data, wait 2-3 minutes after first request

### High cardinality warning?

If you have many unique routes, consider:
- Grouping similar routes (e.g., `/games/:id` instead of `/games/123`)
- The middleware already does this using `c.FullPath()`

### Want per-region metrics?

Add region label to middleware:

```go
"region": os.Getenv("CLOUD_RUN_REGION"),
```

## Best Practices

1. **Monitor regularly**: Check p99 latency weekly
2. **Set alerts**: Alert on p99 > acceptable threshold
3. **Optimize routes**: Focus on high-traffic, high-latency routes
4. **Database queries**: Most latency comes from DB - monitor those separately
5. **Cold starts**: Cloud Run cold starts affect p99 - consider min instances

## References

- [Cloud Run Logging](https://cloud.google.com/run/docs/logging)
- [Log-based Metrics](https://cloud.google.com/logging/docs/logs-based-metrics)
- [Cloud Monitoring Distribution Metrics](https://cloud.google.com/monitoring/api/v3/distribution-metrics)
