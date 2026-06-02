export type AppEnvironment = 'LOCAL' | 'DEV' | 'STAGING' | 'PROD';

export interface EnvironmentConfig {
	environment: AppEnvironment;
	apiUrl: string;
	services: string;
}

function parseEnvironment(raw: string | undefined): AppEnvironment {
	const value = raw?.toString().toUpperCase();
	if (value === 'PROD' || value === 'STAGING' || value === 'DEV' || value === 'LOCAL') {
		return value;
	}
	return 'LOCAL';
}

export const environment: EnvironmentConfig = {
	environment: parseEnvironment(import.meta.env.VITE_ENVIRONMENT),
	apiUrl: import.meta.env.VITE_API_URL ?? '/api',
	services: import.meta.env.VITE_FORGE_SERVICES ?? '',
};
