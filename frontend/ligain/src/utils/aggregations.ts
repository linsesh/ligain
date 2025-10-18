export interface SimplifiedScore {
  PlayerID: string;
  PlayerName: string;
  Points: number;
}

export interface PastMatchEntry {
  match: {
    date: string; // ISO date string
    matchday: number;
    status?: string;
    [key: string]: any;
  };
  scores?: Record<string, any> | null;
  [key: string]: any;
}

export type AggregatedScore = SimplifiedScore;

export interface AggregationResult {
  perMonthLeaderboard: Record<string, AggregatedScore[]>;
  perMatchdayLeaderboard: Record<number, AggregatedScore[]>;
}

function parseMatchDate(value: any): Date | null {
  try {
    const d = new Date(value);
    return Number.isNaN(d.getTime()) ? null : d;
  } catch {
    return null;
  }
}

function getMonthKeyUTC(date: Date): string {
  const year = date.getUTCFullYear();
  const month = String(date.getUTCMonth() + 1).padStart(2, '0');
  return `${year}-${month}`;
}

function upsertAggregate(
  bucket: Record<string, AggregatedScore>,
  playerId: string,
  playerName: string,
  pointsToAdd: number
) {
  if (!bucket[playerId]) {
    bucket[playerId] = { PlayerID: playerId, PlayerName: playerName, Points: 0 };
  }
  bucket[playerId].Points += pointsToAdd;
}

// Accept both API shapes: { PlayerID, PlayerName, Points } or { playerId, playerName, points }
function normalizeScoreShape(raw: any): { playerId?: string; playerName?: string; points?: number } {
  if (!raw || typeof raw !== 'object') return {};
  const playerId = raw.PlayerID ?? raw.playerId;
  const playerName = raw.PlayerName ?? raw.playerName;
  const points = raw.Points ?? raw.points;
  return { playerId, playerName, points };
}

function ensureBuckets(
  perMonth: Record<string, Record<string, AggregatedScore>>,
  perMatchday: Record<number, Record<string, AggregatedScore>>,
  monthKey: string,
  matchday: number
) {
  if (!perMonth[monthKey]) perMonth[monthKey] = {};
  if (!perMatchday[matchday]) perMatchday[matchday] = {};
}

// Compute leaderboards per month (YYYY-MM) and per matchday from pastMatches.
// Input is the JSON shape returned by the backend: an object keyed by matchId.
export function computeMonthlyAndMatchdayScores(
  pastMatches: Record<string, PastMatchEntry> | undefined | null
): AggregationResult {
  const perMonth: Record<string, Record<string, AggregatedScore>> = {};
  const perMatchday: Record<number, Record<string, AggregatedScore>> = {};

  if (!pastMatches) {
    return {
      perMonthLeaderboard: {},
      perMatchdayLeaderboard: {},
    };
  }

  for (const [, entry] of Object.entries(pastMatches)) {
    if (!entry || !entry.match || !entry.scores) continue;
    const matchDate = parseMatchDate(entry.match.date);
    if (!matchDate) continue;

    const monthKey = getMonthKeyUTC(matchDate);
    const matchday = Number(entry.match.matchday);

    for (const [, score] of Object.entries(entry.scores)) {
      const { playerId, playerName, points } = normalizeScoreShape(score);
      if (!playerId) continue;
      const delta = Number(points) || 0;

      ensureBuckets(perMonth, perMatchday, monthKey, matchday);
      upsertAggregate(perMonth[monthKey], playerId, playerName || '', delta);
      upsertAggregate(perMatchday[matchday], playerId, playerName || '', delta);
    }
  }

  // Convert to sorted arrays (descending points, then name asc, then id)
  const perMonthLeaderboard: Record<string, AggregatedScore[]> = {};
  for (const [monthKey, playerMap] of Object.entries(perMonth)) {
    perMonthLeaderboard[monthKey] = Object.values(playerMap).sort((a, b) => {
      if (b.Points !== a.Points) return b.Points - a.Points;
      if (a.PlayerName && b.PlayerName) return a.PlayerName.localeCompare(b.PlayerName);
      return a.PlayerID.localeCompare(b.PlayerID);
    });
  }

  const perMatchdayLeaderboard: Record<number, AggregatedScore[]> = {};
  for (const [matchdayKey, playerMap] of Object.entries(perMatchday)) {
    const md = Number(matchdayKey);
    perMatchdayLeaderboard[md] = Object.values(playerMap).sort((a, b) => {
      if (b.Points !== a.Points) return b.Points - a.Points;
      if (a.PlayerName && b.PlayerName) return a.PlayerName.localeCompare(b.PlayerName);
      return a.PlayerID.localeCompare(b.PlayerID);
    });
  }

  return { perMonthLeaderboard, perMatchdayLeaderboard };
}


// Compute total points per player across all matches in a game.
// Returns a single leaderboard (descending points, then name asc, then id).
export function computeTotalScores(
  pastMatches: Record<string, PastMatchEntry> | undefined | null
): AggregatedScore[] {
  if (!pastMatches) return [];

  const totals: Record<string, AggregatedScore> = {};

  for (const [, entry] of Object.entries(pastMatches)) {
    if (!entry || !entry.scores) continue;
    for (const [, score] of Object.entries(entry.scores)) {
      const { playerId, playerName, points } = normalizeScoreShape(score);
      if (!playerId) continue;
      const delta = Number(points) || 0;
      upsertAggregate(totals, playerId, playerName || '', delta);
    }
  }

  return Object.values(totals).sort((a, b) => {
    if (b.Points !== a.Points) return b.Points - a.Points;
    if (a.PlayerName && b.PlayerName) return a.PlayerName.localeCompare(b.PlayerName);
    return a.PlayerID.localeCompare(b.PlayerID);
  });
}


// Build cumulative points per matchday for each player from an already aggregated perMatchday leaderboard map.
// Returns sorted unique matchdays and a series per player with cumulative values aligned to matchdays.
export function computeCumulativePointsByMatchday(
  perMatchdayLeaderboard: Record<number, AggregatedScore[] | undefined | null>
): { matchdays: number[]; series: { playerId: string; playerName: string; values: number[] }[] } {
  const mdKeys = Object.keys(perMatchdayLeaderboard || {})
    .map((k) => Number(k))
    .filter((n) => !Number.isNaN(n))
    .sort((a, b) => a - b);
  if (mdKeys.length === 0) return { matchdays: [], series: [] };

  const playerIdToName: Record<string, string> = {};
  for (const md of mdKeys) {
    for (const score of (perMatchdayLeaderboard[md] || [])) {
      playerIdToName[score.PlayerID] = score.PlayerName;
    }
  }
  const playerIds = Object.keys(playerIdToName).sort();

  const runningTotals: Record<string, number> = {};
  const valuesByPlayer: Record<string, number[]> = {};
  for (const pid of playerIds) {
    runningTotals[pid] = 0;
    valuesByPlayer[pid] = [];
  }

  for (const md of mdKeys) {
    const deltaByPlayer: Record<string, number> = {};
    for (const s of (perMatchdayLeaderboard[md] || [])) {
      deltaByPlayer[s.PlayerID] = s.Points;
    }
    for (const pid of playerIds) {
      runningTotals[pid] += deltaByPlayer[pid] ?? 0;
      valuesByPlayer[pid].push(runningTotals[pid]);
    }
  }

  const series = playerIds.map((pid) => ({
    playerId: pid,
    playerName: playerIdToName[pid] || '',
    values: valuesByPlayer[pid],
  }));

  return { matchdays: mdKeys, series };
}


