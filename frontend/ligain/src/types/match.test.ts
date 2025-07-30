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

    test('get home team odds', () => {
        const match = createMatch('Manchester United', 'Liverpool', 3, 1);
        expect(match.getHomeTeamOdds()).toBe(1.0);
    });

    test('get away team odds', () => {
        const match = createMatch('Manchester United', 'Liverpool', 3, 1);
        expect(match.getAwayTeamOdds()).toBe(2.0);
    });

    test('get draw odds', () => {
        const match = createMatch('Manchester United', 'Liverpool', 3, 1);
        expect(match.getDrawOdds()).toBe(3.0);
    });

    test('odds are correctly set in constructor', () => {
        const match = new SeasonMatch(
            'Manchester United',
            'Liverpool',
            3,
            1,
            1.5,  // home odds
            2.5,  // away odds
            3.5,  // draw odds
            'finished',
            '2024',
            'Premier League',
            new Date('2024-01-01T15:00:00Z'),
            1
        );
        expect(match.getHomeTeamOdds()).toBe(1.5);
        expect(match.getAwayTeamOdds()).toBe(2.5);
        expect(match.getDrawOdds()).toBe(3.5);
    });

    test('fromJSON creates match with correct odds', () => {
        const jsonData = {
            homeTeam: 'Manchester United',
            awayTeam: 'Liverpool',
            homeGoals: 3,
            awayGoals: 1,
            homeTeamOdds: 1.5,
            awayTeamOdds: 2.5,
            drawOdds: 3.5,
            status: 'finished',
            seasonCode: '2024',
            competitionCode: 'Premier League',
            date: '2024-01-01T15:00:00Z',
            matchday: 1
        };
        
        const match = SeasonMatch.fromJSON(jsonData);
        expect(match.getHomeTeamOdds()).toBe(1.5);
        expect(match.getAwayTeamOdds()).toBe(2.5);
        expect(match.getDrawOdds()).toBe(3.5);
    });

    test('has clear favorite when odds difference > 1.5', () => {
        const match = new SeasonMatch(
            'Manchester United',
            'Liverpool',
            3,
            1,
            1.0,  // home odds
            3.0,  // away odds (difference = 2.0 > 1.5)
            2.0,  // draw odds
            'finished',
            '2024',
            'Premier League',
            new Date('2024-01-01T15:00:00Z'),
            1
        );
        expect(match.hasClearFavorite()).toBe(true);
    });

    test('no clear favorite when odds difference <= 1.5', () => {
        const match = new SeasonMatch(
            'Arsenal',
            'Chelsea',
            0,
            2,
            1.5,  // home odds
            2.0,  // away odds (difference = 0.5 <= 1.5)
            3.0,  // draw odds
            'finished',
            '2024',
            'Premier League',
            new Date('2024-01-01T15:00:00Z'),
            1
        );
        expect(match.hasClearFavorite()).toBe(false);
    });

    test('get favorite team - home team favorite', () => {
        const match = new SeasonMatch(
            'Manchester United',
            'Liverpool',
            3,
            1,
            1.0,  // home odds (lower = favorite)
            3.0,  // away odds
            2.0,  // draw odds
            'finished',
            '2024',
            'Premier League',
            new Date('2024-01-01T15:00:00Z'),
            1
        );
        expect(match.getFavoriteTeam()).toBe('Manchester United');
    });

    test('get favorite team - away team favorite', () => {
        const match = new SeasonMatch(
            'Arsenal',
            'Chelsea',
            0,
            2,
            3.0,  // home odds
            1.0,  // away odds (lower = favorite)
            2.0,  // draw odds
            'finished',
            '2024',
            'Premier League',
            new Date('2024-01-01T15:00:00Z'),
            1
        );
        expect(match.getFavoriteTeam()).toBe('Chelsea');
    });

    test('get favorite team - no clear favorite', () => {
        const match = new SeasonMatch(
            'Tottenham',
            'West Ham',
            1,
            1,
            1.5,  // home odds
            2.0,  // away odds (difference = 0.5 <= 1.5)
            3.0,  // draw odds
            'finished',
            '2024',
            'Premier League',
            new Date('2024-01-01T15:00:00Z'),
            1
        );
        expect(match.getFavoriteTeam()).toBe('');
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

    test('transforms Paris to Paris FC for home team', () => {
        const match = new SeasonMatch(
            'Paris',
            'Liverpool',
            3,
            1,
            1.0,
            2.0,
            3.0,
            'finished',
            '2024',
            'Premier League',
            new Date('2024-01-01T15:00:00Z'),
            1
        );
        expect(match.getHomeTeam()).toBe('Paris FC');
    });

    test('transforms Paris to Paris FC for away team', () => {
        const match = new SeasonMatch(
            'Manchester United',
            'Paris',
            3,
            1,
            1.0,
            2.0,
            3.0,
            'finished',
            '2024',
            'Premier League',
            new Date('2024-01-01T15:00:00Z'),
            1
        );
        expect(match.getAwayTeam()).toBe('Paris FC');
    });

    test('does not transform other team names', () => {
        const match = new SeasonMatch(
            'Manchester United',
            'Liverpool',
            3,
            1,
            1.0,
            2.0,
            3.0,
            'finished',
            '2024',
            'Premier League',
            new Date('2024-01-01T15:00:00Z'),
            1
        );
        expect(match.getHomeTeam()).toBe('Manchester United');
        expect(match.getAwayTeam()).toBe('Liverpool');
    });
}); 