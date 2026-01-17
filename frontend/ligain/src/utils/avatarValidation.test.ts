import { validateAvatarImage, AvatarValidationResult } from './avatarValidation';
import * as FileSystem from 'expo-file-system';
import { Image } from 'react-native';

// Mock expo-file-system
jest.mock('expo-file-system', () => ({
  getInfoAsync: jest.fn(),
}));

// Mock React Native Image
jest.mock('react-native', () => ({
  Image: {
    getSize: jest.fn(),
  },
}));

describe('validateAvatarImage', () => {
  const mockGetInfoAsync = FileSystem.getInfoAsync as jest.Mock;
  const mockGetSize = Image.getSize as jest.Mock;

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('returns valid for a proper image', async () => {
    mockGetInfoAsync.mockResolvedValue({
      exists: true,
      size: 5 * 1024 * 1024, // 5MB
    });

    mockGetSize.mockImplementation((uri, success) => {
      success(500, 500); // 500x500 pixels
    });

    const result = await validateAvatarImage('file:///test/image.jpg');

    expect(result).toEqual<AvatarValidationResult>({
      valid: true,
    });
  });

  it('returns FILE_TOO_LARGE error for images over 10MB', async () => {
    mockGetInfoAsync.mockResolvedValue({
      exists: true,
      size: 15 * 1024 * 1024, // 15MB
    });

    const result = await validateAvatarImage('file:///test/large-image.jpg');

    expect(result).toEqual<AvatarValidationResult>({
      valid: false,
      error: 'FILE_TOO_LARGE',
    });
    // Should not call getSize if file is too large
    expect(mockGetSize).not.toHaveBeenCalled();
  });

  it('returns IMAGE_TOO_SMALL error for images smaller than 100x100', async () => {
    mockGetInfoAsync.mockResolvedValue({
      exists: true,
      size: 1 * 1024 * 1024, // 1MB
    });

    mockGetSize.mockImplementation((uri, success) => {
      success(50, 50); // 50x50 pixels - too small
    });

    const result = await validateAvatarImage('file:///test/small-image.jpg');

    expect(result).toEqual<AvatarValidationResult>({
      valid: false,
      error: 'IMAGE_TOO_SMALL',
    });
  });

  it('returns IMAGE_TOO_SMALL if only width is too small', async () => {
    mockGetInfoAsync.mockResolvedValue({
      exists: true,
      size: 1 * 1024 * 1024,
    });

    mockGetSize.mockImplementation((uri, success) => {
      success(50, 200); // width too small
    });

    const result = await validateAvatarImage('file:///test/narrow-image.jpg');

    expect(result).toEqual<AvatarValidationResult>({
      valid: false,
      error: 'IMAGE_TOO_SMALL',
    });
  });

  it('returns IMAGE_TOO_SMALL if only height is too small', async () => {
    mockGetInfoAsync.mockResolvedValue({
      exists: true,
      size: 1 * 1024 * 1024,
    });

    mockGetSize.mockImplementation((uri, success) => {
      success(200, 50); // height too small
    });

    const result = await validateAvatarImage('file:///test/short-image.jpg');

    expect(result).toEqual<AvatarValidationResult>({
      valid: false,
      error: 'IMAGE_TOO_SMALL',
    });
  });

  it('returns INVALID_IMAGE error when file does not exist', async () => {
    mockGetInfoAsync.mockResolvedValue({
      exists: false,
    });

    const result = await validateAvatarImage('file:///test/nonexistent.jpg');

    expect(result).toEqual<AvatarValidationResult>({
      valid: false,
      error: 'INVALID_IMAGE',
    });
  });

  it('returns INVALID_IMAGE error when getInfoAsync throws', async () => {
    mockGetInfoAsync.mockRejectedValue(new Error('File system error'));

    const result = await validateAvatarImage('file:///test/error.jpg');

    expect(result).toEqual<AvatarValidationResult>({
      valid: false,
      error: 'INVALID_IMAGE',
    });
  });

  it('returns INVALID_IMAGE error when Image.getSize fails', async () => {
    mockGetInfoAsync.mockResolvedValue({
      exists: true,
      size: 1 * 1024 * 1024,
    });

    mockGetSize.mockImplementation((uri, success, failure) => {
      failure(new Error('Cannot load image'));
    });

    const result = await validateAvatarImage('file:///test/corrupt.jpg');

    expect(result).toEqual<AvatarValidationResult>({
      valid: false,
      error: 'INVALID_IMAGE',
    });
  });

  it('accepts images exactly at 10MB limit', async () => {
    mockGetInfoAsync.mockResolvedValue({
      exists: true,
      size: 10 * 1024 * 1024, // exactly 10MB
    });

    mockGetSize.mockImplementation((uri, success) => {
      success(500, 500);
    });

    const result = await validateAvatarImage('file:///test/exact-limit.jpg');

    expect(result).toEqual<AvatarValidationResult>({
      valid: true,
    });
  });

  it('accepts images exactly at 100x100 minimum', async () => {
    mockGetInfoAsync.mockResolvedValue({
      exists: true,
      size: 1 * 1024 * 1024,
    });

    mockGetSize.mockImplementation((uri, success) => {
      success(100, 100); // exactly minimum
    });

    const result = await validateAvatarImage('file:///test/min-size.jpg');

    expect(result).toEqual<AvatarValidationResult>({
      valid: true,
    });
  });
});
