import { SeasonMatch } from './match';
import { BetImpl } from './bet';

describe('Bet', () => {
    const createMatch = (
        homeTeam: string,
        awayTeam: string,
        date: Date,
        status: 'scheduled' | 'in-progress' | 'finished' = 'scheduled'
    ) => {
        return new SeasonMatch(
            homeTeam,
            awayTeam,
            0,  // homeGoals
            0,  // awayGoals
            1.0,  // homeTeamOdds
            2.0,  // awayTeamOdds
            3.0,  // drawOdds
            status,
            '2024',  // seasonCode
            'Premier League',  // competitionCode
            date,
            1  // matchday
        );
    };

    test('home team wins - correct prediction', () => {
        const match = createMatch('Manchester United', 'Liverpool', new Date('2024-01-01T15:00:00Z'), 'finished');
        const bet = new BetImpl(match, 1, 0);
        expect(bet.isBetCorrect()).toBe(true);
    });

    test('away team wins - correct prediction', () => {
        const match = createMatch('Arsenal', 'Chelsea', new Date('2024-01-01T15:00:00Z'), 'finished');
        const bet = new BetImpl(match, 0, 2);
        expect(bet.isBetCorrect()).toBe(true);
    });

    test('draw - correct prediction', () => {
        const match = createMatch('Tottenham', 'West Ham', new Date('2024-01-01T15:00:00Z'), 'finished');
        const bet = new BetImpl(match, 0, 0);
        expect(bet.isBetCorrect()).toBe(true);
    });

    test('home team wins but predicted wrong', () => {
        const match = createMatch('Manchester United', 'Liverpool', new Date('2024-01-01T15:00:00Z'), 'finished');
        const bet = new BetImpl(match, 0, 2);
        expect(bet.isBetCorrect()).toBe(false);
    });

    test('away team wins but predicted wrong', () => {
        const match = createMatch('Arsenal', 'Chelsea', new Date('2024-01-01T15:00:00Z'), 'finished');
        const bet = new BetImpl(match, 2, 0);
        expect(bet.isBetCorrect()).toBe(false);
    });

    test('draw but predicted wrong', () => {
        const match = createMatch('Tottenham', 'West Ham', new Date('2024-01-01T15:00:00Z'), 'finished');
        const bet = new BetImpl(match, 2, 0);
        expect(bet.isBetCorrect()).toBe(false);
    });

    test('perfect prediction', () => {
        const match = createMatch('Manchester United', 'Liverpool', new Date('2024-01-01T15:00:00Z'), 'finished');
        const bet = new BetImpl(match, 3, 1);
        expect(bet.isBetPerfect()).toBe(true);
    });

    test('not perfect prediction', () => {
        const match = createMatch('Manchester United', 'Liverpool', new Date('2024-01-01T15:00:00Z'), 'finished');
        const bet = new BetImpl(match, 2, 1);
        expect(bet.isBetPerfect()).toBe(false);
    });

    test('goal difference same as match', () => {
        const match = createMatch('Manchester United', 'Liverpool', new Date('2024-01-01T15:00:00Z'), 'finished');
        const bet = new BetImpl(match, 2, 0);
        expect(bet.isGoalDifferenceTheSameAsMatch()).toBe(true);
    });

    test('goal difference different from match', () => {
        const match = createMatch('Manchester United', 'Liverpool', new Date('2024-01-01T15:00:00Z'), 'finished');
        const bet = new BetImpl(match, 1, 0);
        expect(bet.isGoalDifferenceTheSameAsMatch()).toBe(false);
    });

    test('total predicted goals', () => {
        const match = createMatch('Manchester United', 'Liverpool', new Date('2024-01-01T15:00:00Z'), 'finished');
        const bet = new BetImpl(match, 2, 1);
        expect(bet.totalPredictedGoals()).toBe(3);
    });

    test('absolute difference total goals with match', () => {
        const match = createMatch('Manchester United', 'Liverpool', new Date('2024-01-01T15:00:00Z'), 'finished');
        const bet = new BetImpl(match, 2, 1);
        expect(bet.absoluteDifferenceTotalGoalsWithMatch()).toBe(1);
    });

    test('get predicted result - home win', () => {
        const match = createMatch('Manchester United', 'Liverpool', new Date('2024-01-01T15:00:00Z'), 'finished');
        const bet = new BetImpl(match, 2, 1);
        expect(bet.getPredictedResult()).toBe('Manchester United');
    });

    test('get predicted result - away win', () => {
        const match = createMatch('Manchester United', 'Liverpool', new Date('2024-01-01T15:00:00Z'), 'finished');
        const bet = new BetImpl(match, 1, 2);
        expect(bet.getPredictedResult()).toBe('Liverpool');
    });

    test('get predicted result - draw', () => {
        const match = createMatch('Manchester United', 'Liverpool', new Date('2024-01-01T15:00:00Z'), 'finished');
        const bet = new BetImpl(match, 1, 1);
        expect(bet.getPredictedResult()).toBe('Draw');
    });

    const createBet = (match: SeasonMatch) => {
        return new BetImpl(match, 2, 1);
    };

    describe('isModifiable', () => {
        const referenceTime = new Date('2024-01-01T15:00:00Z');

        it('should be modifiable before match start', () => {
            const match = createMatch('Home', 'Away', new Date('2024-01-01T16:00:00Z')); // 1 hour after reference time
            const bet = createBet(match);
            expect(bet.isModifiable(referenceTime)).toBe(true);
        });

        it('should not be modifiable after match start', () => {
            const match = createMatch('Home', 'Away', new Date('2024-01-01T14:00:00Z')); // 1 hour before reference time
            const bet = createBet(match);
            expect(bet.isModifiable(referenceTime)).toBe(false);
        });

        it('should not be modifiable at match start', () => {
            const match = createMatch('Home', 'Away', referenceTime);
            const bet = createBet(match);
            expect(bet.isModifiable(referenceTime)).toBe(false);
        });

        it('should not be modifiable when match is in progress', () => {
            const match = createMatch('Home', 'Away', referenceTime, 'in-progress');
            const bet = createBet(match);
            expect(bet.isModifiable(referenceTime)).toBe(false);
        });
    });
}); 