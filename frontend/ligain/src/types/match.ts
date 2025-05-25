import { Bet } from './bet';

export type MatchStatus = 'scheduled' | 'finished';

export class SeasonMatch {
    private homeTeam: string;
    private awayTeam: string;
    private homeGoals: number;
    private awayGoals: number;
    private homeTeamOdds: number;
    private awayTeamOdds: number;
    private drawOdds: number;
    private status: MatchStatus;
    private seasonCode: string;
    private competitionCode: string;
    private date: Date;
    private matchday: number;

    constructor(
        homeTeam: string,
        awayTeam: string,
        homeGoals: number,
        awayGoals: number,
        homeTeamOdds: number,
        awayTeamOdds: number,
        drawOdds: number,
        status: MatchStatus,
        seasonCode: string,
        competitionCode: string,
        date: string | Date,
        matchday: number
    ) {
        this.homeTeam = homeTeam;
        this.awayTeam = awayTeam;
        this.homeGoals = homeGoals;
        this.awayGoals = awayGoals;
        this.homeTeamOdds = homeTeamOdds;
        this.awayTeamOdds = awayTeamOdds;
        this.drawOdds = drawOdds;
        this.status = status;
        this.seasonCode = seasonCode;
        this.competitionCode = competitionCode;
        this.date = typeof date === 'string' ? new Date(date) : date;
        this.matchday = matchday;
    }

    id(): string {
        return `${this.competitionCode}-${this.seasonCode}-${this.homeTeam}-${this.awayTeam}-${this.matchday}`;
    }

    getHomeTeam(): string {
        return this.homeTeam;
    }

    getAwayTeam(): string {
        return this.awayTeam;
    }

    getHomeGoals(): number {
        return this.homeGoals;
    }

    getAwayGoals(): number {
        return this.awayGoals;
    }

    getHomeTeamOdds(): number {
        return this.homeTeamOdds;
    }

    getAwayTeamOdds(): number {
        return this.awayTeamOdds;
    }

    getDrawOdds(): number {
        return this.drawOdds;
    }

    getDate(): Date {
        return this.date;
    }

    isFinished(): boolean {
        return this.status === 'finished';
    }

    getWinner(): string {
        if (this.homeGoals > this.awayGoals) {
            return this.homeTeam;
        }
        if (this.awayGoals > this.homeGoals) {
            return this.awayTeam;
        }
        return 'Draw';
    }

    absoluteGoalDifference(): number {
        return Math.abs(this.homeGoals - this.awayGoals);
    }

    isDraw(): boolean {
        return this.homeGoals === this.awayGoals;
    }

    totalGoals(): number {
        return this.homeGoals + this.awayGoals;
    }

    absoluteDifferenceOddsBetweenHomeAndAway(): number {
        return Math.abs(this.homeTeamOdds - this.awayTeamOdds);
    }

    finish(homeGoals: number, awayGoals: number): void {
        this.homeGoals = homeGoals;
        this.awayGoals = awayGoals;
        this.status = 'finished';
    }

    getSeasonCode(): string {
        return this.seasonCode;
    }

    getCompetitionCode(): string {
        return this.competitionCode;
    }

    getMatchday(): number {
        return this.matchday;
    }

    static fromJSON(json: any): SeasonMatch {
        return new SeasonMatch(
            json.homeTeam,
            json.awayTeam,
            json.homeGoals,
            json.awayGoals,
            json.homeTeamOdds,
            json.awayTeamOdds,
            json.drawOdds,
            json.status,
            json.seasonCode,
            json.competitionCode,
            json.date,
            json.matchday
        );
    }
}

export interface SimplifiedBet {
    predictedHomeGoals: number;
    predictedAwayGoals: number;
}

export interface MatchResult {
    match: SeasonMatch;
    bets: { [key: string]: SimplifiedBet } | null;
    scores: { [key: string]: number } | null;
}

export interface MatchesResponse {
    incomingMatches: { [key: string]: MatchResult };
    pastMatches: { [key: string]: MatchResult };
} 