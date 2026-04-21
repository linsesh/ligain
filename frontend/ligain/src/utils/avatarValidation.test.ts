import { validateAvatarImage, AvatarValidationResult } from './avatarValidation';

describe('validateAvatarImage', () => {
  const uri = 'file:///test/image.jpg';

  it('returns valid for a proper image', async () => {
    const result = await validateAvatarImage(uri, { fileSize: 5 * 1024 * 1024, width: 500, height: 500 });
    expect(result).toEqual<AvatarValidationResult>({ valid: true });
  });

  it('returns FILE_TOO_LARGE for images over 10MB', async () => {
    const result = await validateAvatarImage(uri, { fileSize: 15 * 1024 * 1024, width: 500, height: 500 });
    expect(result).toEqual<AvatarValidationResult>({ valid: false, error: 'FILE_TOO_LARGE' });
  });

  it('accepts images exactly at 10MB limit', async () => {
    const result = await validateAvatarImage(uri, { fileSize: 10 * 1024 * 1024, width: 500, height: 500 });
    expect(result).toEqual<AvatarValidationResult>({ valid: true });
  });

  it('returns IMAGE_TOO_SMALL for images smaller than 100x100', async () => {
    const result = await validateAvatarImage(uri, { fileSize: 1024, width: 50, height: 50 });
    expect(result).toEqual<AvatarValidationResult>({ valid: false, error: 'IMAGE_TOO_SMALL' });
  });

  it('returns IMAGE_TOO_SMALL if only width is too small', async () => {
    const result = await validateAvatarImage(uri, { fileSize: 1024, width: 50, height: 200 });
    expect(result).toEqual<AvatarValidationResult>({ valid: false, error: 'IMAGE_TOO_SMALL' });
  });

  it('returns IMAGE_TOO_SMALL if only height is too small', async () => {
    const result = await validateAvatarImage(uri, { fileSize: 1024, width: 200, height: 50 });
    expect(result).toEqual<AvatarValidationResult>({ valid: false, error: 'IMAGE_TOO_SMALL' });
  });

  it('accepts images exactly at 100x100 minimum', async () => {
    const result = await validateAvatarImage(uri, { fileSize: 1024, width: 100, height: 100 });
    expect(result).toEqual<AvatarValidationResult>({ valid: true });
  });

  it('returns INVALID_IMAGE when dimensions are missing', async () => {
    const result = await validateAvatarImage(uri, { fileSize: 1024 });
    expect(result).toEqual<AvatarValidationResult>({ valid: false, error: 'INVALID_IMAGE' });
  });

  it('skips file size check when fileSize is undefined', async () => {
    const result = await validateAvatarImage(uri, { width: 500, height: 500 });
    expect(result).toEqual<AvatarValidationResult>({ valid: true });
  });
});
