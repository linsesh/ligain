import crashlytics from '@react-native-firebase/crashlytics';

type Environment = 'development' | 'preview' | 'production' | string;

export function initCrashlytics(environment: Environment): void {
	const shouldEnable = environment === 'production';
	crashlytics().setCrashlyticsCollectionEnabled(shouldEnable);
}

export function setCrashUser(userId: string | null | undefined): void {
	if (!userId) return;
	crashlytics().setUserId(String(userId));
}

export function logError(error: unknown, context?: Record<string, unknown>): void {
	if (context) {
		try {
			crashlytics().log(JSON.stringify({ context }));
		} catch {}
	}
	if (error instanceof Error) {
		crashlytics().recordError(error);
	} else {
		crashlytics().recordError(new Error(String(error)));
	}
}

export function logMessage(message: string): void {
	crashlytics().log(message);
}

export function forceTestCrash(): void {
	crashlytics().crash();
}


