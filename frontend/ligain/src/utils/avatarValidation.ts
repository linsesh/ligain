import { AvatarErrorCode } from '../api/types';

const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB
const MIN_DIMENSION = 100; // 100px minimum width and height

export interface AvatarValidationResult {
  valid: boolean;
  error?: AvatarErrorCode;
}

export interface AvatarAssetMetadata {
  fileSize?: number;
  width?: number;
  height?: number;
}

export async function validateAvatarImage(uri: string, metadata: AvatarAssetMetadata): Promise<AvatarValidationResult> {
  const { fileSize, width, height } = metadata;

  if (fileSize !== undefined && fileSize > MAX_FILE_SIZE) {
    return { valid: false, error: 'FILE_TOO_LARGE' };
  }

  if (!width || !height) {
    return { valid: false, error: 'INVALID_IMAGE' };
  }

  if (width < MIN_DIMENSION || height < MIN_DIMENSION) {
    return { valid: false, error: 'IMAGE_TOO_SMALL' };
  }

  return { valid: true };
}

