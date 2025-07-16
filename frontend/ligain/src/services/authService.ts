import { GoogleSignin, statusCodes } from '@react-native-google-signin/google-signin';
import { Platform } from 'react-native';
import { API_CONFIG, getApiHeaders } from '../config/api';

// Configure Google Sign-In (only once, here)
GoogleSignin.configure({
  webClientId: process.env.EXPO_PUBLIC_WEB_GOOGLE_CLIENT_ID, // Required for getting the idToken on iOS
  iosClientId: process.env.EXPO_PUBLIC_IOS_GOOGLE_CLIENT_ID, // Required for iOS
  offlineAccess: true, // if you want to access Google API on behalf of the user
});

export interface AuthResult {
  provider: 'google' | 'apple' | 'guest';
  token: string;
  email: string;
  name: string;
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
    
    if (!webClientId || !iosClientId) {
      throw new Error('Google Sign-In not properly configured. Please check your environment variables.');
    }
    
    try {
      // Check if your device supports Google Play
      await GoogleSignin.hasPlayServices();
      
      // Sign in - userInfo is the User object directly
      const userInfo = await GoogleSignin.signIn();
      console.log('🔐 Google Sign-In - Sign in successful:', userInfo);
      console.log('🔐 Google Sign-In - User info structure:', {
        hasUser: !!userInfo,
        userKeys: userInfo ? Object.keys(userInfo as any) : 'no user object',
        email: (userInfo as any)?.email,
        name: (userInfo as any)?.name,
        givenName: (userInfo as any)?.givenName,
        familyName: (userInfo as any)?.familyName,
      });
      
      // Log the entire userInfo object to see its structure
      console.log('🔐 Google Sign-In - Full userInfo object:', JSON.stringify(userInfo, null, 2));
      
      // Simple debug: log each property individually
      console.log('🔐 Google Sign-In - Debug individual properties:');
      console.log('  - userInfo type:', typeof userInfo);
      console.log('  - userInfo keys:', userInfo ? Object.keys(userInfo as any) : 'null');
      console.log('  - userInfo.email:', (userInfo as any)?.email);
      console.log('  - userInfo.name:', (userInfo as any)?.name);
      console.log('  - userInfo.givenName:', (userInfo as any)?.givenName);
      console.log('  - userInfo.familyName:', (userInfo as any)?.familyName);

      // Get tokens separately
      const tokens = await GoogleSignin.getTokens();
      console.log('🔐 Google Sign-In - Tokens retrieved:', {
        accessToken: tokens.accessToken ? '***token***' : 'NO_TOKEN',
        idToken: tokens.idToken ? '***token***' : 'NO_TOKEN'
      });

      // Extract email and name from userInfo (userInfo is the User object directly)
      let email = (userInfo as any)?.data?.user?.email || '';
      let name = (userInfo as any)?.data?.user?.name || 
                 ((userInfo as any)?.data?.user?.givenName && (userInfo as any)?.data?.user?.familyName 
                   ? `${(userInfo as any).data.user.givenName} ${(userInfo as any).data.user.familyName}` 
                   : (userInfo as any)?.data?.user?.givenName || (userInfo as any)?.data?.user?.familyName || '');

      // If email is still empty, try to extract it from the ID token
      if (!email && tokens.idToken) {
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

      return {
        provider: 'google',
        token: tokens.idToken!, // Use ID token for backend verification
        email: email,
        name: name,
      };
    } catch (error: any) {
      console.error('🔐 Google Sign-In - Error during sign in:', error);
      
      if (error.code === statusCodes.SIGN_IN_CANCELLED) {
        throw new Error('Sign-in was cancelled');
      } else if (error.code === statusCodes.IN_PROGRESS) {
        throw new Error('Sign-in is already in progress');
      } else if (error.code === statusCodes.PLAY_SERVICES_NOT_AVAILABLE) {
        throw new Error('Google Play Services not available');
      } else {
        throw new Error(`Sign-in failed: ${error.message}`);
      }
    }
  }

  static async signInWithApple(): Promise<AuthResult> {
    if (Platform.OS !== 'ios') {
      throw new Error('Apple sign-in is only supported on iOS');
    }

    // For now, return a mock result for Apple
    // TODO: Implement real Apple Sign-In
    throw new Error('Apple Sign-In not yet implemented');
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
      };
    } catch (error) {
      console.error('🔐 Guest Sign-In - Error:', error);
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