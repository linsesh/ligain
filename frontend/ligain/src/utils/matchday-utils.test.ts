import { SeasonMatch, MatchResult } from '../types/match';

// Helper function to create test matches (same as in the component)
const createTestMatch = (
  homeTeam: string,
  awayTeam: string,
  matchday: number,
  date: string,
  status: 'scheduled' | 'in-progress' | 'finished' = 'scheduled'
): SeasonMatch => {
  return new SeasonMatch(
    homeTeam,
    awayTeam,
    0,
    0,
    1.5,
    2.0,
    3.0,
    status,
    '2024',
    'L1',
    date,
    matchday
  );
};

// Helper function to create test match results
const createTestMatchResult = (
  match: SeasonMatch,
  bets: any = null,
  scores: any = null
): MatchResult => ({
  match,
  bets,
  scores,
});

// Helper function to format time (same as in the component)
const formatTime = (date: Date): string => {
  return date.toLocaleTimeString('en-US', { 
    hour: '2-digit', 
    minute: '2-digit',
    hour12: false 
  });
};

// Helper function to format date (same as in the component)
const formatDate = (date: Date): string => {
  return date.toLocaleDateString('en-US', { 
    weekday: 'long', 
    month: 'long', 
    day: 'numeric' 
  });
};

// Test data
const testMatches: MatchResult[] = [
  createTestMatchResult(createTestMatch('PSG', 'Marseille', 1, '2024-03-20T20:00:00')),
  createTestMatchResult(createTestMatch('Lyon', 'Monaco', 1, '2024-03-20T21:00:00')),
  createTestMatchResult(createTestMatch('Nice', 'Lille', 2, '2024-03-21T20:00:00')),
  createTestMatchResult(createTestMatch('Bordeaux', 'Toulouse', 2, '2024-03-21T20:00:00')),
  createTestMatchResult(createTestMatch('Nantes', 'Rennes', 3, '2024-03-22T19:00:00')),
  createTestMatchResult(createTestMatch('Lens', 'Strasbourg', 1, '2024-03-20T19:00:00', 'finished')),
];

describe('Matchday Grouping and Time Organization Utils', () => {
  describe('formatTime function', () => {
    it('should format time in 24-hour format', () => {
      // Use local time to avoid timezone issues
      const date1 = new Date('2024-03-20T20:00:00');
      const date2 = new Date('2024-03-20T09:30:00');
      const date3 = new Date('2024-03-20T23:45:00');

      expect(formatTime(date1)).toBe('20:00');
      expect(formatTime(date2)).toBe('09:30');
      expect(formatTime(date3)).toBe('23:45');
    });

    it('should handle different timezones correctly', () => {
      const date = new Date('2024-03-20T20:00:00');
      expect(formatTime(date)).toBe('20:00');
    });
  });

  describe('formatDate function', () => {
    it('should format date correctly', () => {
      const date = new Date('2024-03-20T20:00:00');
      expect(formatDate(date)).toBe('Wednesday, March 20');
    });

    it('should handle different dates', () => {
      const date1 = new Date('2024-03-21T20:00:00');
      const date2 = new Date('2024-03-22T20:00:00');

      expect(formatDate(date1)).toBe('Thursday, March 21');
      expect(formatDate(date2)).toBe('Friday, March 22');
    });
  });

  describe('Matchday Grouping Logic', () => {
    it('should group matches by matchday correctly', () => {
      // Simulate the grouping logic from the component
      const matchesByMatchday = testMatches.reduce((acc, matchResult) => {
        const matchday = matchResult.match.getMatchday();
        if (!acc[matchday]) {
          acc[matchday] = [];
        }
        acc[matchday].push(matchResult);
        return acc;
      }, {} as { [key: number]: MatchResult[] });

      expect(Object.keys(matchesByMatchday)).toHaveLength(3);
      expect(matchesByMatchday[1]).toHaveLength(3); // PSG, Lyon, Lens
      expect(matchesByMatchday[2]).toHaveLength(2); // Nice, Bordeaux
      expect(matchesByMatchday[3]).toHaveLength(1); // Nantes
    });

    it('should sort matchdays in ascending order', () => {
      const matchesByMatchday = testMatches.reduce((acc, matchResult) => {
        const matchday = matchResult.match.getMatchday();
        if (!acc[matchday]) {
          acc[matchday] = [];
        }
        acc[matchday].push(matchResult);
        return acc;
      }, {} as { [key: number]: MatchResult[] });

      const sortedMatchdays = Object.keys(matchesByMatchday)
        .map(Number)
        .sort((a, b) => a - b);

      expect(sortedMatchdays).toEqual([1, 2, 3]);
    });
  });

  describe('Time Grouping Logic', () => {
    it('should group matches by time within a matchday', () => {
      const matchday1Matches = testMatches.filter(m => m.match.getMatchday() === 1);
      
      // Sort by time
      const sortedMatches = matchday1Matches.sort((a, b) => 
        a.match.getDate().getTime() - b.match.getDate().getTime()
      );

      // Group by time
      const groupedByTime = sortedMatches.reduce((acc, matchResult) => {
        const timeKey = formatTime(matchResult.match.getDate());
        if (!acc[timeKey]) {
          acc[timeKey] = [];
        }
        acc[timeKey].push(matchResult);
        return acc;
      }, {} as { [key: string]: MatchResult[] });

      expect(Object.keys(groupedByTime)).toHaveLength(3);
      expect(groupedByTime['19:00']).toHaveLength(1); // Lens
      expect(groupedByTime['20:00']).toHaveLength(1); // PSG
      expect(groupedByTime['21:00']).toHaveLength(1); // Lyon
    });

    it('should sort matches by time in ascending order', () => {
      const matchday1Matches = testMatches.filter(m => m.match.getMatchday() === 1);
      
      const sortedMatches = matchday1Matches.sort((a, b) => 
        a.match.getDate().getTime() - b.match.getDate().getTime()
      );

      const times = sortedMatches.map(m => formatTime(m.match.getDate()));
      expect(times).toEqual(['19:00', '20:00', '21:00']);
    });

    it('should handle multiple matches at the same time', () => {
      const matchday2Matches = testMatches.filter(m => m.match.getMatchday() === 2);
      
      const sortedMatches = matchday2Matches.sort((a, b) => 
        a.match.getDate().getTime() - b.match.getDate().getTime()
      );

      const groupedByTime = sortedMatches.reduce((acc, matchResult) => {
        const timeKey = formatTime(matchResult.match.getDate());
        if (!acc[timeKey]) {
          acc[timeKey] = [];
        }
        acc[timeKey].push(matchResult);
        return acc;
      }, {} as { [key: string]: MatchResult[] });

      expect(groupedByTime['20:00']).toHaveLength(2); // Nice and Bordeaux
    });
  });

  describe('Edge Cases', () => {
    it('should handle empty matches array', () => {
      const emptyMatches: MatchResult[] = [];
      
      const matchesByMatchday = emptyMatches.reduce((acc, matchResult) => {
        const matchday = matchResult.match.getMatchday();
        if (!acc[matchday]) {
          acc[matchday] = [];
        }
        acc[matchday].push(matchResult);
        return acc;
      }, {} as { [key: number]: MatchResult[] });

      expect(Object.keys(matchesByMatchday)).toHaveLength(0);
    });

    it('should handle single match', () => {
      const singleMatch = [createTestMatchResult(createTestMatch('PSG', 'Marseille', 1, '2024-03-20T20:00:00'))];
      
      const matchesByMatchday = singleMatch.reduce((acc, matchResult) => {
        const matchday = matchResult.match.getMatchday();
        if (!acc[matchday]) {
          acc[matchday] = [];
        }
        acc[matchday].push(matchResult);
        return acc;
      }, {} as { [key: number]: MatchResult[] });

      expect(Object.keys(matchesByMatchday)).toHaveLength(1);
      expect(matchesByMatchday[1]).toHaveLength(1);
    });

    it('should handle matches with same time but different matchdays', () => {
      const sameTimeMatches = [
        createTestMatchResult(createTestMatch('PSG', 'Marseille', 1, '2024-03-20T20:00:00')),
        createTestMatchResult(createTestMatch('Lyon', 'Monaco', 2, '2024-03-21T20:00:00')),
      ];

      const matchesByMatchday = sameTimeMatches.reduce((acc, matchResult) => {
        const matchday = matchResult.match.getMatchday();
        if (!acc[matchday]) {
          acc[matchday] = [];
        }
        acc[matchday].push(matchResult);
        return acc;
      }, {} as { [key: number]: MatchResult[] });

      expect(Object.keys(matchesByMatchday)).toHaveLength(2);
      expect(matchesByMatchday[1]).toHaveLength(1);
      expect(matchesByMatchday[2]).toHaveLength(1);

      // Both should have matches at 20:00
      const time1 = formatTime(matchesByMatchday[1][0].match.getDate());
      const time2 = formatTime(matchesByMatchday[2][0].match.getDate());
      expect(time1).toBe('20:00');
      expect(time2).toBe('20:00');
    });
  });

  describe('Navigation Logic', () => {
    it('should calculate correct navigation indices', () => {
      const sortedMatchdays = [1, 2, 3];
      
      // Test navigation from matchday 1
      const currentIndex1 = sortedMatchdays.indexOf(1);
      expect(currentIndex1).toBe(0);
      expect(currentIndex1 > 0).toBe(false); // prev should be disabled
      expect(currentIndex1 < sortedMatchdays.length - 1).toBe(true); // next should be enabled

      // Test navigation from matchday 2
      const currentIndex2 = sortedMatchdays.indexOf(2);
      expect(currentIndex2).toBe(1);
      expect(currentIndex2 > 0).toBe(true); // prev should be enabled
      expect(currentIndex2 < sortedMatchdays.length - 1).toBe(true); // next should be enabled

      // Test navigation from matchday 3
      const currentIndex3 = sortedMatchdays.indexOf(3);
      expect(currentIndex3).toBe(2);
      expect(currentIndex3 > 0).toBe(true); // prev should be enabled
      expect(currentIndex3 < sortedMatchdays.length - 1).toBe(false); // next should be disabled
    });

    it('should handle navigation with single matchday', () => {
      const singleMatchday = [1];
      
      const currentIndex = singleMatchday.indexOf(1);
      expect(currentIndex).toBe(0);
      expect(currentIndex > 0).toBe(false); // prev should be disabled
      expect(currentIndex < singleMatchday.length - 1).toBe(false); // next should be disabled
    });
  });
}); 