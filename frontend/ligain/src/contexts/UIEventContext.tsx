import React, { createContext, useContext, useState } from 'react';

interface UIEventContextType {
  openJoinOrCreate: boolean;
  setOpenJoinOrCreate: (value: boolean) => void;
}

const UIEventContext = createContext<UIEventContextType | undefined>(undefined);

export const UIEventProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [openJoinOrCreate, _setOpenJoinOrCreate] = useState(false);

  const setOpenJoinOrCreate = (value: boolean) => {
    console.log('[UIEventContext] setOpenJoinOrCreate:', value);
    _setOpenJoinOrCreate(value);
  };

  React.useEffect(() => {
    console.log('[UIEventContext] Provider render - openJoinOrCreate:', openJoinOrCreate);
  });

  return (
    <UIEventContext.Provider value={{ openJoinOrCreate, setOpenJoinOrCreate }}>
      {children}
    </UIEventContext.Provider>
  );
};

export const useUIEvent = () => {
  const ctx = useContext(UIEventContext);
  if (!ctx) throw new Error('useUIEvent must be used within a UIEventProvider');
  return ctx;
}; 