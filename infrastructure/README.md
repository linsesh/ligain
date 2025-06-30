# Ligain Infrastructure

This directory contains the Pulumi infrastructure code for deploying the Ligain backend to Google Cloud Run.

## Security and Configuration Files

### Important: Sensitive Configuration Files

The following files contain sensitive information and are **NOT** committed to version control:
- `Pulumi.dev.yaml` - Development environment configuration with real API keys
- `Pulumi.prod.yaml` - Production environment configuration with real API keys

### Template Files

Template files are provided to show the configuration structure:
- `Pulumi.dev.yaml.template` - Template for development configuration
- `Pulumi.prod.yaml.template` - Template for production configuration

### Setting Up Configuration Files

1. **Copy the template files**:
   ```bash
   cp infrastructure/Pulumi.dev.yaml.template infrastructure/Pulumi.dev.yaml
   cp infrastructure/Pulumi.prod.yaml.template infrastructure/Pulumi.prod.yaml
   ```

2. **Edit the files** with your actual values:
   - Replace `your-gcp-project-id` with your actual GCP project ID
   - Replace `your-dev-api-key` with your actual API key
   - Replace `your-sportmonk-api-token` with your actual SportMonk token
   - Update database URLs and allowed origins

## Environment Variables Management

### Overview

Environment variables are managed through Pulumi configuration files, which provides several benefits:
- **Security**: Sensitive values are not hardcoded in the infrastructure code
- **Environment-specific**: Different values for dev and prod environments
- **Version control**: Configuration changes are tracked in git
- **Secrets management**: Can integrate with Google Secret Manager for sensitive data

### Current Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `DATABASE_URL` | PostgreSQL connection string | Yes |
| `API_KEY` | API key for backend authentication (used by frontend/mobile app) | Yes |
| `SPORTSMONK_API_TOKEN` | API token for SportMonk API (external football data service) | Yes |
| `ALLOWED_ORIGINS` | Comma-separated list of allowed CORS origins | Yes |
| `ENV` | Environment name (dev/prod) | Auto-set |
| `PORT` | Server port (defaults to 8080) | Auto-set |

### Configuration Files

- `Pulumi.dev.yaml` - Development environment configuration
- `Pulumi.prod.yaml` - Production environment configuration

### Setting Environment Variables

#### For Development
```bash
cd infrastructure
pulumi config set ligain:database_url "postgresql://user:pass@host:port/db" --stack dev
pulumi config set ligain:api_key "your-dev-api-key" --stack dev
pulumi config set ligain:sportsmonk_api_token "your-dev-sportsmonk-api-token" --stack dev
pulumi config set ligain:allowed_origins "http://localhost:3000,https://your-dev-domain.com" --stack dev
```

#### For Production
```bash
cd infrastructure
pulumi config set ligain:database_url "postgresql://user:pass@host:port/db" --stack prod
pulumi config set ligain:api_key "your-prod-api-key" --stack prod
pulumi config set ligain:sportsmonk_api_token "your-prod-sportsmonk-api-token" --stack prod
pulumi config set ligain:allowed_origins "https://your-prod-domain.com" --stack prod
```

### Viewing Current Configuration

```bash
# View all configuration for current stack
pulumi config

# View specific configuration
pulumi config get ligain:database_url
```

### Security Best Practices

1. **Never commit sensitive values** to version control
2. **Use Google Secret Manager** for production secrets:
   ```bash
   # Create a secret
   gcloud secrets create ligain-database-url --data-file=-
   
   # Reference in Pulumi config
   pulumi config set ligain:database_url "$(gcloud secrets versions access latest --secret=ligain-database-url)" --stack prod
   ```

3. **Rotate API keys regularly**
4. **Use different values for dev and prod environments**

### Alternative Approaches

#### 1. Google Secret Manager (Recommended for Production)
For production environments, consider using Google Secret Manager:

```go
// In main.go, add secret manager integration
secretClient, err := secretmanager.NewClient(ctx)
if err != nil {
    return err
}
defer secretClient.Close()

// Access secrets
databaseURL, err := secretClient.AccessSecretVersion(ctx, "projects/your-project/secrets/database-url/versions/latest")
```

#### 2. Environment-Specific Files
You could also use environment-specific `.env` files, but this is less secure and harder to manage in containerized environments.

#### 3. Cloud Run Environment Variables UI
You can set environment variables directly in the Google Cloud Console, but this makes infrastructure changes harder to track and version control.

### Deployment

After updating environment variables:

```bash
# Deploy to dev
pulumi up --stack dev

# Deploy to prod
pulumi up --stack prod
```

### Troubleshooting

1. **Check environment variables in Cloud Run**:
   ```bash
   gcloud run services describe server-dev --region=europe-west1 --format="value(spec.template.spec.containers[0].env[].name,spec.template.spec.containers[0].env[].value)"
   ```

2. **View logs**:
   ```bash
   gcloud logs read "resource.type=cloud_run_revision AND resource.labels.service_name=server-dev" --limit=50
   ```

3. **Test locally** with environment variables:
   ```bash
   export DATABASE_URL="your-db-url"
   export API_KEY="your-api-key"
   export SPORTSMONK_API_TOKEN="your-sportsmonk-api-token"
   export ALLOWED_ORIGINS="http://localhost:3000"
   go run backend/main.go
   ``` 