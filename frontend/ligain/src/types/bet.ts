import { SeasonMatch, SimplifiedBet } from './match';

export interface Bet extends SimplifiedBet {
    match: SeasonMatch;
    isBetCorrect(): boolean;
    isBetPerfect(): boolean;
    absoluteDifferenceGoalDifferenceWithMatch(): number;
    isGoalDifferenceTheSameAsMatch(): boolean;
    absoluteGoalDifference(): number;
    totalPredictedGoals(): number;
    absoluteDifferenceTotalGoalsWithMatch(): number;
    getPredictedResult(): string;
}

export class BetImpl implements Bet {
    constructor(
        public match: SeasonMatch,
        public predictedHomeGoals: number,
        public predictedAwayGoals: number
    ) {}

    isBetCorrect(): boolean {
        if (this.match.getHomeGoals() > this.match.getAwayGoals()) {
            return this.predictedHomeGoals > this.predictedAwayGoals;
        }
        if (this.match.getHomeGoals() < this.match.getAwayGoals()) {
            return this.predictedHomeGoals < this.predictedAwayGoals;
        }
        return this.predictedHomeGoals === this.predictedAwayGoals;
    }

    isBetPerfect(): boolean {
        return this.predictedHomeGoals === this.match.getHomeGoals() && 
               this.predictedAwayGoals === this.match.getAwayGoals();
    }

    absoluteDifferenceGoalDifferenceWithMatch(): number {
        return Math.abs(this.match.absoluteGoalDifference() - this.absoluteGoalDifference());
    }

    isGoalDifferenceTheSameAsMatch(): boolean {
        return this.absoluteDifferenceGoalDifferenceWithMatch() === 0;
    }

    absoluteGoalDifference(): number {
        return Math.abs(this.predictedHomeGoals - this.predictedAwayGoals);
    }

    totalPredictedGoals(): number {
        return this.predictedHomeGoals + this.predictedAwayGoals;
    }

    absoluteDifferenceTotalGoalsWithMatch(): number {
        return Math.abs(this.totalPredictedGoals() - this.match.totalGoals());
    }

    getPredictedResult(): string {
        if (this.predictedHomeGoals > this.predictedAwayGoals) {
            return this.match.getHomeTeam();
        }
        if (this.predictedHomeGoals < this.predictedAwayGoals) {
            return this.match.getAwayTeam();
        }
        return 'Draw';
    }

    isModifiable(now: Date): boolean {
        return !this.match.isFinished() && !this.match.isInProgress() && now < this.match.getDate();
    }

    static fromJSON(json: any, match: SeasonMatch): BetImpl {
        return new BetImpl(
            match,
            json.predictedHomeGoals,
            json.predictedAwayGoals
        );
    }

    toJSON(): any {
        return {
            matchId: this.match.id(),
            predictedHomeGoals: this.predictedHomeGoals,
            predictedAwayGoals: this.predictedAwayGoals
        };
    }
} 