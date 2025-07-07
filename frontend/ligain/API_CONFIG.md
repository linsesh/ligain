# API Configuration

## Environment Variables

**IMPORTANT**: Never commit API keys to version control! Create a `.env` file in the frontend directory:

```bash
# API Configuration
API_BASE_URL=https://server-dev-4c7b2bc-uyqlakruuq-ew.a.run.app
API_KEY=your_api_key_here

# Environment
NODE_ENV=development
```

## Setup Instructions

1. **Create `.env` file** in `frontend/ligain/` directory
2. **Add `.env` to `.gitignore`** to prevent committing secrets
3. **Set your API key** in the `.env` file
4. **Restart your development server** after creating the `.env` file

## Security Best Practices

### For Development:
- ✅ Use environment variables in `.env` file
- ✅ Add `.env` to `.gitignore` to prevent committing secrets
- ✅ Use different API keys for different environments
- ❌ Never hardcode API keys in source code
- ❌ Never commit `.env` files to version control

### For Production:
- Use secure key management (AWS Secrets Manager, Google Secret Manager, etc.)
- Never commit API keys to version control
- Use different API keys for different environments (dev/staging/prod)

## Current Configuration

The app uses `expo-constants` to access environment variables through `app.config.ts`. The configuration is centralized in `src/config/api.ts`.

## API Key Management

The API key is used to authenticate requests to your Cloud Run backend. The key is sent in the `X-API-Key` header with every request.

## Error Handling

If the API key is not configured, the app will throw an error with a helpful message: "API_KEY is not configured. Please set it in your environment variables." 