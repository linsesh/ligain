import { GoogleSignin, statusCodes } from '@react-native-google-signin/google-signin';
import { Platform } from 'react-native';
import * as AppleAuthentication from 'expo-apple-authentication';
import { API_CONFIG, getApiHeaders } from '../config/api';

// Configure Google Sign-In (only once, here)
const googleSignInConfig: Record<string, any> = {
  offlineAccess: true,
  webClientId: process.env.EXPO_PUBLIC_WEB_GOOGLE_CLIENT_ID,
};

if (Platform.OS === 'ios') {
  googleSignInConfig.iosClientId = process.env.EXPO_PUBLIC_IOS_GOOGLE_CLIENT_ID;
}

GoogleSignin.configure(googleSignInConfig);

export interface AuthResult {
  provider: 'google' | 'apple' | 'guest';
  token: string;
  email: string;
  name: string;
  playerId?: string; // Optional for guest users
}

export class AuthService {
  static async signInWithGoogle(): Promise<AuthResult> {
    console.log('🔐 Google Sign-In - Starting sign in process');
    
    // Check if Google Sign-In is properly configured
    const webClientId = process.env.EXPO_PUBLIC_WEB_GOOGLE_CLIENT_ID;
    const iosClientId = process.env.EXPO_PUBLIC_IOS_GOOGLE_CLIENT_ID;
    
    console.log('🔐 Google Sign-In - Configuration check:', {
      webClientId: webClientId ? 'configured' : 'NOT_CONFIGURED',
      iosClientId: iosClientId ? 'configured' : 'NOT_CONFIGURED',
    });
    
    if (!webClientId) {
      throw new Error('Google Sign-In not properly configured: missing EXPO_PUBLIC_WEB_GOOGLE_CLIENT_ID');
    }
    if (Platform.OS === 'ios' && !iosClientId) {
      throw new Error('Google Sign-In not properly configured: missing EXPO_PUBLIC_IOS_GOOGLE_CLIENT_ID on iOS');
    }
    
    try {
      // Check if your device supports Google Play
      if (Platform.OS === 'android') {
        await GoogleSignin.hasPlayServices({ showPlayServicesUpdateDialog: true });
      }
      
            // Sign in - userInfo is the User object directly
      let userInfo;
      try {
        userInfo = await GoogleSignin.signIn();
      } catch (signInError: any) {
        console.log('🔐 Google Sign-In - SignIn error caught:', {
          code: signInError.code,
          message: signInError.message,
          statusCodes: statusCodes,
          SIGN_IN_CANCELLED: statusCodes.SIGN_IN_CANCELLED,
          isCancelled: signInError.code === statusCodes.SIGN_IN_CANCELLED
        });
        
        // Handle sign-in cancellation immediately
        if (signInError.code === statusCodes.SIGN_IN_CANCELLED) {
          console.log('🔐 Google Sign-In - Cancellation detected, throwing cancellation error');
          throw new Error('Sign-in was cancelled');
        }
        // Re-throw other sign-in errors
        console.log('🔐 Google Sign-In - Re-throwing non-cancellation error');
        throw signInError;
      }
      
      console.log('🔐 Google Sign-In - Sign in successful:', userInfo);
      console.log('🔐 Google Sign-In - User info structure:', {
        hasUser: !!userInfo,
        userKeys: userInfo ? Object.keys(userInfo as any) : 'no user object',
        email: (userInfo as any)?.email,
        name: (userInfo as any)?.name,
        givenName: (userInfo as any)?.givenName,
        familyName: (userInfo as any)?.familyName,
        type: (userInfo as any)?.type,
        data: (userInfo as any)?.data,
      });
      
      // Check if this is a cancellation result (not an error)
      if (userInfo && (userInfo as any).type === 'cancelled') {
        console.log('🔐 Google Sign-In - Cancellation result detected');
        throw new Error('Sign-in was cancelled');
      }
      
      // Get tokens separately
      const tokens = await GoogleSignin.getTokens();
      console.log('🔐 Google Sign-In - Tokens retrieved:', {
        accessToken: tokens.accessToken ? '***token***' : 'NO_TOKEN',
        idToken: tokens.idToken ? '***token***' : 'NO_TOKEN'
      });

      // Extract email and name from userInfo (userInfo is the User object directly)
      let email = (userInfo as any)?.email || '';
      let name = (userInfo as any)?.name || 
                 ((userInfo as any)?.givenName && (userInfo as any)?.familyName 
                   ? `${(userInfo as any).givenName} ${(userInfo as any).familyName}` 
                   : (userInfo as any)?.givenName || (userInfo as any)?.familyName || '');

      // If email or name is still empty, try to extract it from the ID token
      if ((!email || !name) && tokens.idToken) {
        try {
          // Decode the JWT token to get user info
          const tokenParts = tokens.idToken.split('.');
          if (tokenParts.length === 3) {
            const payload = tokenParts[1];
            // Add padding if needed
            const paddedPayload = payload + '='.repeat((4 - payload.length % 4) % 4);
            const decodedPayload = atob(paddedPayload.replace(/-/g, '+').replace(/_/g, '/'));
            const claims = JSON.parse(decodedPayload);
            
            console.log('🔐 Google Sign-In - JWT claims:', {
              email: claims.email,
              name: claims.name,
              given_name: claims.given_name,
              family_name: claims.family_name,
            });
            
            // Use JWT claims as fallback
            if (!email && claims.email) {
              email = claims.email;
            }
            if (!name && claims.name) {
              name = claims.name;
            } else if (!name && claims.given_name && claims.family_name) {
              name = `${claims.given_name} ${claims.family_name}`;
            } else if (!name && claims.given_name) {
              name = claims.given_name;
            } else if (!name && claims.family_name) {
              name = claims.family_name;
            }
          }
        } catch (decodeError) {
          console.warn('🔐 Google Sign-In - Failed to decode JWT token:', decodeError);
        }
      }

      console.log('🔐 Google Sign-In - Final extracted user data:', { email, name });

      if (!email) {
        throw new Error('Email not provided by Google Sign-In');
      }

      if (!name) {
        throw new Error('Name not provided by Google Sign-In');
      }

      return {
        provider: 'google',
        token: tokens.idToken!, // Use ID token for backend verification
        email: email,
        name: name,
      };
    } catch (error: any) {
      console.error('🔐 Google Sign-In - Error during sign in:', error);
      
      // SIGN_IN_CANCELLED is already handled above
      if (error.code === statusCodes.IN_PROGRESS) {
        throw new Error('Sign-in is already in progress');
      } else if (error.code === statusCodes.PLAY_SERVICES_NOT_AVAILABLE) {
        throw new Error('Google Play Services not available');
      } else if (error.code === statusCodes.SIGN_IN_REQUIRED) {
        throw new Error('Sign-in required');
      } else {
        throw new Error(`Sign-in failed: ${error.message}`);
      }
    }
  }

  static async signInWithApple(): Promise<AuthResult> {
    if (Platform.OS !== 'ios') {
      throw new Error('Apple sign-in is only supported on iOS');
    }

    console.log('🔐 Apple Sign-In - Starting sign in process');

    try {
      // Check if Apple Authentication is available
      const isAvailable = await AppleAuthentication.isAvailableAsync();
      if (!isAvailable) {
        throw new Error('Apple Sign-In is not available on this device');
      }

      // Perform Apple Sign-In
      const credential = await AppleAuthentication.signInAsync({
        requestedScopes: [
          AppleAuthentication.AppleAuthenticationScope.FULL_NAME,
          AppleAuthentication.AppleAuthenticationScope.EMAIL,
        ],
      });

      console.log('🔐 Apple Sign-In - Sign in successful:', {
        user: credential.user,
        email: credential.email,
        fullName: credential.fullName,
        identityToken: credential.identityToken ? '***token***' : 'NO_TOKEN',
        authorizationCode: credential.authorizationCode ? '***code***' : 'NO_CODE'
      });

      // Extract user information
      const email = credential.email || '';
      const fullName = credential.fullName;
      const name = fullName ? `${fullName.givenName || ''} ${fullName.familyName || ''}`.trim() : '';
      
      // Use identity token as the authentication token
      const token = credential.identityToken || '';
      
      if (!token) {
        throw new Error('Failed to get identity token from Apple');
      }

      return {
        provider: 'apple',
        token,
        email,
        name,
      };
    } catch (error: any) {
      console.error('🔐 Apple Sign-In - Error:', error);
      
      if (error.code === 'ERR_CANCELED') {
        throw new Error('Sign-in was canceled by the user');
      } else if (error.code === 'ERR_INVALID_RESPONSE') {
        throw new Error('Invalid response from Apple');
      } else if (error.code === 'ERR_NOT_AVAILABLE') {
        throw new Error('Apple Sign-In is not available on this device');
      } else if (error.code === 'ERR_REQUEST_EXPIRED') {
        throw new Error('Apple Sign-In request expired');
      } else if (error.code === 'ERR_REQUEST_NOT_HANDLED') {
        throw new Error('Apple Sign-In request not handled');
      } else {
        throw new Error(`Apple Sign-In failed: ${error.message}`);
      }
    }
  }

  // Get available sign-in options for the current platform
  static getAvailableProviders(): ('google' | 'apple' | 'guest')[] {
    if (Platform.OS === 'ios') {
      return ['google', 'apple', 'guest'];
    } else if (Platform.OS === 'android') {
      return ['google', 'guest'];
    } else {
      return ['google', 'guest']; // web
    }
  }

  // Mock authentication for development/testing
  static async mockSignIn(provider: 'google' | 'apple' | 'guest'): Promise<AuthResult> {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve({
          provider,
          token: `mock_${provider}_token_${Date.now()}`,
          email: provider === 'guest' ? '' : `mock_${provider}_user@example.com`,
          name: provider === 'guest' ? 'Player 1' : `Mock ${provider.charAt(0).toUpperCase() + provider.slice(1)} User`,
        });
      }, 1000);
    });
  }

  // Guest authentication
  static async signInAsGuest(displayName: string): Promise<AuthResult> {
    console.log('🔐 Guest Sign-In - Starting guest sign in process');
    
    try {
      const response = await fetch(`${API_CONFIG.BASE_URL}/api/auth/signin/guest`, {
        method: 'POST',
        headers: {
          ...getApiHeaders(),
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: displayName,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || `Guest authentication failed with status ${response.status}`);
      }

      const data = await response.json();
      console.log('🔐 Guest Sign-In - Success:', data);
      
      return {
        provider: 'guest',
        token: data.token,
        email: '', // Guest users don't have email
        name: data.player.name,
        playerId: data.player.id, // Include the player ID from backend
      };
    } catch (error) {
      console.error('🔐 Guest Sign-In - Error:', error);
      
      // Handle network errors (server unreachable, etc.)
      if (error instanceof TypeError && error.message.includes('fetch')) {
        throw new Error('Ligain servers are not available for now. Please try again later.');
      }
      
      // Re-throw other errors
      throw error;
    }
  }

  // Test provider availability in development
  static testProviderAvailability(): void {
    console.log('🔐 AuthService - Testing provider availability');
    console.log('🔐 AuthService - Available providers:', this.getAvailableProviders());
    console.log('🔐 AuthService - Platform:', Platform.OS);
    console.log('🔐 AuthService - Google Sign-In configured:', {
      webClientId: process.env.EXPO_PUBLIC_WEB_GOOGLE_CLIENT_ID ? 'configured' : 'NOT_CONFIGURED',
      iosClientId: process.env.EXPO_PUBLIC_IOS_GOOGLE_CLIENT_ID ? 'configured' : 'NOT_CONFIGURED',
    });
  }
} 