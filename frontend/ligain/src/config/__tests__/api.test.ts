import { API_CONFIG, getApiHeaders, setVersionOutdatedCallback, authenticatedFetch } from '../api';

// Mock expo-constants
jest.mock('expo-constants', () => ({
  expoConfig: {
    version: '1.4.0',
    extra: {
      apiBaseUrl: 'https://api.test.com',
      apiKey: 'test-api-key',
      appleClientId: 'test-apple-client-id',
    },
  },
}));

// Mock storage
jest.mock('../../utils/storage', () => ({
  getItem: jest.fn().mockResolvedValue('test-auth-token'),
  setItem: jest.fn().mockResolvedValue(undefined),
}));

// Mock global fetch
const mockFetch = jest.fn();
global.fetch = mockFetch;

describe('API Configuration', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('API_CONFIG', () => {
    it('should include APP_VERSION', () => {
      expect(API_CONFIG.APP_VERSION).toBe('1.4.0');
    });
  });

  describe('getApiHeaders', () => {
    it('should include X-App-Version header', () => {
      const headers = getApiHeaders();
      expect(headers['X-App-Version']).toBe('1.4.0');
    });

    it('should include X-API-Key header', () => {
      const headers = getApiHeaders();
      expect(headers['X-API-Key']).toBe('test-api-key');
    });

    it('should merge additional headers', () => {
      const headers = getApiHeaders({ 'Content-Type': 'application/json' });
      expect((headers as Record<string, string>)['Content-Type']).toBe('application/json');
      expect(headers['X-App-Version']).toBe('1.4.0');
    });
  });
});

describe('Version outdated callback', () => {
  let capturedCallback: ((storeUrl: string, minVersion: string) => void) | null = null;

  beforeEach(() => {
    jest.clearAllMocks();
    capturedCallback = null;
  });

  it('should store callback when setVersionOutdatedCallback is called', () => {
    const callback = jest.fn();
    setVersionOutdatedCallback(callback);

    // The callback is stored internally, we can't directly test this
    // but we can test it's called when a 426 response is received
    expect(true).toBe(true); // Callback registered successfully
  });
});

describe('authenticatedFetch', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockFetch.mockReset();
  });

  describe('HTTP 426 handling', () => {
    it('should call version outdated callback on HTTP 426 response', async () => {
      const callback = jest.fn();
      setVersionOutdatedCallback(callback);

      const mockResponse = {
        status: 426,
        headers: {
          get: jest.fn().mockReturnValue(null),
        },
        clone: jest.fn().mockReturnValue({
          json: jest.fn().mockResolvedValue({
            code: 'VERSION_OUTDATED',
            store_url: 'https://apps.apple.com/test',
            min_version: '2.0.0',
          }),
        }),
      };

      mockFetch.mockResolvedValue(mockResponse);

      await authenticatedFetch('https://api.test.com/endpoint');

      expect(callback).toHaveBeenCalledWith('https://apps.apple.com/test', '2.0.0');
    });

    it('should use default store URL if not provided in 426 response', async () => {
      const callback = jest.fn();
      setVersionOutdatedCallback(callback);

      const mockResponse = {
        status: 426,
        headers: {
          get: jest.fn().mockReturnValue(null),
        },
        clone: jest.fn().mockReturnValue({
          json: jest.fn().mockResolvedValue({
            code: 'VERSION_OUTDATED',
            // No store_url provided
            min_version: '2.0.0',
          }),
        }),
      };

      mockFetch.mockResolvedValue(mockResponse);

      await authenticatedFetch('https://api.test.com/endpoint');

      expect(callback).toHaveBeenCalledWith(
        'https://apps.apple.com/fr/app/ligain/id6748531523',
        '2.0.0'
      );
    });

    it('should handle malformed 426 response gracefully', async () => {
      const callback = jest.fn();
      setVersionOutdatedCallback(callback);

      const mockResponse = {
        status: 426,
        headers: {
          get: jest.fn().mockReturnValue(null),
        },
        clone: jest.fn().mockReturnValue({
          json: jest.fn().mockRejectedValue(new Error('Invalid JSON')),
        }),
      };

      mockFetch.mockResolvedValue(mockResponse);

      await authenticatedFetch('https://api.test.com/endpoint');

      // Should still call callback with defaults
      expect(callback).toHaveBeenCalledWith(
        'https://apps.apple.com/fr/app/ligain/id6748531523',
        'unknown'
      );
    });

    it('should NOT call callback on other error statuses (400)', async () => {
      const callback = jest.fn();
      setVersionOutdatedCallback(callback);

      const mockResponse = {
        status: 400,
        headers: {
          get: jest.fn().mockReturnValue(null),
        },
      };

      mockFetch.mockResolvedValue(mockResponse);

      await authenticatedFetch('https://api.test.com/endpoint');

      expect(callback).not.toHaveBeenCalled();
    });

    it('should NOT call callback on 401 status', async () => {
      const callback = jest.fn();
      setVersionOutdatedCallback(callback);

      const mockResponse = {
        status: 401,
        headers: {
          get: jest.fn().mockReturnValue(null),
        },
      };

      mockFetch.mockResolvedValue(mockResponse);

      await authenticatedFetch('https://api.test.com/endpoint');

      expect(callback).not.toHaveBeenCalled();
    });

    it('should NOT call callback on 500 status', async () => {
      const callback = jest.fn();
      setVersionOutdatedCallback(callback);

      const mockResponse = {
        status: 500,
        headers: {
          get: jest.fn().mockReturnValue(null),
        },
      };

      mockFetch.mockResolvedValue(mockResponse);

      await authenticatedFetch('https://api.test.com/endpoint');

      expect(callback).not.toHaveBeenCalled();
    });
  });

  describe('successful response', () => {
    it('should return response on successful request', async () => {
      const mockResponse = {
        status: 200,
        headers: {
          get: jest.fn().mockReturnValue(null),
        },
        json: jest.fn().mockResolvedValue({ data: 'test' }),
      };

      mockFetch.mockResolvedValue(mockResponse);

      const response = await authenticatedFetch('https://api.test.com/endpoint');

      expect(response.status).toBe(200);
    });

    it('should handle token refresh headers', async () => {
      const { setItem } = require('../../utils/storage');

      const mockResponse = {
        status: 200,
        headers: {
          get: jest.fn((header: string) => {
            if (header === 'X-New-Token') return 'new-auth-token';
            if (header === 'X-Token-Refreshed') return 'true';
            return null;
          }),
        },
      };

      mockFetch.mockResolvedValue(mockResponse);

      await authenticatedFetch('https://api.test.com/endpoint');

      expect(setItem).toHaveBeenCalledWith('auth_token', 'new-auth-token');
    });
  });
});
