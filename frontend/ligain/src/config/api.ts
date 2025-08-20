import Constants from 'expo-constants';
import { getItem } from '../utils/storage';

export const API_CONFIG = {
  BASE_URL: Constants.expoConfig?.extra?.apiBaseUrl,
  API_KEY: Constants.expoConfig?.extra?.apiKey,
  APPLE_CLIENT_ID: Constants.expoConfig?.extra?.appleClientId,
} as const;

export const getApiHeaders = (additionalHeaders?: Record<string, string>) => {
  console.log('🔧 API - API_KEY configured:', !!API_CONFIG.API_KEY);
  console.log('🔧 API - BASE_URL:', API_CONFIG.BASE_URL);
  
  if (!API_CONFIG.API_KEY) {
    console.error('❌ API - API_KEY is not configured!');
    throw new Error('API_KEY is not configured. Please set it in your environment variables.');
  }
  return {
    'X-API-Key': API_CONFIG.API_KEY,
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

// Enhanced fetch function that handles automatic token refresh
export const authenticatedFetch = async (url: string, options: RequestInit = {}): Promise<Response> => {
  console.log('🔧 API - Making authenticated request to:', url);
  
  // Get headers with current token
  const headers = await getAuthenticatedHeaders(options.headers as Record<string, string>);
  
  // Make the request
  const response = await fetch(url, {
    ...options,
    headers,
  });
  
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