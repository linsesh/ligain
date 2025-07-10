import { SeasonMatch } from './match';

describe('Match', () => {
    const createMatch = (
        homeTeam: string,
        awayTeam: string,
        homeGoals: number,
        awayGoals: number
    ) => {
        return new SeasonMatch(
            homeTeam,
            awayTeam,
            homeGoals,
            awayGoals,
            1.0,
            2.0,
            3.0,
            'finished',
            '2024',
            'Premier League',
            new Date('2024-01-01T15:00:00Z'),
            1
        );
    };

    test('home team wins', () => {
        const match = createMatch('Manchester United', 'Liverpool', 3, 1);
        expect(match.getWinner()).toBe('Manchester United');
    });

    test('away team wins', () => {
        const match = createMatch('Arsenal', 'Chelsea', 0, 2);
        expect(match.getWinner()).toBe('Chelsea');
    });

    test('draw', () => {
        const match = createMatch('Tottenham', 'West Ham', 1, 1);
        expect(match.getWinner()).toBe('Draw');
    });

    test('absolute goal difference', () => {
        const match = createMatch('Manchester United', 'Liverpool', 3, 1);
        expect(match.absoluteGoalDifference()).toBe(2);
    });

    test('is draw', () => {
        const match = createMatch('Tottenham', 'West Ham', 1, 1);
        expect(match.isDraw()).toBe(true);
    });

    test('total goals', () => {
        const match = createMatch('Manchester United', 'Liverpool', 3, 1);
        expect(match.totalGoals()).toBe(4);
    });

    test('absolute difference odds between home and away', () => {
        const match = createMatch('Manchester United', 'Liverpool', 3, 1);
        expect(match.absoluteDifferenceOddsBetweenHomeAndAway()).toBe(1.0);
    });

    test('is finished', () => {
        const match = createMatch('Manchester United', 'Liverpool', 3, 1);
        expect(match.isFinished()).toBe(true);
    });

    test('finish match', () => {
        const match = new SeasonMatch(
            'Manchester United',
            'Liverpool',
            0,
            0,
            1.0,
            2.0,
            3.0,
            'scheduled',
            '2024',
            'Premier League',
            new Date('2024-01-01T15:00:00Z'),
            1
        );
        match.finish(3, 1);
        expect(match.getHomeGoals()).toBe(3);
        expect(match.getAwayGoals()).toBe(1);
        expect(match.isFinished()).toBe(true);
    });
}); 