export type MatchStatus = 'scheduled' | 'finished';

export interface Match {
    id(): string;
    getSeasonCode(): string;
    getCompetitionCode(): string;
    getDate(): Date;
    getHomeTeam(): string;
    getAwayTeam(): string;
    getHomeGoals(): number;
    getAwayGoals(): number;
    getHomeTeamOdds(): number;
    getAwayTeamOdds(): number;
    getDrawOdds(): number;
    absoluteGoalDifference(): number;
    isDraw(): boolean;
    totalGoals(): number;
    absoluteDifferenceOddsBetweenHomeAndAway(): number;
    isFinished(): boolean;
    finish(homeGoals: number, awayGoals: number): void;
    getWinner(): string;
}

export class SeasonMatch implements Match {
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
        seasonCode: string,
        competitionCode: string,
        date: Date,
        matchday: number,
        homeTeamOdds: number = 0,
        awayTeamOdds: number = 0,
        drawOdds: number = 0,
        homeGoals: number = 0,
        awayGoals: number = 0,
        status: MatchStatus = 'scheduled'
    ) {
        this.homeTeam = homeTeam;
        this.awayTeam = awayTeam;
        this.seasonCode = seasonCode;
        this.competitionCode = competitionCode;
        this.date = date;
        this.matchday = matchday;
        this.homeTeamOdds = homeTeamOdds;
        this.awayTeamOdds = awayTeamOdds;
        this.drawOdds = drawOdds;
        this.homeGoals = homeGoals;
        this.awayGoals = awayGoals;
        this.status = status;
    }

    id(): string {
        return `${this.competitionCode}-${this.seasonCode}-${this.homeTeam}-${this.awayTeam}-${this.matchday}`;
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

    isFinished(): boolean {
        return this.status === 'finished';
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

    getDate(): Date {
        return this.date;
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

    static fromJSON(json: any): SeasonMatch {
        return new SeasonMatch(
            json.homeTeam,
            json.awayTeam,
            json.seasonCode,
            json.competitionCode,
            new Date(json.date),
            json.matchday,
            json.homeTeamOdds,
            json.awayTeamOdds,
            json.drawOdds,
            json.homeGoals,
            json.awayGoals,
            json.status
        );
    }

    toJSON(): any {
        return {
            id: this.id(),
            seasonCode: this.seasonCode,
            competitionCode: this.competitionCode,
            date: this.date.toISOString(),
            homeTeam: this.homeTeam,
            awayTeam: this.awayTeam,
            homeGoals: this.homeGoals,
            awayGoals: this.awayGoals,
            homeTeamOdds: this.homeTeamOdds,
            awayTeamOdds: this.awayTeamOdds,
            drawOdds: this.drawOdds,
            status: this.status,
            matchday: this.matchday
        };
    }
} 