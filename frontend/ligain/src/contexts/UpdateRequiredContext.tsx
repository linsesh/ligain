import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { setVersionOutdatedCallback } from '../config/api';

interface UpdateRequiredContextType {
  isUpdateRequired: boolean;
  storeUrl: string | null;
  minVersion: string | null;
  setUpdateRequired: (storeUrl: string, minVersion: string) => void;
}

const UpdateRequiredContext = createContext<UpdateRequiredContextType | undefined>(undefined);

export const useUpdateRequired = () => {
  const context = useContext(UpdateRequiredContext);
  if (context === undefined) {
    throw new Error('useUpdateRequired must be used within an UpdateRequiredProvider');
  }
  return context;
};

interface UpdateRequiredProviderProps {
  children: ReactNode;
}

export const UpdateRequiredProvider: React.FC<UpdateRequiredProviderProps> = ({ children }) => {
  const [isUpdateRequired, setIsUpdateRequired] = useState(false);
  const [storeUrl, setStoreUrl] = useState<string | null>(null);
  const [minVersion, setMinVersion] = useState<string | null>(null);

  const setUpdateRequired = (url: string, version: string) => {
    console.log('📱 UpdateRequired - Update required, store URL:', url, 'min version:', version);
    setStoreUrl(url);
    setMinVersion(version);
    setIsUpdateRequired(true);
  };

  // Register the callback with the API module on mount
  useEffect(() => {
    setVersionOutdatedCallback(setUpdateRequired);
  }, []);

  const value: UpdateRequiredContextType = {
    isUpdateRequired,
    storeUrl,
    minVersion,
    setUpdateRequired,
  };

  return (
    <UpdateRequiredContext.Provider value={value}>
      {children}
    </UpdateRequiredContext.Provider>
  );
};
