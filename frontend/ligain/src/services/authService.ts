import * as AuthSession from 'expo-auth-session';
import * as WebBrowser from 'expo-web-browser';
import { Platform } from 'react-native';

// Complete the auth session
WebBrowser.maybeCompleteAuthSession();

// Google OAuth configuration
const GOOGLE_CLIENT_ID = '628283030166-t524lojsr97phq29002qm062kdbok7ec.apps.googleusercontent.com';
const GOOGLE_REDIRECT_URI = AuthSession.makeRedirectUri({
  scheme: 'ligain',
  path: 'auth',
});

// Apple OAuth configuration (for iOS only)
const APPLE_CLIENT_ID = 'com.ligain.app';
const APPLE_REDIRECT_URI = AuthSession.makeRedirectUri({
  scheme: 'ligain',
  path: 'auth',
});

export interface AuthResult {
  provider: 'google' | 'apple';
  token: string;
  email: string;
  name: string;
}

export class AuthService {
  static async signInWithGoogle(): Promise<AuthResult> {
    const request = new AuthSession.AuthRequest({
      clientId: GOOGLE_CLIENT_ID,
      scopes: ['openid', 'profile', 'email'],
      redirectUri: GOOGLE_REDIRECT_URI,
      responseType: AuthSession.ResponseType.Code,
      extraParams: {
        access_type: 'offline',
      },
    });

    const result = await request.promptAsync({
      authorizationEndpoint: 'https://accounts.google.com/oauth/authorize',
    });

    if (result.type === 'success') {
      // Exchange code for token
      const tokenResponse = await AuthSession.exchangeCodeAsync(
        {
          clientId: GOOGLE_CLIENT_ID,
          code: result.params.code,
          redirectUri: GOOGLE_REDIRECT_URI,
          extraParams: {
            code_verifier: request.codeVerifier!,
          },
        },
        {
          tokenEndpoint: 'https://oauth2.googleapis.com/token',
        }
      );

      // Get user info
      const userInfoResponse = await fetch(
        `https://www.googleapis.com/oauth2/v2/userinfo?access_token=${tokenResponse.accessToken}`
      );
      const userInfo = await userInfoResponse.json();

      return {
        provider: 'google',
        token: tokenResponse.accessToken,
        email: userInfo.email,
        name: userInfo.name,
      };
    } else {
      throw new Error('Google sign-in was cancelled or failed');
    }
  }

  static async signInWithApple(): Promise<AuthResult> {
    if (Platform.OS !== 'ios') {
      throw new Error('Apple sign-in is only supported on iOS');
    }

    const request = new AuthSession.AuthRequest({
      clientId: APPLE_CLIENT_ID,
      scopes: ['name', 'email'],
      redirectUri: APPLE_REDIRECT_URI,
      responseType: AuthSession.ResponseType.Code,
      extraParams: {
        response_mode: 'form_post',
      },
    });

    const result = await request.promptAsync({
      authorizationEndpoint: 'https://appleid.apple.com/auth/authorize',
    });

    if (result.type === 'success') {
      // Exchange code for token
      const tokenResponse = await AuthSession.exchangeCodeAsync(
        {
          clientId: APPLE_CLIENT_ID,
          code: result.params.code,
          redirectUri: APPLE_REDIRECT_URI,
          extraParams: {
            code_verifier: request.codeVerifier!,
          },
        },
        {
          tokenEndpoint: 'https://appleid.apple.com/auth/token',
        }
      );

      // Parse user info from the response
      const userInfo = result.params.user ? JSON.parse(result.params.user) : {};

      return {
        provider: 'apple',
        token: tokenResponse.accessToken,
        email: userInfo.email || '',
        name: userInfo.name ? `${userInfo.name.firstName} ${userInfo.name.lastName}` : '',
      };
    } else {
      throw new Error('Apple sign-in was cancelled or failed');
    }
  }

  // Get available sign-in options for the current platform
  static getAvailableProviders(): ('google' | 'apple')[] {
    if (Platform.OS === 'ios') {
      return ['google', 'apple'];
    } else if (Platform.OS === 'android') {
      return ['google'];
    } else {
      return ['google']; // web
    }
  }

  // Mock authentication for development/testing
  static async mockSignIn(provider: 'google' | 'apple'): Promise<AuthResult> {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve({
          provider,
          token: `mock_${provider}_token_${Date.now()}`,
          email: `test@${provider}.com`,
          name: `Test ${provider.charAt(0).toUpperCase() + provider.slice(1)} User`,
        });
      }, 1000);
    });
  }

  // Test function to verify provider availability (for development)
  static testProviderAvailability(): void {
    const providers = this.getAvailableProviders();
    console.log(`Available providers on ${Platform.OS}:`, providers);
    
    if (Platform.OS === 'ios') {
      console.log('✅ iOS: Should show Google and Apple');
      console.log('Expected: ["google", "apple"]');
      console.log('Actual:', providers);
    } else if (Platform.OS === 'android') {
      console.log('✅ Android: Should show Google only');
      console.log('Expected: ["google"]');
      console.log('Actual:', providers);
    } else {
      console.log('✅ Web: Should show Google only');
      console.log('Expected: ["google"]');
      console.log('Actual:', providers);
    }
  }
} 