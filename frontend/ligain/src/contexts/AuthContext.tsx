import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { Platform } from 'react-native';
import { API_CONFIG, getApiHeaders } from '../config/api';
import { getItem, setItem, multiRemove, isUsingMemoryFallback } from '../utils/storage';

export interface Player {
  id: string;
  name: string;
  email?: string;
  provider?: string;
  provider_id?: string;
  created_at?: string;
  updated_at?: string;
}

interface AuthContextType {
  player: Player | null;
  isLoading: boolean;
  signIn: (provider: 'google' | 'apple' | 'guest', token: string, email: string, name: string) => Promise<void>;
  signOut: () => Promise<void>;
  checkAuth: () => Promise<void>;
  setPlayer: (player: Player | null) => void;
  // Modal state management
  showNameModal: boolean;
  setShowNameModal: (show: boolean) => void;
  authResult: any;
  setAuthResult: (result: any) => void;
  selectedProvider: 'google' | 'apple' | 'guest' | null;
  setSelectedProvider: (provider: 'google' | 'apple' | 'guest' | null) => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [player, setPlayer] = useState<Player | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  
  // Modal state management
  const [showNameModal, setShowNameModal] = useState(false);
  const [authResult, setAuthResult] = useState<any>(null);
  const [selectedProvider, setSelectedProvider] = useState<'google' | 'apple' | 'guest' | null>(null);

  const AUTH_TOKEN_KEY = 'auth_token';
  const PLAYER_DATA_KEY = 'player_data';

  useEffect(() => {
    checkAuth();
  }, []);

  const checkAuth = async () => {
    try {
      console.log('🔍 AuthContext - Starting checkAuth');
      setIsLoading(true);
      
      // Check if we're using memory fallback (AsyncStorage not available)
      if (isUsingMemoryFallback()) {
        console.warn('⚠️ Using memory storage fallback - data will be lost on app restart');
      }
      
      console.log('🔍 AuthContext - Storage keys being checked:', { AUTH_TOKEN_KEY, PLAYER_DATA_KEY });
      const token = await getItem(AUTH_TOKEN_KEY);
      const playerData = await getItem(PLAYER_DATA_KEY);
      
      console.log('🔍 AuthContext - Token exists:', !!token, 'PlayerData exists:', !!playerData);

      if (token && playerData) {
        console.log('🔍 AuthContext - Validating token with backend');
        // Validate token with backend
        const response = await fetch(`${API_CONFIG.BASE_URL}/api/auth/me`, {
          headers: {
            ...getApiHeaders(),
            'Authorization': `Bearer ${token}`,
          },
        });

        console.log('🔍 AuthContext - Backend response status:', response.status);

        if (response.ok) {
          const data = await response.json();
          console.log('✅ AuthContext - Token valid, setting player:', data.player?.name);
          setPlayer(data.player);
        } else {
          console.log('❌ AuthContext - Token invalid, clearing storage');
          // Token is invalid, clear storage
          await multiRemove([AUTH_TOKEN_KEY, PLAYER_DATA_KEY]);
          setPlayer(null);
        }
      } else {
        console.log('❌ AuthContext - No token or player data found');
        setPlayer(null);
      }
    } catch (error) {
      console.error('❌ AuthContext - Error checking auth:', error);
      setPlayer(null);
    } finally {
      console.log('🔍 AuthContext - Setting isLoading to false');
      setIsLoading(false);
    }
  };

  const signIn = async (provider: 'google' | 'apple' | 'guest', token: string, email: string, name: string) => {
    console.log('🔐 AuthContext - Starting signIn method');
    console.log('🔐 AuthContext - Parameters:', {
      provider,
      email,
      name,
      token: token ? '***token***' : 'NO_TOKEN'
    });
    
    try {
      setIsLoading(true);
      
      console.log('🔐 AuthContext - API_CONFIG:', {
        BASE_URL: API_CONFIG.BASE_URL,
        API_KEY: API_CONFIG.API_KEY ? 'configured' : 'NOT_CONFIGURED',
        GAME_ID: API_CONFIG.GAME_ID
      });

      // Test API headers configuration
      try {
        const headers = getApiHeaders();
        console.log('🔐 AuthContext - API headers configured successfully:', {
          hasApiKey: !!headers['X-API-Key'],
          hasContentType: 'Content-Type' in headers
        });
      } catch (headerError) {
        console.error('🔐 AuthContext - Failed to get API headers:', headerError);
        throw new Error(`API configuration error: ${headerError instanceof Error ? headerError.message : 'Unknown error'}`);
      }

      let response;
      
      if (provider === 'guest') {
        // Guest authentication uses a different endpoint
        const requestBody = { name };
        
        console.log('🔐 AuthContext - Making guest request to:', `${API_CONFIG.BASE_URL}/api/auth/signin/guest`);
        console.log('🔐 AuthContext - Guest request body:', requestBody);

        response = await fetch(`${API_CONFIG.BASE_URL}/api/auth/signin/guest`, {
          method: 'POST',
          headers: {
            ...getApiHeaders(),
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(requestBody),
        });
      } else {
        // OAuth authentication
        const requestBody = {
          provider,
          token,
          email,
          name,
        };
        
        console.log('🔐 AuthContext - Making OAuth request to:', `${API_CONFIG.BASE_URL}/api/auth/signin`);
        console.log('🔐 AuthContext - OAuth request body:', {
          ...requestBody,
          token: token ? '***token***' : 'NO_TOKEN'
        });

        response = await fetch(`${API_CONFIG.BASE_URL}/api/auth/signin`, {
          method: 'POST',
          headers: {
            ...getApiHeaders(),
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(requestBody),
        });
      }

      console.log('🔐 AuthContext - Response received:', {
        status: response.status,
        statusText: response.statusText,
        ok: response.ok,
        headers: Object.fromEntries(response.headers.entries())
      });

      if (!response.ok) {
        let errorData;
        try {
          errorData = await response.json();
          console.error('🔐 AuthContext - Error response data:', errorData);
        } catch (parseError) {
          console.error('🔐 AuthContext - Failed to parse error response:', parseError);
          errorData = { error: `HTTP ${response.status}: ${response.statusText}` };
        }
        throw new Error(errorData.error || `Authentication failed with status ${response.status}`);
      }

      const data = await response.json();
      console.log('🔐 AuthContext - Success response data:', {
        hasToken: !!data.token,
        player: data.player ? {
          id: data.player.id,
          name: data.player.name,
          email: data.player.email,
          provider: data.player.provider
        } : 'NO_PLAYER_DATA'
      });
      
      // Store token and player data
      console.log('🔐 AuthContext - Storing token and player data');
      console.log('🔐 AuthContext - Token to store:', data.token ? `${data.token.substring(0, 10)}...` : 'null');
      console.log('🔐 AuthContext - Player data to store:', JSON.stringify(data.player));
      await setItem(AUTH_TOKEN_KEY, data.token);
      await setItem(PLAYER_DATA_KEY, JSON.stringify(data.player));
      
      // Verify storage
      const storedToken = await getItem(AUTH_TOKEN_KEY);
      const storedPlayerData = await getItem(PLAYER_DATA_KEY);
      console.log('🔐 AuthContext - Verification - Stored token exists:', !!storedToken);
      console.log('🔐 AuthContext - Verification - Stored player data exists:', !!storedPlayerData);
      
      setPlayer(data.player);
      console.log('🔐 AuthContext - Sign in completed successfully');
    } catch (error) {
      console.error('🔐 AuthContext - Sign in error:', error);
      console.error('🔐 AuthContext - Error details:', {
        name: error instanceof Error ? error.name : 'Unknown',
        message: error instanceof Error ? error.message : 'Unknown error',
        stack: error instanceof Error ? error.stack : 'No stack trace'
      });
      throw error;
    } finally {
      setIsLoading(false);
      console.log('🔐 AuthContext - Sign in method completed');
    }
  };

  const signOut = async () => {
    try {
      const token = await getItem(AUTH_TOKEN_KEY);
      
      if (token) {
        // Call backend to invalidate token
        await fetch(`${API_CONFIG.BASE_URL}/api/auth/signout`, {
          method: 'POST',
          headers: {
            ...getApiHeaders(),
            'Authorization': `Bearer ${token}`,
          },
        });
      }
    } catch (error) {
      console.error('Sign out error:', error);
    } finally {
      // Clear local storage regardless of backend response
      await multiRemove([AUTH_TOKEN_KEY, PLAYER_DATA_KEY]);
      setPlayer(null);
    }
  };

  const value: AuthContextType = {
    player,
    isLoading,
    signIn,
    signOut,
    checkAuth,
    setPlayer,
    // Modal state management
    showNameModal,
    setShowNameModal,
    authResult,
    setAuthResult,
    selectedProvider,
    setSelectedProvider,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}; 