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
  signIn: (provider: 'google' | 'apple', token: string, email: string, name: string) => Promise<void>;
  signOut: () => Promise<void>;
  checkAuth: () => Promise<void>;
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

  const signIn = async (provider: 'google' | 'apple', token: string, email: string, name: string) => {
    try {
      setIsLoading(true);

      const response = await fetch(`${API_CONFIG.BASE_URL}/api/auth/signin`, {
        method: 'POST',
        headers: {
          ...getApiHeaders(),
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          provider,
          token,
          email,
          name,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Authentication failed');
      }

      const data = await response.json();
      
      // Store token and player data
      await setItem(AUTH_TOKEN_KEY, data.token);
      await setItem(PLAYER_DATA_KEY, JSON.stringify(data.player));
      
      setPlayer(data.player);
    } catch (error) {
      console.error('Sign in error:', error);
      throw error;
    } finally {
      setIsLoading(false);
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
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}; 