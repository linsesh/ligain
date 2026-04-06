import { renderHook, act } from '@testing-library/react-native';
import { Alert } from 'react-native';
import { useBetPlacement } from '../useBetPlacement';

// Mock dependencies
jest.mock('react-native', () => ({
  Alert: { alert: jest.fn() },
}));

const mockSubmitBet = jest.fn();
jest.mock('../useBetSubmission', () => ({
  useBetSubmission: jest.fn((gameId: string, onFail?: (matchId: string) => void) => ({
    submitBet: mockSubmitBet,
    isSubmitting: false,
    error: null,
    lastFailedMatchId: null,
  })),
}));

const mockPlaceBet = jest.fn();
jest.mock('../../src/api', () => ({
  useGamesApi: () => ({
    placeBet: mockPlaceBet,
  }),
}));

let mockGames: any[] = [];
jest.mock('../../src/contexts/GamesContext', () => ({
  useGames: () => ({ games: mockGames }),
}));

let mockAutoReplicateEnabled = true;
jest.mock('../../src/hooks/useAutoReplicateBets', () => ({
  useAutoReplicateBets: () => ({ enabled: mockAutoReplicateEnabled, isLoading: false, toggle: jest.fn() }),
}));

jest.mock('../../src/hooks/useTranslation', () => ({
  useTranslation: () => ({
    t: (key: string, params?: any) => {
      if (params?.gameNames) return `${key}:${params.gameNames}`;
      return key;
    },
  }),
}));

describe('useBetPlacement', () => {
  const GAME_ID = 'game-1';
  const MATCH_ID = 'match-1';
  const HOME_GOALS = 2;
  const AWAY_GOALS = 1;

  const CURRENT_GAME = {
    gameId: GAME_ID,
    name: 'My Game',
    seasonYear: '2025/2026',
    competitionName: 'Ligue 1',
    status: 'active',
  };

  const SIBLING_GAME = {
    gameId: 'game-2',
    name: 'Friends Game',
    seasonYear: '2025/2026',
    competitionName: 'Ligue 1',
    status: 'active',
  };

  const ANOTHER_SIBLING = {
    gameId: 'game-3',
    name: 'Work Game',
    seasonYear: '2025/2026',
    competitionName: 'Ligue 1',
    status: 'active',
  };

  beforeEach(() => {
    jest.clearAllMocks();
    mockGames = [CURRENT_GAME];
    mockAutoReplicateEnabled = true;
    mockSubmitBet.mockResolvedValue(undefined);
    mockPlaceBet.mockResolvedValue(undefined);
  });

  it('places primary bet with no siblings and shows no alert', async () => {
    const { result } = renderHook(() => useBetPlacement(GAME_ID));

    await act(async () => {
      await result.current.placeBet(MATCH_ID, HOME_GOALS, AWAY_GOALS);
    });

    expect(mockSubmitBet).toHaveBeenCalledWith(MATCH_ID, HOME_GOALS, AWAY_GOALS);
    expect(mockPlaceBet).not.toHaveBeenCalled();
    expect(Alert.alert).not.toHaveBeenCalled();
  });

  it('does not replicate when auto-replicate is disabled', async () => {
    mockAutoReplicateEnabled = false;
    mockGames = [CURRENT_GAME, SIBLING_GAME];

    const { result } = renderHook(() => useBetPlacement(GAME_ID));

    await act(async () => {
      await result.current.placeBet(MATCH_ID, HOME_GOALS, AWAY_GOALS);
    });

    expect(mockSubmitBet).toHaveBeenCalledWith(MATCH_ID, HOME_GOALS, AWAY_GOALS);
    expect(mockPlaceBet).not.toHaveBeenCalled();
    expect(Alert.alert).not.toHaveBeenCalled();
  });

  it('replicates to sibling game silently on success', async () => {
    mockGames = [CURRENT_GAME, SIBLING_GAME];

    const { result } = renderHook(() => useBetPlacement(GAME_ID));

    await act(async () => {
      await result.current.placeBet(MATCH_ID, HOME_GOALS, AWAY_GOALS);
    });

    expect(mockSubmitBet).toHaveBeenCalledWith(MATCH_ID, HOME_GOALS, AWAY_GOALS);
    expect(mockPlaceBet).toHaveBeenCalledWith(SIBLING_GAME.gameId, MATCH_ID, HOME_GOALS, AWAY_GOALS);
    expect(Alert.alert).not.toHaveBeenCalled();
  });

  it('shows alert with sibling name when replication fails for 1 game', async () => {
    mockGames = [CURRENT_GAME, SIBLING_GAME];
    mockPlaceBet.mockRejectedValue(new Error('Network error'));

    const { result } = renderHook(() => useBetPlacement(GAME_ID));

    await act(async () => {
      await result.current.placeBet(MATCH_ID, HOME_GOALS, AWAY_GOALS);
    });

    expect(mockSubmitBet).toHaveBeenCalledWith(MATCH_ID, HOME_GOALS, AWAY_GOALS);
    expect(Alert.alert).toHaveBeenCalledWith(
      'games.betReplicationFailed',
      expect.stringContaining(SIBLING_GAME.name)
    );
  });

  it('shows alert listing only failed games when one of two siblings fails', async () => {
    mockGames = [CURRENT_GAME, SIBLING_GAME, ANOTHER_SIBLING];
    mockPlaceBet
      .mockResolvedValueOnce(undefined) // sibling 1 succeeds
      .mockRejectedValueOnce(new Error('Network error')); // sibling 2 fails

    const { result } = renderHook(() => useBetPlacement(GAME_ID));

    await act(async () => {
      await result.current.placeBet(MATCH_ID, HOME_GOALS, AWAY_GOALS);
    });

    expect(Alert.alert).toHaveBeenCalledWith(
      'games.betReplicationFailed',
      expect.stringContaining(ANOTHER_SIBLING.name)
    );
    expect((Alert.alert as jest.Mock).mock.calls[0][1]).not.toContain(SIBLING_GAME.name);
  });

  it('shows alert listing all game names when all replications fail', async () => {
    mockGames = [CURRENT_GAME, SIBLING_GAME, ANOTHER_SIBLING];
    mockPlaceBet.mockRejectedValue(new Error('Network error'));

    const { result } = renderHook(() => useBetPlacement(GAME_ID));

    await act(async () => {
      await result.current.placeBet(MATCH_ID, HOME_GOALS, AWAY_GOALS);
    });

    expect(Alert.alert).toHaveBeenCalledWith(
      'games.betReplicationFailed',
      expect.stringContaining(SIBLING_GAME.name)
    );
    expect((Alert.alert as jest.Mock).mock.calls[0][1]).toContain(ANOTHER_SIBLING.name);
  });

  it('does not attempt replication when primary bet fails', async () => {
    mockGames = [CURRENT_GAME, SIBLING_GAME];
    mockSubmitBet.mockRejectedValue(new Error('Primary bet failed'));

    const onFail = jest.fn();
    const { result } = renderHook(() => useBetPlacement(GAME_ID, onFail));

    await act(async () => {
      try {
        await result.current.placeBet(MATCH_ID, HOME_GOALS, AWAY_GOALS);
      } catch {
        // expected
      }
    });

    expect(mockSubmitBet).toHaveBeenCalledWith(MATCH_ID, HOME_GOALS, AWAY_GOALS);
    expect(mockPlaceBet).not.toHaveBeenCalled();
    expect(Alert.alert).not.toHaveBeenCalled();
  });

  it('does not replicate to games in different leagues', async () => {
    const differentLeagueGame = {
      gameId: 'game-other',
      name: 'Other League',
      seasonYear: '2024/2025',
      competitionName: 'Ligue 1',
      status: 'active',
    };
    mockGames = [CURRENT_GAME, differentLeagueGame];

    const { result } = renderHook(() => useBetPlacement(GAME_ID));

    await act(async () => {
      await result.current.placeBet(MATCH_ID, HOME_GOALS, AWAY_GOALS);
    });

    expect(mockPlaceBet).not.toHaveBeenCalled();
    expect(Alert.alert).not.toHaveBeenCalled();
  });
});
