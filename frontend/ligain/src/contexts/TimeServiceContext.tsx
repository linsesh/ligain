import React, { createContext, useContext } from 'react';
import { TimeService, RealTimeService } from '../services/timeService';

// Create a context for the time service
const TimeServiceContext = createContext<TimeService | null>(null);

// Hook to use the time service
export const useTimeService = () => {
  const context = useContext(TimeServiceContext);
  if (!context) {
    throw new Error('useTimeService must be used within a TimeServiceProvider');
  }
  return context;
};

// Provider component
export const TimeServiceProvider: React.FC<{ service: TimeService; children: React.ReactNode }> = ({ service, children }) => {
  return (
    <TimeServiceContext.Provider value={service}>
      {children}
    </TimeServiceContext.Provider>
  );
}; 