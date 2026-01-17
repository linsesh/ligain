/**
 * API Provider with Dependency Injection
 *
 * This provider creates the appropriate API implementations based on the
 * environment configuration. The rest of the application receives the APIs
 * through context and doesn't know whether it's talking to real or mock APIs.
 */

import React, { createContext, useContext, useMemo, ReactNode } from 'react';
import Constants from 'expo-constants';
import { Api, AuthApi, GamesApi } from './types';
import { RealAuthApi, RealGamesApi } from './realApi';
import { MockAuthApi, MockGamesApi } from './mockApi';

// Context for API injection
const ApiContext = createContext<Api | null>(null);

interface ApiProviderProps {
  children: ReactNode;
}

/**
 * ApiProvider
 *
 * Wraps the application and provides API implementations via context.
 * Reads mockMode from Expo config to determine which implementations to use.
 */
export const ApiProvider: React.FC<ApiProviderProps> = ({ children }) => {
  const isMockMode = Constants.expoConfig?.extra?.mockMode === true;

  const api = useMemo<Api>(() => {
    if (isMockMode) {
      console.log('[ApiProvider] Running in MOCK mode');
      return {
        auth: new MockAuthApi(),
        games: new MockGamesApi(),
      };
    }

    console.log('[ApiProvider] Running in REAL mode');
    return {
      auth: new RealAuthApi(),
      games: new RealGamesApi(),
    };
  }, [isMockMode]);

  return <ApiContext.Provider value={api}>{children}</ApiContext.Provider>;
};

/**
 * useApi hook
 *
 * Returns the injected API implementations.
 * Components use this to access APIs without knowing the concrete implementation.
 */
export const useApi = (): Api => {
  const api = useContext(ApiContext);
  if (!api) {
    throw new Error('useApi must be used within an ApiProvider');
  }
  return api;
};

/**
 * useAuthApi hook
 *
 * Convenience hook for accessing just the Auth API
 */
export const useAuthApi = (): AuthApi => {
  const { auth } = useApi();
  return auth;
};

/**
 * useGamesApi hook
 *
 * Convenience hook for accessing just the Games API
 */
export const useGamesApi = (): GamesApi => {
  const { games } = useApi();
  return games;
};
