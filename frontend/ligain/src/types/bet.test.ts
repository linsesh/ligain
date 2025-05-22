import { SeasonMatch } from './match';
import { BetImpl } from './bet';

describe('Bet', () => {
    const createMatch = (
        homeTeam: string,
        awayTeam: string,
        homeGoals: number,
        awayGoals: number
    ) => {
        return new SeasonMatch(
            homeTeam,
            awayTeam,
            '2024',
            'Premier League',
            new Date('2024-01-01T15:00:00Z'),
            1,
            1.0,
            2.0,
            3.0,
            homeGoals,
            awayGoals,
            'finished'
        );
    };

    test('home team wins - correct prediction', () => {
        const match = createMatch('Manchester United', 'Liverpool', 3, 1);
        const bet = new BetImpl(match, 1, 0);
        expect(bet.isBetCorrect()).toBe(true);
    });

    test('away team wins - correct prediction', () => {
        const match = createMatch('Arsenal', 'Chelsea', 0, 2);
        const bet = new BetImpl(match, 0, 2);
        expect(bet.isBetCorrect()).toBe(true);
    });

    test('draw - correct prediction', () => {
        const match = createMatch('Tottenham', 'West Ham', 1, 1);
        const bet = new BetImpl(match, 0, 0);
        expect(bet.isBetCorrect()).toBe(true);
    });

    test('home team wins but predicted wrong', () => {
        const match = createMatch('Manchester United', 'Liverpool', 3, 1);
        const bet = new BetImpl(match, 0, 2);
        expect(bet.isBetCorrect()).toBe(false);
    });

    test('away team wins but predicted wrong', () => {
        const match = createMatch('Arsenal', 'Chelsea', 0, 2);
        const bet = new BetImpl(match, 2, 0);
        expect(bet.isBetCorrect()).toBe(false);
    });

    test('draw but predicted wrong', () => {
        const match = createMatch('Tottenham', 'West Ham', 1, 1);
        const bet = new BetImpl(match, 2, 0);
        expect(bet.isBetCorrect()).toBe(false);
    });

    test('perfect prediction', () => {
        const match = createMatch('Manchester United', 'Liverpool', 3, 1);
        const bet = new BetImpl(match, 3, 1);
        expect(bet.isBetPerfect()).toBe(true);
    });

    test('not perfect prediction', () => {
        const match = createMatch('Manchester United', 'Liverpool', 3, 1);
        const bet = new BetImpl(match, 2, 1);
        expect(bet.isBetPerfect()).toBe(false);
    });

    test('goal difference same as match', () => {
        const match = createMatch('Manchester United', 'Liverpool', 3, 1);
        const bet = new BetImpl(match, 2, 0);
        expect(bet.isGoalDifferenceTheSameAsMatch()).toBe(true);
    });

    test('goal difference different from match', () => {
        const match = createMatch('Manchester United', 'Liverpool', 3, 1);
        const bet = new BetImpl(match, 1, 0);
        expect(bet.isGoalDifferenceTheSameAsMatch()).toBe(false);
    });

    test('total predicted goals', () => {
        const match = createMatch('Manchester United', 'Liverpool', 3, 1);
        const bet = new BetImpl(match, 2, 1);
        expect(bet.totalPredictedGoals()).toBe(3);
    });

    test('absolute difference total goals with match', () => {
        const match = createMatch('Manchester United', 'Liverpool', 3, 1);
        const bet = new BetImpl(match, 2, 1);
        expect(bet.absoluteDifferenceTotalGoalsWithMatch()).toBe(1);
    });

    test('get predicted result - home win', () => {
        const match = createMatch('Manchester United', 'Liverpool', 3, 1);
        const bet = new BetImpl(match, 2, 1);
        expect(bet.getPredictedResult()).toBe('Manchester United');
    });

    test('get predicted result - away win', () => {
        const match = createMatch('Manchester United', 'Liverpool', 3, 1);
        const bet = new BetImpl(match, 1, 2);
        expect(bet.getPredictedResult()).toBe('Liverpool');
    });

    test('get predicted result - draw', () => {
        const match = createMatch('Manchester United', 'Liverpool', 3, 1);
        const bet = new BetImpl(match, 1, 1);
        expect(bet.getPredictedResult()).toBe('Draw');
    });
}); 