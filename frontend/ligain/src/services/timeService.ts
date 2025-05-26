export interface TimeService {
    now(): Date;
}

export class RealTimeService implements TimeService {
    now(): Date {
        return new Date();
    }
}

export class MockTimeService implements TimeService {
    private currentTime: Date;

    constructor(initialTime: Date) {
        this.currentTime = initialTime;
    }

    now(): Date {
        return this.currentTime;
    }

    setTime(time: Date): void {
        this.currentTime = time;
    }
} 