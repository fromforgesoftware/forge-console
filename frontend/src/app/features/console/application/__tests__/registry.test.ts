import { describe, it, expect } from 'vitest';
import { plugins, pluginRoutes } from '../registry';

describe('plugin registry', () => {
	it('registers the aegis, talos and gjallarhorn plugins', () => {
		const ids = plugins.map((p) => p.serviceId);
		expect(ids).toEqual(expect.arrayContaining(['aegis', 'talos', 'gjallarhorn']));
	});

	it('flattens plugin pages into authenticated routes', () => {
		const routes = pluginRoutes();
		const realms = routes.find((r) => r.path === '/aegis/realms');
		expect(realms).toBeDefined();
		expect(realms?.meta?.requiresAuth).toBe(true);
	});

	it('mounts the audit timeline and the delivery board', () => {
		const routes = pluginRoutes();
		expect(routes.some((r) => r.path === '/talos/audit-events')).toBe(true);
		expect(routes.some((r) => r.path === '/gjallarhorn/notifications')).toBe(true);
	});
});
