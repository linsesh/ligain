import * as FileSystem from 'expo-file-system';
import { Image } from 'react-native';
import { AvatarErrorCode } from '../api/types';

const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB
const MIN_DIMENSION = 100; // 100px minimum width and height

export interface AvatarValidationResult {
  valid: boolean;
  error?: AvatarErrorCode;
}

/**
 * Validates an avatar image for upload.
 * Checks:
 * - File exists
 * - File size is under 10MB
 * - Image dimensions are at least 100x100 pixels
 *
 * @param uri - The local file URI of the image
 * @returns Validation result with error code if invalid
 */
export async function validateAvatarImage(uri: string): Promise<AvatarValidationResult> {
  try {
    // Check if file exists and get size
    const fileInfo = await FileSystem.getInfoAsync(uri);

    if (!fileInfo.exists) {
      return { valid: false, error: 'INVALID_IMAGE' };
    }

    // Check file size
    if (fileInfo.size && fileInfo.size > MAX_FILE_SIZE) {
      return { valid: false, error: 'FILE_TOO_LARGE' };
    }

    // Check image dimensions
    const dimensions = await getImageDimensions(uri);

    if (!dimensions) {
      return { valid: false, error: 'INVALID_IMAGE' };
    }

    if (dimensions.width < MIN_DIMENSION || dimensions.height < MIN_DIMENSION) {
      return { valid: false, error: 'IMAGE_TOO_SMALL' };
    }

    return { valid: true };
  } catch {
    return { valid: false, error: 'INVALID_IMAGE' };
  }
}

/**
 * Gets the dimensions of an image from its URI
 */
function getImageDimensions(
  uri: string
): Promise<{ width: number; height: number } | null> {
  return new Promise((resolve) => {
    Image.getSize(
      uri,
      (width, height) => resolve({ width, height }),
      () => resolve(null)
    );
  });
}
