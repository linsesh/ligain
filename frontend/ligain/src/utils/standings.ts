import { MatchResult } from '../types/match';

export interface TeamStanding {
  teamName: string; // raw name, suitable for logo lookup
  played: number;
  won: number;
  drawn: number;
  lost: number;
  goalsFor: number;
  goalsAgainst: number;
  goalDifference: number;
  points: number; // 3 per win, 1 per draw
  position: number; // 1-based rank
}

interface TeamRecord {
  played: number;
  won: number;
  drawn: number;
  lost: number;
  goalsFor: number;
  goalsAgainst: number;
}

// Computes the league table from a set of matches.
// Only finished matches are included; non-finished matches are ignored.
// Sort order: points desc → goal difference desc → goals for desc → teamName asc.
export function computeLeagueStandings(
  matches: Record<string, MatchResult>
): TeamStanding[] {
  const teams: Record<string, TeamRecord> = {};

  const ensureTeam = (name: string) => {
    if (!teams[name]) {
      teams[name] = { played: 0, won: 0, drawn: 0, lost: 0, goalsFor: 0, goalsAgainst: 0 };
    }
  };

  for (const { match } of Object.values(matches)) {
    if (!match.isFinished()) continue;

    const home = match.getHomeTeamForLogo();
    const away = match.getAwayTeamForLogo();
    const hg = match.getHomeGoals();
    const ag = match.getAwayGoals();

    ensureTeam(home);
    ensureTeam(away);

    teams[home].played++;
    teams[away].played++;
    teams[home].goalsFor += hg;
    teams[home].goalsAgainst += ag;
    teams[away].goalsFor += ag;
    teams[away].goalsAgainst += hg;

    if (hg > ag) {
      teams[home].won++;
      teams[away].lost++;
    } else if (hg < ag) {
      teams[away].won++;
      teams[home].lost++;
    } else {
      teams[home].drawn++;
      teams[away].drawn++;
    }
  }

  const sorted = Object.entries(teams)
    .map(([teamName, record]) => {
      const points = record.won * 3 + record.drawn;
      const goalDifference = record.goalsFor - record.goalsAgainst;
      return { teamName, ...record, goalDifference, points };
    })
    .sort((a, b) => {
      if (b.points !== a.points) return b.points - a.points;
      if (b.goalDifference !== a.goalDifference) return b.goalDifference - a.goalDifference;
      if (b.goalsFor !== a.goalsFor) return b.goalsFor - a.goalsFor;
      return a.teamName.localeCompare(b.teamName);
    });

  return sorted.map((entry, idx) => ({ ...entry, position: idx + 1 }));
}

const MATCH_WINDOW_MS = 4 * 60 * 60 * 1000; // 4 hours

// Determines the "current" matchday for fingerprinting:
//   1. If any matchday has a match that started within the last 4h, return the one
//      whose most recently started match has the latest start time.
//   2. Otherwise, return the matchday of the closest upcoming match.
//   3. Returns undefined when there are no matches or all are in the past beyond 4h.
export function resolveCurrentMatchday(
  matches: Record<string, MatchResult>,
  now: Date
): number | undefined {
  const matchList = Object.values(matches);

  // Collect, per matchday, the latest start time that falls within the 4h window.
  const recentByMatchday: Record<number, Date> = {};

  for (const { match } of matchList) {
    const matchDate = match.getDate();
    const elapsed = now.getTime() - matchDate.getTime();

    if (matchDate <= now && elapsed < MATCH_WINDOW_MS) {
      const md = match.getMatchday();
      if (!recentByMatchday[md] || matchDate > recentByMatchday[md]) {
        recentByMatchday[md] = matchDate;
      }
    }
  }

  // If any matchday qualifies, return the one with the most recently started match.
  if (Object.keys(recentByMatchday).length > 0) {
    let bestMatchday: number | undefined;
    let bestDate: Date | undefined;

    for (const [md, date] of Object.entries(recentByMatchday)) {
      if (!bestDate || date > bestDate) {
        bestDate = date;
        bestMatchday = Number(md);
      }
    }

    return bestMatchday;
  }

  // Fallback: closest future match.
  let closestDate: Date | undefined;
  let closestMatchday: number | undefined;

  for (const { match } of matchList) {
    const matchDate = match.getDate();
    if (matchDate > now) {
      if (!closestDate || matchDate < closestDate) {
        closestDate = matchDate;
        closestMatchday = match.getMatchday();
      }
    }
  }

  return closestMatchday;
}

export type FormResult = 'W' | 'D' | 'L';

// Returns the last `limit` finished match outcomes for a team, ordered oldest→newest
// (suitable for left-to-right display).
export function computeTeamForm(
  teamName: string,
  matches: Record<string, MatchResult>,
  limit = 5
): FormResult[] {
  return Object.values(matches)
    .filter(({ match }) =>
      match.isFinished() &&
      (match.getHomeTeamForLogo() === teamName || match.getAwayTeamForLogo() === teamName)
    )
    .sort((a, b) => a.match.getMatchday() - b.match.getMatchday())
    .slice(-limit)
    .map(({ match }) => {
      if (match.isDraw()) return 'D';
      const isHome = match.getHomeTeamForLogo() === teamName;
      const teamGoals = isHome ? match.getHomeGoals() : match.getAwayGoals();
      const oppGoals  = isHome ? match.getAwayGoals() : match.getHomeGoals();
      return teamGoals > oppGoals ? 'W' : 'L';
    });
}

// Returns a stable string that changes if and only if a new match becomes
// finished in the given matchday.  Pass undefined to get an empty string.
export function computeFingerprint(
  matches: Record<string, MatchResult>,
  currentMatchday: number | undefined
): string {
  if (currentMatchday === undefined) return '';

  return Object.values(matches)
    .filter(({ match }) => match.getMatchday() === currentMatchday && match.isFinished())
    .map(({ match }) => match.id())
    .sort()
    .join(',');
}
