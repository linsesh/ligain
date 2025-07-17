import Constants from 'expo-constants';
import { getItem } from '../utils/storage';

export const API_CONFIG = {
  BASE_URL: Constants.expoConfig?.extra?.apiBaseUrl,
  API_KEY: Constants.expoConfig?.extra?.apiKey,
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