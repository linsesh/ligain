/**
 * Utility function to convert HTTP status codes to human-readable error messages
 */
export const getHumanReadableError = (status: number, originalError?: string): string => {
  switch (status) {
    case 401:
      return 'Authentication went wrong, please refresh the page and retry';
    case 403:
      return 'Authentication went wrong, please refresh the page and retry';
    case 404:
      return 'Service not found. Please try again later';
    case 422:
      return 'Invalid information provided. Please check your details';
    case 429:
      return 'Too many requests. Please wait a moment and try again';
    case 500:
      return 'Server error. Please try again later';
    case 502:
    case 503:
    case 504:
      return 'Service temporarily unavailable. Please try again later';
    default:
      // Keep the original error message for unknown status codes
      return originalError || `Something went wrong (${status})`;
  }
};

/**
 * Handles a fetch Response, throwing a human-readable error if not ok
 */
export async function handleApiError(response: Response): Promise<never> {
  let errorData;
  try {
    errorData = await response.json();
  } catch (parseError) {
    errorData = { error: `HTTP ${response.status}: ${response.statusText}` };
  }
  const humanReadableError = getHumanReadableError(response.status, errorData.error);
  throw new Error(humanReadableError);
} 