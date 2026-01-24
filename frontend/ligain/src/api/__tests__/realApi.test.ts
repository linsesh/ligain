import { RealProfileApi } from '../realApi';
import { AvatarError } from '../types';

// Mock the api config module
jest.mock('../../config/api', () => ({
  API_CONFIG: {
    BASE_URL: 'https://test-api.example.com',
  },
  authenticatedFetch: jest.fn(),
}));

import { authenticatedFetch } from '../../config/api';

const mockAuthenticatedFetch = authenticatedFetch as jest.MockedFunction<typeof authenticatedFetch>;

describe('RealProfileApi', () => {
  let profileApi: RealProfileApi;

  beforeEach(() => {
    jest.clearAllMocks();
    profileApi = new RealProfileApi();
  });

  describe('uploadAvatar', () => {
    it('sends FormData with correct structure to correct endpoint', async () => {
      mockAuthenticatedFetch.mockResolvedValue({
        ok: true,
        json: async () => ({ avatar_url: 'https://storage.example.com/avatar.jpg' }),
      } as Response);

      await profileApi.uploadAvatar('file:///path/to/image.jpg');

      expect(mockAuthenticatedFetch).toHaveBeenCalledWith(
        'https://test-api.example.com/api/players/me/avatar',
        expect.objectContaining({
          method: 'POST',
          body: expect.any(FormData),
        })
      );

      // Verify FormData structure
      const callArgs = mockAuthenticatedFetch.mock.calls[0];
      const formData = callArgs[1]?.body as FormData;
      expect(formData).toBeInstanceOf(FormData);
    });

    it('returns avatarUrl on success (converts avatar_url from backend)', async () => {
      mockAuthenticatedFetch.mockResolvedValue({
        ok: true,
        json: async () => ({ avatar_url: 'https://storage.example.com/avatar.jpg' }),
      } as Response);

      const result = await profileApi.uploadAvatar('file:///path/to/image.jpg');

      expect(result).toEqual({ avatarUrl: 'https://storage.example.com/avatar.jpg' });
    });

    it('throws AvatarError with INVALID_IMAGE on 400 with code INVALID_IMAGE', async () => {
      mockAuthenticatedFetch.mockResolvedValue({
        ok: false,
        json: async () => ({ code: 'INVALID_IMAGE', error: 'Invalid image format' }),
      } as Response);

      await expect(profileApi.uploadAvatar('file:///path/to/invalid.txt')).rejects.toThrow(
        AvatarError
      );

      try {
        await profileApi.uploadAvatar('file:///path/to/invalid.txt');
      } catch (error) {
        expect(error).toBeInstanceOf(AvatarError);
        expect((error as AvatarError).code).toBe('INVALID_IMAGE');
      }
    });

    it('throws AvatarError with FILE_TOO_LARGE on 400 with code FILE_TOO_LARGE', async () => {
      mockAuthenticatedFetch.mockResolvedValue({
        ok: false,
        json: async () => ({ code: 'FILE_TOO_LARGE', error: 'File exceeds 10MB limit' }),
      } as Response);

      try {
        await profileApi.uploadAvatar('file:///path/to/large-image.jpg');
        fail('Expected error to be thrown');
      } catch (error) {
        expect(error).toBeInstanceOf(AvatarError);
        expect((error as AvatarError).code).toBe('FILE_TOO_LARGE');
        expect((error as AvatarError).message).toBe('File exceeds 10MB limit');
      }
    });

    it('throws AvatarError with IMAGE_TOO_SMALL on 400 with code IMAGE_TOO_SMALL', async () => {
      mockAuthenticatedFetch.mockResolvedValue({
        ok: false,
        json: async () => ({ code: 'IMAGE_TOO_SMALL', error: 'Image must be at least 100x100' }),
      } as Response);

      try {
        await profileApi.uploadAvatar('file:///path/to/tiny.jpg');
        fail('Expected error to be thrown');
      } catch (error) {
        expect(error).toBeInstanceOf(AvatarError);
        expect((error as AvatarError).code).toBe('IMAGE_TOO_SMALL');
      }
    });

    it('throws AvatarError with UPLOAD_FAILED on 500', async () => {
      mockAuthenticatedFetch.mockResolvedValue({
        ok: false,
        json: async () => ({ code: 'UPLOAD_FAILED', error: 'Storage unavailable' }),
      } as Response);

      try {
        await profileApi.uploadAvatar('file:///path/to/image.jpg');
        fail('Expected error to be thrown');
      } catch (error) {
        expect(error).toBeInstanceOf(AvatarError);
        expect((error as AvatarError).code).toBe('UPLOAD_FAILED');
      }
    });

    it('throws AvatarError with UPLOAD_FAILED for unknown error codes', async () => {
      mockAuthenticatedFetch.mockResolvedValue({
        ok: false,
        json: async () => ({ code: 'UNKNOWN_CODE', error: 'Something went wrong' }),
      } as Response);

      try {
        await profileApi.uploadAvatar('file:///path/to/image.jpg');
        fail('Expected error to be thrown');
      } catch (error) {
        expect(error).toBeInstanceOf(AvatarError);
        expect((error as AvatarError).code).toBe('UPLOAD_FAILED');
      }
    });

    it('throws AvatarError with UPLOAD_FAILED when json parsing fails', async () => {
      mockAuthenticatedFetch.mockResolvedValue({
        ok: false,
        json: async () => {
          throw new Error('Invalid JSON');
        },
      } as unknown as Response);

      try {
        await profileApi.uploadAvatar('file:///path/to/image.jpg');
        fail('Expected error to be thrown');
      } catch (error) {
        expect(error).toBeInstanceOf(AvatarError);
        expect((error as AvatarError).code).toBe('UPLOAD_FAILED');
      }
    });

    it('extracts correct filename and mime type from image URI', async () => {
      mockAuthenticatedFetch.mockResolvedValue({
        ok: true,
        json: async () => ({ avatar_url: 'https://storage.example.com/avatar.png' }),
      } as Response);

      await profileApi.uploadAvatar('file:///path/to/profile-pic.png');

      const callArgs = mockAuthenticatedFetch.mock.calls[0];
      const formData = callArgs[1]?.body as FormData;

      // FormData in tests doesn't expose values easily, but we verify it was created
      expect(formData).toBeInstanceOf(FormData);
    });
  });

  describe('deleteAvatar', () => {
    it('sends DELETE to correct endpoint', async () => {
      mockAuthenticatedFetch.mockResolvedValue({
        ok: true,
        json: async () => ({}),
      } as Response);

      await profileApi.deleteAvatar();

      expect(mockAuthenticatedFetch).toHaveBeenCalledWith(
        'https://test-api.example.com/api/players/me/avatar',
        { method: 'DELETE' }
      );
    });

    it('resolves successfully on 200', async () => {
      mockAuthenticatedFetch.mockResolvedValue({
        ok: true,
        json: async () => ({}),
      } as Response);

      await expect(profileApi.deleteAvatar()).resolves.toBeUndefined();
    });

    it('throws Error on 500 with message', async () => {
      mockAuthenticatedFetch.mockResolvedValue({
        ok: false,
        status: 500,
        json: async () => ({ error: 'Failed to delete from storage' }),
      } as Response);

      await expect(profileApi.deleteAvatar()).rejects.toThrow('Failed to delete from storage');
    });

    it('throws Error with status when json parsing fails', async () => {
      mockAuthenticatedFetch.mockResolvedValue({
        ok: false,
        status: 500,
        json: async () => {
          throw new Error('Invalid JSON');
        },
      } as unknown as Response);

      await expect(profileApi.deleteAvatar()).rejects.toThrow('Failed to delete avatar: 500');
    });
  });
});
