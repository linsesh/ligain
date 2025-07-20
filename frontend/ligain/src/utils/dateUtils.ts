import { getCurrentLocale } from '../i18n';

/**
 * Format a date to display time (HH:MM)
 */
export const formatTime = (date: Date): string => {
  return date.toLocaleTimeString(getCurrentLocale(), { 
    hour: '2-digit', 
    minute: '2-digit',
    hour12: false 
  });
};

/**
 * Format a date to display full date (e.g., "Monday, January 15")
 */
export const formatDate = (date: Date): string => {
  return date.toLocaleDateString(getCurrentLocale(), { 
    weekday: 'long', 
    month: 'long', 
    day: 'numeric' 
  });
};

/**
 * Format a date to display short date (e.g., "01/15/2024")
 */
export const formatShortDate = (date: Date): string => {
  return date.toLocaleDateString(getCurrentLocale());
};

/**
 * Format a date to display date and time
 */
export const formatDateTime = (date: Date): string => {
  return date.toLocaleString(getCurrentLocale(), {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false
  });
}; 