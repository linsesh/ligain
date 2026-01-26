import Constants from 'expo-constants';
import { getItem } from '../utils/storage';

export const API_CONFIG = {
  BASE_URL: Constants.expoConfig?.extra?.apiBaseUrl,
  API_KEY: Constants.expoConfig?.extra?.apiKey,
  APPLE_CLIENT_ID: Constants.expoConfig?.extra?.appleClientId,
  APP_VERSION: Constants.expoConfig?.version || '0.0.0',
} as const;

// Default App Store URL for forced updates
const DEFAULT_STORE_URL = 'https://apps.apple.com/fr/app/ligain/id6748531523';

// Callback for handling version outdated responses
let versionOutdatedCallback: ((storeUrl: string, minVersion: string) => void) | null = null;

/**
 * Register a callback to be called when the app version is outdated (HTTP 426)
 */
export const setVersionOutdatedCallback = (callback: (storeUrl: string, minVersion: string) => void) => {
  versionOutdatedCallback = callback;
};

export const getApiHeaders = (additionalHeaders?: Record<string, string>) => {
  console.log('🔧 API - API_KEY configured:', !!API_CONFIG.API_KEY);
  console.log('🔧 API - BASE_URL:', API_CONFIG.BASE_URL);

  if (!API_CONFIG.API_KEY) {
    console.error('❌ API - API_KEY is not configured!');
    throw new Error('API_KEY is not configured. Please set it in your environment variables.');
  }
  return {
    'X-API-Key': API_CONFIG.API_KEY,
    'X-App-Version': API_CONFIG.APP_VERSION,
    ...additionalHeaders,
  };
};

export const getAuthenticatedHeaders = async (additionalHeaders?: Record<string, string>): Promise<Record<string, string>> => {
  console.log('🔧 API - Getting authenticated headers');
  
  const baseHeaders = getApiHeaders(additionalHeaders);
  
  try {
    const token = await getItem('auth_token');
    console.log('🔧 API - Auth token exists:', !!token);
    console.log('🔧 API - Token value:', token ? `${token.substring(0, 10)}...` : 'null');
    
    if (token) {
      return {
        ...baseHeaders,
        'Authorization': `Bearer ${token}`,
      };
    } else {
      console.warn('⚠️ API - No auth token found, returning unauthenticated headers');
      return baseHeaders;
    }
  } catch (error) {
    console.error('❌ API - Error getting auth token:', error);
    return baseHeaders;
  }
};

// Enhanced fetch function that handles automatic token refresh and version checks
export const authenticatedFetch = async (url: string, options: RequestInit = {}): Promise<Response> => {
  console.log('🔧 API - Making authenticated request to:', url);

  // Get headers with current token
  const headers = await getAuthenticatedHeaders(options.headers as Record<string, string>);

  // Make the request
  const response = await fetch(url, {
    ...options,
    headers,
  });

  // Check for version outdated response (HTTP 426)
  if (response.status === 426) {
    console.log('⚠️ API - App version outdated, update required');
    try {
      const errorData = await response.clone().json();
      if (errorData.code === 'VERSION_OUTDATED' && versionOutdatedCallback) {
        const storeUrl = errorData.store_url || DEFAULT_STORE_URL;
        const minVersion = errorData.min_version || 'unknown';
        versionOutdatedCallback(storeUrl, minVersion);
      }
    } catch (parseError) {
      // If we can't parse the response, still trigger the callback with defaults
      console.error('❌ API - Failed to parse 426 response:', parseError);
      if (versionOutdatedCallback) {
        versionOutdatedCallback(DEFAULT_STORE_URL, 'unknown');
      }
    }
    return response;
  }

  // Check if we got a new token in the response headers
  const newToken = response.headers.get('X-New-Token');
  const tokenRefreshed = response.headers.get('X-Token-Refreshed');

  if (newToken && tokenRefreshed === 'true') {
    console.log('🔄 API - Token was refreshed automatically, updating stored token');
    const { setItem } = await import('../utils/storage');
    await setItem('auth_token', newToken);
  }

  return response;
}; 