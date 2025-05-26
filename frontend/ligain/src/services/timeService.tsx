import React, { createContext, useContext } from 'react';

export interface TimeService {
    now(): Date;
}

export class RealTimeService implements TimeService {
    now(): Date {
        return new Date();
    }
}

export class MockTimeService implements TimeService {
    private currentTime: Date;

    constructor(initialTime: Date) {
        this.currentTime = initialTime;
    }

    now(): Date {
        return this.currentTime;
    }

    setTime(time: Date): void {
        this.currentTime = time;
    }
}

// Create a context for the time service
export const TimeServiceContext = createContext<TimeService>(new RealTimeService());

// Hook to use the time service
export const useTimeService = () => useContext(TimeServiceContext);

// Provider component
export const TimeServiceProvider: React.FC<{ service: TimeService; children: React.ReactNode }> = ({ service, children }) => {
    return (
        <TimeServiceContext.Provider value={service}>
            {children}
        </TimeServiceContext.Provider>
    );
}; 