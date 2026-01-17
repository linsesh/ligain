import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { getItem, setItem, multiRemove, isUsingMemoryFallback } from '../utils/storage';
import { useAuthApi, useProfileApi } from '../api';

export interface Player {
  id: string;
  name: string;
  email?: string;
  provider?: string;
  provider_id?: string;
  created_at?: string;
  updated_at?: string;
  avatarUrl?: string | null;
}

interface AuthContextType {
  player: Player | null;
  isLoading: boolean;
  signIn: (provider: 'google' | 'apple' | 'guest', token: string, email: string, name: string) => Promise<{ needDisplayName?: boolean; suggestedName?: string; error?: string } | void>;
  signOut: () => Promise<void>;
  checkAuth: () => Promise<void>;
  setPlayer: (player: Player | null) => void;
  // Avatar methods
  uploadAvatar: (imageUri: string) => Promise<void>;
  deleteAvatar: () => Promise<void>;
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
  const authApi = useAuthApi();
  const profileApi = useProfileApi();
  const [player, setPlayer] = useState<Player | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Modal state management
  const [showNameModal, setShowNameModal] = useState(false);
  const [authResult, setAuthResult] = useState<any>(null);
  const [selectedProvider, setSelectedProvider] = useState<'google' | 'apple' | 'guest' | null>(null);
  const [isMounted, setIsMounted] = useState(true);

  const AUTH_TOKEN_KEY = 'auth_token';
  const PLAYER_DATA_KEY = 'player_data';

  // Track if component is mounted to prevent state updates after unmount
  useEffect(() => {
    setIsMounted(true);
    return () => setIsMounted(false);
  }, []);

  useEffect(() => {
    checkAuth();
  }, []);

  const checkAuth = async () => {
    if (!isMounted) return;

    try {
      console.log('AuthContext - Starting checkAuth');
      setIsLoading(true);

      // Check if we're using memory fallback (AsyncStorage not available)
      if (isUsingMemoryFallback()) {
        console.warn('Using memory storage fallback - data will be lost on app restart');
      }

      // Always delegate to the API implementation
      // AuthContext has no knowledge of mock vs real - just uses the injected API
      // - RealAuthApi.checkAuth() internally checks stored token, returns null if invalid
      // - MockAuthApi.checkAuth() always returns mock player - it's self-contained
      try {
        const result = await authApi.checkAuth();

        if (!isMounted) return;

        if (result) {
          console.log('AuthContext - Auth valid, setting player:', result.player?.name);
          setPlayer(result.player);
        } else {
          console.log('AuthContext - Auth invalid, clearing storage');
          await multiRemove([AUTH_TOKEN_KEY, PLAYER_DATA_KEY]);
          if (!isMounted) return;
          setPlayer(null);
        }
      } catch (fetchError) {
        if (!isMounted) return;
        console.error('AuthContext - Error during auth check:', fetchError);
        setPlayer(null);
      }
    } catch (error) {
      if (!isMounted) return;
      console.error('AuthContext - Error checking auth:', error);
      setPlayer(null);
    } finally {
      if (!isMounted) return;
      console.log('AuthContext - Setting isLoading to false');
      setIsLoading(false);
    }
  };

  const signIn = async (provider: 'google' | 'apple' | 'guest', token: string, email: string, name: string) => {
    console.log('AuthContext - Starting signIn method');
    console.log('AuthContext - Parameters:', { provider, email, name, token: token ? '***token***' : 'NO_TOKEN' });

    try {
      setIsLoading(true);

      let response;

      try {
        if (provider === 'guest') {
          response = await authApi.signInGuest(name);
        } else {
          response = await authApi.signIn(provider, token, email, name);
        }

        console.log('AuthContext - Success response:', {
          hasToken: !!response.token,
          player: response.player ? { id: response.player.id, name: response.player.name } : 'NO_PLAYER_DATA',
          status: response.status
        });

        // Check if backend is requesting display name (two-step flow)
        if (response.status === 'need_display_name') {
          console.log('AuthContext - Backend requesting display name');
          return {
            needDisplayName: true,
            suggestedName: response.suggestedName || '',
            error: response.error
          };
        }

        // Store token and player data
        console.log('AuthContext - Storing token and player data');
        await setItem(AUTH_TOKEN_KEY, response.token);
        await setItem(PLAYER_DATA_KEY, JSON.stringify(response.player));

        setPlayer(response.player);
        console.log('AuthContext - Sign in completed successfully');
      } catch (fetchError) {
        // Handle network errors (server unreachable, etc.)
        if (fetchError instanceof TypeError && fetchError.message.includes('fetch')) {
          console.error('AuthContext - Network error (server unreachable):', fetchError);
          throw new Error('Ligain servers are not available for now. Please try again later.');
        }

        // Re-throw other errors
        throw fetchError;
      }
    } catch (error) {
      console.error('AuthContext - Sign in error:', error);
      throw error;
    } finally {
      setIsLoading(false);
      console.log('AuthContext - Sign in method completed');
    }
  };

  const signOut = async () => {
    try {
      const token = await getItem(AUTH_TOKEN_KEY);

      if (token) {
        try {
          await authApi.signOut();
        } catch (fetchError) {
          // Ignore errors during signout, continue with local cleanup
          console.warn('AuthContext - Error during signout, continuing with local cleanup');
        }
      }
    } catch (error) {
      console.error('Sign out error:', error);
    } finally {
      // Clear local storage regardless of backend response
      await multiRemove([AUTH_TOKEN_KEY, PLAYER_DATA_KEY]);
      setPlayer(null);
    }
  };

  const uploadAvatar = async (imageUri: string) => {
    console.log('AuthContext - Starting uploadAvatar');
    try {
      const response = await profileApi.uploadAvatar(imageUri);
      console.log('AuthContext - Avatar uploaded successfully:', response.avatarUrl);

      // Update local player state with new avatar URL
      if (player) {
        const updatedPlayer = { ...player, avatarUrl: response.avatarUrl };
        setPlayer(updatedPlayer);
        await setItem(PLAYER_DATA_KEY, JSON.stringify(updatedPlayer));
      }
    } catch (error) {
      console.error('AuthContext - Upload avatar error:', error);
      throw error;
    }
  };

  const deleteAvatar = async () => {
    console.log('AuthContext - Starting deleteAvatar');
    try {
      await profileApi.deleteAvatar();
      console.log('AuthContext - Avatar deleted successfully');

      // Update local player state to remove avatar URL
      if (player) {
        const updatedPlayer = { ...player, avatarUrl: null };
        setPlayer(updatedPlayer);
        await setItem(PLAYER_DATA_KEY, JSON.stringify(updatedPlayer));
      }
    } catch (error) {
      console.error('AuthContext - Delete avatar error:', error);
      throw error;
    }
  };

  const value: AuthContextType = {
    player,
    isLoading,
    signIn,
    signOut,
    checkAuth,
    setPlayer,
    // Avatar methods
    uploadAvatar,
    deleteAvatar,
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
