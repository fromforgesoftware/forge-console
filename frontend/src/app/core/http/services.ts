import { environment } from '@/environments/environment';

// App API calls go through the Forge backend's admin-API gateway
// (/api/proxy/<id>), which authenticates the user via the session cookie
// and forwards to the app's admin API. VITE_FORGE_SERVICES can override a
// service's base URL for local dev against a directly-reachable app.
export function parseServices(raw: string | undefined): Record<string, string> {
	const out: Record<string, string> = {};
	if (!raw) return out;
	for (const pair of raw.split(',')) {
		const [id, ...rest] = pair.split('=');
		const url = rest.join('=').trim();
		if (id?.trim() && url) out[id.trim()] = url.replace(/\/$/, '');
	}
	return out;
}

const services = parseServices(environment.services);

export function apiBaseFor(serviceId: string): string {
	return services[serviceId] ?? `/api/proxy/${serviceId}`;
}
