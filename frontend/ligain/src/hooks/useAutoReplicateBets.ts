import { useState, useEffect } from 'react';
import { getItem, setItem } from '../utils/storage';

const AUTO_REPLICATE_KEY = 'auto_replicate_bets_enabled';

export const useAutoReplicateBets = () => {
  const [enabled, setEnabled] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    getItem(AUTO_REPLICATE_KEY).then(val => {
      setEnabled(val !== null ? val === 'true' : true);
      setIsLoading(false);
    });
  }, []);

  const toggle = async (value: boolean) => {
    setEnabled(value);
    await setItem(AUTO_REPLICATE_KEY, value.toString());
  };

  return { enabled, isLoading, toggle };
};
