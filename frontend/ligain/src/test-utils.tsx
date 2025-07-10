import React, { ReactElement } from 'react';
import { render, RenderOptions } from '@testing-library/react';
import { AuthProvider } from './contexts/AuthContext';
import { TimeServiceProvider } from './contexts/TimeServiceContext';
import { RealTimeService } from './services/timeService';

// Mock time service for testing
class MockTimeService extends RealTimeService {
  private mockTime: Date = new Date('2024-01-01T12:00:00Z');
  
  now(): Date {
    return this.mockTime;
  }
  
  setMockTime(time: Date) {
    this.mockTime = time;
  }
}

// Custom render function that includes providers
const AllTheProviders = ({ children }: { children: React.ReactNode }) => {
  return (
    <AuthProvider>
      <TimeServiceProvider service={new MockTimeService()}>
        {children}
      </TimeServiceProvider>
    </AuthProvider>
  );
};

const customRender = (
  ui: ReactElement,
  options?: Omit<RenderOptions, 'wrapper'>,
) => render(ui, { wrapper: AllTheProviders, ...options });

// Re-export everything
export * from '@testing-library/react';
export { customRender as render };

// Export mock time service for tests
export { MockTimeService }; 