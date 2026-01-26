import React from 'react';
import { renderHook, act } from '@testing-library/react-native';
import { UpdateRequiredProvider, useUpdateRequired } from '../UpdateRequiredContext';

// Mock the api config module
jest.mock('../../config/api', () => ({
  setVersionOutdatedCallback: jest.fn(),
}));

describe('UpdateRequiredContext', () => {
  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <UpdateRequiredProvider>{children}</UpdateRequiredProvider>
  );

  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('initial state', () => {
    it('isUpdateRequired should be false initially', () => {
      const { result } = renderHook(() => useUpdateRequired(), { wrapper });
      expect(result.current.isUpdateRequired).toBe(false);
    });

    it('storeUrl should be null initially', () => {
      const { result } = renderHook(() => useUpdateRequired(), { wrapper });
      expect(result.current.storeUrl).toBeNull();
    });

    it('minVersion should be null initially', () => {
      const { result } = renderHook(() => useUpdateRequired(), { wrapper });
      expect(result.current.minVersion).toBeNull();
    });
  });

  describe('setUpdateRequired', () => {
    it('should set isUpdateRequired to true', () => {
      const { result } = renderHook(() => useUpdateRequired(), { wrapper });

      act(() => {
        result.current.setUpdateRequired('https://apps.apple.com/app/test', '1.5.0');
      });

      expect(result.current.isUpdateRequired).toBe(true);
    });

    it('should set storeUrl correctly', () => {
      const { result } = renderHook(() => useUpdateRequired(), { wrapper });
      const testUrl = 'https://apps.apple.com/fr/app/ligain/id6748531523';

      act(() => {
        result.current.setUpdateRequired(testUrl, '1.5.0');
      });

      expect(result.current.storeUrl).toBe(testUrl);
    });

    it('should set minVersion correctly', () => {
      const { result } = renderHook(() => useUpdateRequired(), { wrapper });
      const testVersion = '2.0.0';

      act(() => {
        result.current.setUpdateRequired('https://example.com', testVersion);
      });

      expect(result.current.minVersion).toBe(testVersion);
    });
  });

  describe('context provider requirement', () => {
    it('should throw error when used outside provider', () => {
      // Suppress console.error for this test since we expect an error
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation(() => {});

      expect(() => {
        renderHook(() => useUpdateRequired());
      }).toThrow('useUpdateRequired must be used within an UpdateRequiredProvider');

      consoleSpy.mockRestore();
    });
  });

  describe('callback registration', () => {
    it('should register callback with api module on mount', () => {
      const { setVersionOutdatedCallback } = require('../../config/api');

      renderHook(() => useUpdateRequired(), { wrapper });

      expect(setVersionOutdatedCallback).toHaveBeenCalledTimes(1);
      expect(setVersionOutdatedCallback).toHaveBeenCalledWith(expect.any(Function));
    });
  });
});
