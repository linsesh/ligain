import { useRouter } from 'expo-router';
import * as Haptics from 'expo-haptics';

/**
 * Determines the post-bet navigation action:
 * - If there are remaining unbet matches in the matchday, navigate to the next one.
 * - If all matches are bet, navigate back to the match list.
 */
export function usePostBetNavigation(
  remainingCount: number,
  navigateToNextMatch: () => void,
) {
  const router = useRouter();
  const isLastMatch = remainingCount === 0;

  const handlePress = () => {
    Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Medium);
    if (isLastMatch) {
      router.replace('/(tabs)/matches');
    } else {
      navigateToNextMatch();
    }
  };

  return { handlePress, isLastMatch };
}
