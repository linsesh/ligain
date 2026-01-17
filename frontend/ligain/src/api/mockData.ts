/**
 * Mock Data Fixtures
 *
 * Realistic mock data for UI development and testing.
 * Uses exact team names from teamLogos.ts for proper logo display.
 *
 * Available team names:
 * 'Paris Saint Germain', 'Olympique Lyonnais', 'Olympique Marseille',
 * 'Monaco', 'LOSC Lille', 'Nice', 'Lens', 'Rennes', 'Strasbourg', 'Nantes',
 * 'Brest', 'Toulouse', 'Auxerre', 'Le Havre', 'Angers SCO', 'Metz', 'Paris'
 *
 * LEADERBOARD SUMMARY (Game 1 - Entre potes, Matchdays 15-19):
 * - Marie: 25 pts (best predictor, good at home wins)
 * - Sophie: 21 pts (consistent, good at draws)
 * - TestPlayer: 18 pts (middle of the pack)
 * - Lucas: 14 pts (struggles with predictions)
 */

import { Player } from '../contexts/AuthContext';
import { Game } from '../contexts/GamesContext';
import { MatchData, MatchesResponse } from './types';

// Helper to create dates relative to now
const now = new Date();
const addDays = (days: number): Date => {
  const date = new Date(now);
  date.setDate(date.getDate() + days);
  return date;
};

const addHours = (hours: number): Date => {
  const date = new Date(now);
  date.setHours(date.getHours() + hours);
  return date;
};

const subtractDays = (days: number): Date => {
  const date = new Date(now);
  date.setDate(date.getDate() - days);
  return date;
};

// ============================================================================
// MOCK PLAYERS
// ============================================================================

export const MOCK_CURRENT_PLAYER: Player = {
  id: 'mock-player-1',
  name: 'TestPlayer',
  email: 'test@example.com',
  provider: 'mock',
  provider_id: 'mock-1',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
};

export const MOCK_PLAYERS: Player[] = [
  MOCK_CURRENT_PLAYER,
  {
    id: 'mock-player-2',
    name: 'Marie',
    email: 'marie@example.com',
    provider: 'google',
    provider_id: 'google-2',
  },
  {
    id: 'mock-player-3',
    name: 'Lucas',
    email: 'lucas@example.com',
    provider: 'apple',
    provider_id: 'apple-3',
  },
  {
    id: 'mock-player-4',
    name: 'Sophie',
    email: 'sophie@example.com',
    provider: 'google',
    provider_id: 'google-4',
  },
];

// ============================================================================
// MOCK GAMES
// ============================================================================

export const MOCK_GAMES: Game[] = [
  {
    gameId: 'game-1',
    name: 'Entre potes',
    seasonYear: '2025/2026',
    competitionName: 'Ligue 1',
    status: 'active',
    code: 'ABCD',
    players: MOCK_PLAYERS,
  },
  {
    gameId: 'game-2',
    name: 'Family League',
    seasonYear: '2025/2026',
    competitionName: 'Ligue 1',
    status: 'active',
    code: 'EFGH',
    players: [MOCK_PLAYERS[0], MOCK_PLAYERS[1], MOCK_PLAYERS[3]],
  },
  {
    gameId: 'game-3',
    name: 'Work Buddies',
    seasonYear: '2025/2026',
    competitionName: 'Ligue 1',
    status: 'active',
    code: 'IJKL',
    players: [MOCK_PLAYERS[0], MOCK_PLAYERS[2]],
  },
];

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

const createIncomingMatch = (
  id: string,
  homeTeam: string,
  awayTeam: string,
  date: Date,
  matchday: number,
  odds: [number, number, number],
  userBet?: string
): MatchData => ({
  match: {
    id,
    date: date.toISOString(),
    homeTeam,
    awayTeam,
    homeOdds: odds[0],
    drawOdds: odds[1],
    awayOdds: odds[2],
    matchday,
    status: 'scheduled',
  },
  bet: userBet ? { prediction: userBet } : undefined,
});

const createPastMatch = (
  id: string,
  homeTeam: string,
  awayTeam: string,
  date: Date,
  matchday: number,
  score: [number, number],
  bets: { playerId: string; playerName: string; prediction: string }[]
): MatchData => {
  const actualResult =
    score[0] > score[1] ? 'home' : score[0] < score[1] ? 'away' : 'draw';
  const betsWithPoints = bets.map((bet) => ({
    ...bet,
    points: bet.prediction === actualResult ? 1 : 0,
  }));

  return {
    match: {
      id,
      date: date.toISOString(),
      homeTeam,
      awayTeam,
      matchday,
      status: 'finished',
      homeScore: score[0],
      awayScore: score[1],
    },
    allBets: betsWithPoints,
  };
};

// ============================================================================
// MOCK MATCHES - GAME 1 (Entre potes) - COMPREHENSIVE DATA
// ============================================================================

// Incoming matches for Matchday 20
const incomingMatchesGame1: Record<string, MatchData> = {
  // Upcoming matches without user bets
  'match-20-1': createIncomingMatch(
    'match-20-1',
    'Paris Saint Germain',
    'Olympique Lyonnais',
    addHours(21),
    20,
    [1.45, 4.2, 6.5]
  ),
  'match-20-2': createIncomingMatch(
    'match-20-2',
    'Olympique Marseille',
    'Monaco',
    addHours(19),
    20,
    [2.1, 3.4, 3.2]
  ),
  'match-20-3': createIncomingMatch(
    'match-20-3',
    'LOSC Lille',
    'Nice',
    addHours(17),
    20,
    [1.9, 3.5, 3.8]
  ),
  'match-20-4': createIncomingMatch(
    'match-20-4',
    'Lens',
    'Rennes',
    addHours(15),
    20,
    [2.2, 3.3, 3.1]
  ),
  'match-20-5': createIncomingMatch(
    'match-20-5',
    'Strasbourg',
    'Nantes',
    addHours(15),
    20,
    [2.5, 3.2, 2.8]
  ),
  // Matches with existing user bets
  'match-20-6': createIncomingMatch(
    'match-20-6',
    'Brest',
    'Toulouse',
    addDays(2),
    20,
    [2.3, 3.2, 3.0],
    'home'
  ),
  'match-20-7': createIncomingMatch(
    'match-20-7',
    'Auxerre',
    'Le Havre',
    addDays(2),
    20,
    [2.6, 3.1, 2.7],
    'draw'
  ),
  'match-20-8': createIncomingMatch(
    'match-20-8',
    'Angers SCO',
    'Metz',
    addDays(2),
    20,
    [2.4, 3.2, 2.9],
    'away'
  ),
  'match-20-9': createIncomingMatch(
    'match-20-9',
    'Paris',
    'Lorient',
    addDays(3),
    20,
    [2.1, 3.3, 3.4]
  ),
};

// Past matches across multiple matchdays
const pastMatchesGame1: Record<string, MatchData> = {
  // ========== MATCHDAY 19 (7 days ago) ==========
  // Results: 4 home wins, 2 draws, 3 away wins
  // Points: TestPlayer: 4, Marie: 6, Lucas: 2, Sophie: 5
  'match-19-1': createPastMatch(
    'match-19-1',
    'Paris Saint Germain',
    'Olympique Marseille',
    subtractDays(7),
    19,
    [3, 1], // Home win - PSG dominates the classique
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'home' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'away' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'home' },
    ]
  ),
  'match-19-2': createPastMatch(
    'match-19-2',
    'Olympique Lyonnais',
    'Monaco',
    subtractDays(7),
    19,
    [1, 1], // Draw - tight game
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'draw' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'home' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'draw' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'draw' },
    ]
  ),
  'match-19-3': createPastMatch(
    'match-19-3',
    'LOSC Lille',
    'Lens',
    subtractDays(7),
    19,
    [2, 0], // Home win - Derby du Nord
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'home' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'home' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'draw' },
    ]
  ),
  'match-19-4': createPastMatch(
    'match-19-4',
    'Nice',
    'Rennes',
    subtractDays(7),
    19,
    [0, 2], // Away win - Rennes surprises
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'away' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'draw' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'away' },
    ]
  ),
  'match-19-5': createPastMatch(
    'match-19-5',
    'Nantes',
    'Strasbourg',
    subtractDays(7),
    19,
    [2, 2], // Draw - entertaining game
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'draw' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'away' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'home' },
    ]
  ),
  'match-19-6': createPastMatch(
    'match-19-6',
    'Toulouse',
    'Brest',
    subtractDays(6),
    19,
    [1, 0], // Home win - narrow victory
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'draw' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'home' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'away' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'home' },
    ]
  ),
  'match-19-7': createPastMatch(
    'match-19-7',
    'Le Havre',
    'Auxerre',
    subtractDays(6),
    19,
    [0, 1], // Away win
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'away' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'away' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'home' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'draw' },
    ]
  ),
  'match-19-8': createPastMatch(
    'match-19-8',
    'Metz',
    'Angers SCO',
    subtractDays(6),
    19,
    [3, 2], // Home win - high scoring
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'draw' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'home' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'away' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'home' },
    ]
  ),
  'match-19-9': createPastMatch(
    'match-19-9',
    'Paris',
    'Lorient',
    subtractDays(6),
    19,
    [1, 2], // Away win - upset
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'draw' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'away' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'draw' },
    ]
  ),

  // ========== MATCHDAY 18 (14 days ago) ==========
  // Results: 3 home wins, 3 draws, 3 away wins
  // Points: TestPlayer: 5, Marie: 6, Lucas: 3, Sophie: 4
  'match-18-1': createPastMatch(
    'match-18-1',
    'Monaco',
    'Paris Saint Germain',
    subtractDays(14),
    18,
    [2, 2], // Draw - Monaco holds PSG
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'draw' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'away' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'home' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'draw' },
    ]
  ),
  'match-18-2': createPastMatch(
    'match-18-2',
    'Rennes',
    'Olympique Lyonnais',
    subtractDays(14),
    18,
    [1, 3], // Away win - Lyon impressive
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'away' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'away' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'draw' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'away' },
    ]
  ),
  'match-18-3': createPastMatch(
    'match-18-3',
    'Toulouse',
    'LOSC Lille',
    subtractDays(14),
    18,
    [0, 2], // Away win - Lille dominant
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'away' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'home' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'away' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'draw' },
    ]
  ),
  'match-18-4': createPastMatch(
    'match-18-4',
    'Olympique Marseille',
    'Nice',
    subtractDays(14),
    18,
    [2, 1], // Home win - OM wins at Velodrome
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'home' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'draw' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'home' },
    ]
  ),
  'match-18-5': createPastMatch(
    'match-18-5',
    'Lens',
    'Nantes',
    subtractDays(14),
    18,
    [3, 0], // Home win - Lens crushes
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'home' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'home' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'draw' },
    ]
  ),
  'match-18-6': createPastMatch(
    'match-18-6',
    'Strasbourg',
    'Brest',
    subtractDays(13),
    18,
    [1, 1], // Draw
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'draw' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'draw' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'away' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'home' },
    ]
  ),
  'match-18-7': createPastMatch(
    'match-18-7',
    'Auxerre',
    'Metz',
    subtractDays(13),
    18,
    [2, 1], // Home win
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'draw' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'home' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'away' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'draw' },
    ]
  ),
  'match-18-8': createPastMatch(
    'match-18-8',
    'Angers SCO',
    'Le Havre',
    subtractDays(13),
    18,
    [0, 0], // Draw - goalless
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'draw' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'draw' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'away' },
    ]
  ),
  'match-18-9': createPastMatch(
    'match-18-9',
    'Lorient',
    'Paris',
    subtractDays(13),
    18,
    [2, 3], // Away win - thriller
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'away' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'home' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'draw' },
    ]
  ),

  // ========== MATCHDAY 17 (21 days ago) ==========
  // Results: 5 home wins, 2 draws, 2 away wins
  // Points: TestPlayer: 4, Marie: 5, Lucas: 4, Sophie: 5
  'match-17-1': createPastMatch(
    'match-17-1',
    'Paris Saint Germain',
    'Monaco',
    subtractDays(21),
    17,
    [4, 1], // Home win - PSG dominant
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'home' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'draw' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'home' },
    ]
  ),
  'match-17-2': createPastMatch(
    'match-17-2',
    'Olympique Lyonnais',
    'Olympique Marseille',
    subtractDays(21),
    17,
    [2, 2], // Draw - Olympico thriller
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'draw' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'away' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'draw' },
    ]
  ),
  'match-17-3': createPastMatch(
    'match-17-3',
    'LOSC Lille',
    'Rennes',
    subtractDays(21),
    17,
    [1, 0], // Home win
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'home' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'home' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'away' },
    ]
  ),
  'match-17-4': createPastMatch(
    'match-17-4',
    'Nice',
    'Lens',
    subtractDays(21),
    17,
    [0, 1], // Away win
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'draw' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'away' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'away' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'home' },
    ]
  ),
  'match-17-5': createPastMatch(
    'match-17-5',
    'Nantes',
    'Toulouse',
    subtractDays(20),
    17,
    [2, 0], // Home win
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'away' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'home' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'home' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'home' },
    ]
  ),
  'match-17-6': createPastMatch(
    'match-17-6',
    'Brest',
    'Strasbourg',
    subtractDays(20),
    17,
    [3, 1], // Home win
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'draw' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'home' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'home' },
    ]
  ),
  'match-17-7': createPastMatch(
    'match-17-7',
    'Metz',
    'Auxerre',
    subtractDays(20),
    17,
    [1, 1], // Draw
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'home' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'draw' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'draw' },
    ]
  ),
  'match-17-8': createPastMatch(
    'match-17-8',
    'Le Havre',
    'Angers SCO',
    subtractDays(20),
    17,
    [2, 0], // Home win
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'draw' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'home' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'away' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'home' },
    ]
  ),
  'match-17-9': createPastMatch(
    'match-17-9',
    'Paris',
    'Lorient',
    subtractDays(20),
    17,
    [0, 2], // Away win
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'away' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'draw' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'away' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'away' },
    ]
  ),

  // ========== MATCHDAY 16 (28 days ago) ==========
  // Results: 4 home wins, 1 draw, 4 away wins
  // Points: TestPlayer: 3, Marie: 5, Lucas: 2, Sophie: 4
  'match-16-1': createPastMatch(
    'match-16-1',
    'Monaco',
    'Olympique Lyonnais',
    subtractDays(28),
    16,
    [1, 2], // Away win
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'away' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'draw' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'away' },
    ]
  ),
  'match-16-2': createPastMatch(
    'match-16-2',
    'Olympique Marseille',
    'Paris Saint Germain',
    subtractDays(28),
    16,
    [1, 1], // Draw - Le Classique
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'draw' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'away' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'home' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'draw' },
    ]
  ),
  'match-16-3': createPastMatch(
    'match-16-3',
    'Rennes',
    'LOSC Lille',
    subtractDays(28),
    16,
    [2, 1], // Home win
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'away' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'home' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'away' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'home' },
    ]
  ),
  'match-16-4': createPastMatch(
    'match-16-4',
    'Lens',
    'Nice',
    subtractDays(28),
    16,
    [3, 2], // Home win - high scoring
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'home' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'draw' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'away' },
    ]
  ),
  'match-16-5': createPastMatch(
    'match-16-5',
    'Toulouse',
    'Nantes',
    subtractDays(27),
    16,
    [0, 1], // Away win
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'away' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'home' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'draw' },
    ]
  ),
  'match-16-6': createPastMatch(
    'match-16-6',
    'Strasbourg',
    'Metz',
    subtractDays(27),
    16,
    [2, 0], // Home win
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'draw' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'home' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'away' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'home' },
    ]
  ),
  'match-16-7': createPastMatch(
    'match-16-7',
    'Auxerre',
    'Le Havre',
    subtractDays(27),
    16,
    [1, 2], // Away win
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'away' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'draw' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'home' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'away' },
    ]
  ),
  'match-16-8': createPastMatch(
    'match-16-8',
    'Angers SCO',
    'Brest',
    subtractDays(27),
    16,
    [2, 1], // Home win
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'draw' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'home' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'away' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'home' },
    ]
  ),
  'match-16-9': createPastMatch(
    'match-16-9',
    'Lorient',
    'Paris',
    subtractDays(27),
    16,
    [0, 3], // Away win
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'away' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'draw' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'draw' },
    ]
  ),

  // ========== MATCHDAY 15 (35 days ago) ==========
  // Results: 3 home wins, 4 draws, 2 away wins
  // Points: TestPlayer: 2, Marie: 3, Lucas: 3, Sophie: 3
  'match-15-1': createPastMatch(
    'match-15-1',
    'Paris Saint Germain',
    'LOSC Lille',
    subtractDays(35),
    15,
    [2, 0], // Home win
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'home' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'draw' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'home' },
    ]
  ),
  'match-15-2': createPastMatch(
    'match-15-2',
    'Olympique Lyonnais',
    'Rennes',
    subtractDays(35),
    15,
    [0, 0], // Draw
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'draw' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'draw' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'away' },
    ]
  ),
  'match-15-3': createPastMatch(
    'match-15-3',
    'Monaco',
    'Olympique Marseille',
    subtractDays(35),
    15,
    [1, 1], // Draw
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'away' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'home' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'draw' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'draw' },
    ]
  ),
  'match-15-4': createPastMatch(
    'match-15-4',
    'Nice',
    'Nantes',
    subtractDays(35),
    15,
    [2, 2], // Draw
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'draw' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'away' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'draw' },
    ]
  ),
  'match-15-5': createPastMatch(
    'match-15-5',
    'Lens',
    'Toulouse',
    subtractDays(34),
    15,
    [1, 0], // Home win
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'draw' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'draw' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'home' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'home' },
    ]
  ),
  'match-15-6': createPastMatch(
    'match-15-6',
    'Brest',
    'Auxerre',
    subtractDays(34),
    15,
    [0, 1], // Away win
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'home' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'away' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'draw' },
    ]
  ),
  'match-15-7': createPastMatch(
    'match-15-7',
    'Strasbourg',
    'Angers SCO',
    subtractDays(34),
    15,
    [3, 1], // Home win
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'draw' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'home' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'home' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'away' },
    ]
  ),
  'match-15-8': createPastMatch(
    'match-15-8',
    'Metz',
    'Le Havre',
    subtractDays(34),
    15,
    [1, 1], // Draw
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'away' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'draw' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'draw' },
    ]
  ),
  'match-15-9': createPastMatch(
    'match-15-9',
    'Paris',
    'Lorient',
    subtractDays(34),
    15,
    [1, 2], // Away win
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'away' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'away' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'home' },
    ]
  ),
};

export const MOCK_MATCHES_GAME_1: MatchesResponse = {
  incomingMatches: incomingMatchesGame1,
  pastMatches: pastMatchesGame1,
};

// ============================================================================
// MOCK MATCHES - GAME 2 (Family League) - 3 PLAYERS
// ============================================================================

const incomingMatchesGame2: Record<string, MatchData> = {
  'match-g2-20-1': createIncomingMatch(
    'match-g2-20-1',
    'Olympique Marseille',
    'Paris Saint Germain',
    addDays(3),
    20,
    [3.5, 3.4, 2.0]
  ),
  'match-g2-20-2': createIncomingMatch(
    'match-g2-20-2',
    'Nice',
    'Monaco',
    addDays(3),
    20,
    [2.4, 3.3, 2.8],
    'home'
  ),
  'match-g2-20-3': createIncomingMatch(
    'match-g2-20-3',
    'Lens',
    'Olympique Lyonnais',
    addDays(3),
    20,
    [2.2, 3.3, 3.0]
  ),
};

const pastMatchesGame2: Record<string, MatchData> = {
  // Matchday 19
  'match-g2-19-1': createPastMatch(
    'match-g2-19-1',
    'Lens',
    'LOSC Lille',
    subtractDays(5),
    19,
    [1, 2], // Away win - Derby du Nord
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'away' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'home' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'away' },
    ]
  ),
  'match-g2-19-2': createPastMatch(
    'match-g2-19-2',
    'Paris Saint Germain',
    'Monaco',
    subtractDays(5),
    19,
    [3, 0], // Home win
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'home' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'draw' },
    ]
  ),
  'match-g2-19-3': createPastMatch(
    'match-g2-19-3',
    'Olympique Lyonnais',
    'Olympique Marseille',
    subtractDays(5),
    19,
    [2, 2], // Draw - Olympico
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'draw' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'draw' },
    ]
  ),
  // Matchday 18
  'match-g2-18-1': createPastMatch(
    'match-g2-18-1',
    'Nice',
    'Paris Saint Germain',
    subtractDays(12),
    18,
    [0, 2], // Away win
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'away' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'away' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'home' },
    ]
  ),
  'match-g2-18-2': createPastMatch(
    'match-g2-18-2',
    'Olympique Marseille',
    'LOSC Lille',
    subtractDays(12),
    18,
    [1, 1], // Draw
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'draw' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'draw' },
    ]
  ),
  'match-g2-18-3': createPastMatch(
    'match-g2-18-3',
    'Monaco',
    'Lens',
    subtractDays(12),
    18,
    [2, 1], // Home win
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'draw' },
      { playerId: 'mock-player-2', playerName: 'Marie', prediction: 'home' },
      { playerId: 'mock-player-4', playerName: 'Sophie', prediction: 'home' },
    ]
  ),
};

export const MOCK_MATCHES_GAME_2: MatchesResponse = {
  incomingMatches: incomingMatchesGame2,
  pastMatches: pastMatchesGame2,
};

// ============================================================================
// MOCK MATCHES - GAME 3 (Work Buddies) - 2 PLAYERS
// ============================================================================

const incomingMatchesGame3: Record<string, MatchData> = {
  'match-g3-20-1': createIncomingMatch(
    'match-g3-20-1',
    'Brest',
    'Angers SCO',
    addDays(1),
    20,
    [2.0, 3.3, 3.6],
    'home'
  ),
  'match-g3-20-2': createIncomingMatch(
    'match-g3-20-2',
    'Rennes',
    'Nantes',
    addDays(1),
    20,
    [1.9, 3.4, 3.8]
  ),
};

const pastMatchesGame3: Record<string, MatchData> = {
  // Matchday 19
  'match-g3-19-1': createPastMatch(
    'match-g3-19-1',
    'Metz',
    'Le Havre',
    subtractDays(4),
    19,
    [0, 0], // Draw
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'draw' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'home' },
    ]
  ),
  'match-g3-19-2': createPastMatch(
    'match-g3-19-2',
    'Auxerre',
    'Strasbourg',
    subtractDays(4),
    19,
    [2, 1], // Home win
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'away' },
    ]
  ),
  // Matchday 18
  'match-g3-18-1': createPastMatch(
    'match-g3-18-1',
    'Toulouse',
    'Brest',
    subtractDays(11),
    18,
    [1, 3], // Away win
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'away' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'home' },
    ]
  ),
  'match-g3-18-2': createPastMatch(
    'match-g3-18-2',
    'Angers SCO',
    'Nantes',
    subtractDays(11),
    18,
    [0, 0], // Draw
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'draw' },
    ]
  ),
  // Matchday 17
  'match-g3-17-1': createPastMatch(
    'match-g3-17-1',
    'Le Havre',
    'Auxerre',
    subtractDays(18),
    17,
    [1, 2], // Away win
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'draw' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'away' },
    ]
  ),
  'match-g3-17-2': createPastMatch(
    'match-g3-17-2',
    'Strasbourg',
    'Metz',
    subtractDays(18),
    17,
    [3, 3], // Draw - high scoring
    [
      { playerId: 'mock-player-1', playerName: 'TestPlayer', prediction: 'home' },
      { playerId: 'mock-player-3', playerName: 'Lucas', prediction: 'draw' },
    ]
  ),
};

export const MOCK_MATCHES_GAME_3: MatchesResponse = {
  incomingMatches: incomingMatchesGame3,
  pastMatches: pastMatchesGame3,
};

// ============================================================================
// HELPER: Get matches for a game
// ============================================================================

export const getMockMatchesForGame = (gameId: string): MatchesResponse => {
  switch (gameId) {
    case 'game-1':
      return MOCK_MATCHES_GAME_1;
    case 'game-2':
      return MOCK_MATCHES_GAME_2;
    case 'game-3':
      return MOCK_MATCHES_GAME_3;
    default:
      return { incomingMatches: {}, pastMatches: {} };
  }
};

// ============================================================================
// MOCK AUTH TOKEN
// ============================================================================

export const MOCK_AUTH_TOKEN = 'mock-auth-token-abc123';
