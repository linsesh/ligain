import { renderHook, act } from '@testing-library/react-native';
import * as Haptics from 'expo-haptics';
import { usePostBetNavigation } from '../usePostBetNavigation';

const mockReplace = jest.fn();

jest.mock('expo-router', () => ({
  useRouter: () => ({ replace: mockReplace }),
}));

jest.mock('expo-haptics', () => ({
  impactAsync: jest.fn(),
  ImpactFeedbackStyle: { Medium: 'medium' },
}));

describe('usePostBetNavigation', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('isLastMatch', () => {
    it('is false when remainingCount > 0', () => {
      const { result } = renderHook(() => usePostBetNavigation(3, jest.fn()));
      expect(result.current.isLastMatch).toBe(false);
    });

    it('is true when remainingCount === 0', () => {
      const { result } = renderHook(() => usePostBetNavigation(0, jest.fn()));
      expect(result.current.isLastMatch).toBe(true);
    });
  });

  describe('handlePress — navigation', () => {
    it('calls navigateToNextMatch when there are remaining matches', () => {
      const navigateToNextMatch = jest.fn();
      const { result } = renderHook(() => usePostBetNavigation(2, navigateToNextMatch));
      act(() => { result.current.handlePress(); });
      expect(navigateToNextMatch).toHaveBeenCalledTimes(1);
      expect(mockReplace).not.toHaveBeenCalled();
    });

    it('navigates to match list when no matches remain', () => {
      const navigateToNextMatch = jest.fn();
      const { result } = renderHook(() => usePostBetNavigation(0, navigateToNextMatch));
      act(() => { result.current.handlePress(); });
      expect(mockReplace).toHaveBeenCalledWith('/(tabs)/matches');
      expect(navigateToNextMatch).not.toHaveBeenCalled();
    });
  });

  describe('handlePress — haptics', () => {
    it('triggers haptic feedback on every press', () => {
      const { result } = renderHook(() => usePostBetNavigation(1, jest.fn()));
      act(() => { result.current.handlePress(); });
      expect(Haptics.impactAsync).toHaveBeenCalledTimes(1);
      expect(Haptics.impactAsync).toHaveBeenCalledWith('medium');
    });

    it('triggers haptic feedback even when navigating to match list', () => {
      const { result } = renderHook(() => usePostBetNavigation(0, jest.fn()));
      act(() => { result.current.handlePress(); });
      expect(Haptics.impactAsync).toHaveBeenCalledTimes(1);
    });
  });
});
