import { Alert } from 'react-native';
import { useGamesApi } from '../src/api';
import { useGames } from '../src/contexts/GamesContext';
import { useTranslation } from '../src/hooks/useTranslation';
import { useAutoReplicateBets } from '../src/hooks/useAutoReplicateBets';
import { useBetSubmission } from './useBetSubmission';

export const useBetPlacement = (gameId: string, onFail?: (matchId: string) => void) => {
  const gamesApi = useGamesApi();
  const { games } = useGames();
  const { t } = useTranslation();
  const { enabled: autoReplicateEnabled } = useAutoReplicateBets();
  const { submitBet, isSubmitting, error, lastFailedMatchId } = useBetSubmission(gameId, onFail);

  const placeBet = async (matchId: string, homeGoals: number, awayGoals: number): Promise<void> => {
    // Primary bet — throws on failure (useBetSubmission handles error state + onFail)
    await submitBet(matchId, homeGoals, awayGoals);

    if (!autoReplicateEnabled) return;

    const currentGame = games.find(g => g.gameId === gameId);
    if (!currentGame) return;

    const siblingGames = games.filter(
      g =>
        g.gameId !== gameId &&
        g.seasonYear === currentGame.seasonYear &&
        g.competitionName === currentGame.competitionName
    );
    if (siblingGames.length === 0) return;

    const failedGameNames: string[] = [];
    for (const sibling of siblingGames) {
      try {
        await gamesApi.placeBet(sibling.gameId, matchId, homeGoals, awayGoals);
      } catch {
        failedGameNames.push(sibling.name);
      }
    }

    if (failedGameNames.length > 0) {
      Alert.alert(
        t('games.betReplicationFailed'),
        t('games.betReplicationFailedMessage', { gameNames: failedGameNames.join(', ') })
      );
    }
  };

  return { placeBet, isSubmitting, error, lastFailedMatchId };
};
