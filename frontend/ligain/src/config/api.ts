import Constants from 'expo-constants';

// Get API configuration from environment variables
export const API_CONFIG = {
  BASE_URL: Constants.expoConfig?.extra?.apiBaseUrl || 'https://server-dev-4c7b2bc-uyqlakruuq-ew.a.run.app',
  API_KEY: Constants.expoConfig?.extra?.apiKey,
  GAME_ID: '123e4567-e89b-12d3-a456-426614174000',
} as const;

// Helper function to get API headers
export const getApiHeaders = (additionalHeaders?: Record<string, string>) => {
  if (!API_CONFIG.API_KEY) {
    throw new Error('API_KEY is not configured. Please set it in your environment variables.');
  }
  
  return {
    'X-API-Key': API_CONFIG.API_KEY,
    ...additionalHeaders,
  };
}; 