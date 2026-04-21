import { useEffect } from 'react';
import { requestNotificationPermissionIfNeeded } from './useNotifications';

export const useFirstInstallNotificationPrompt = () => {
  useEffect(() => {
    requestNotificationPermissionIfNeeded();
  }, []);
};
