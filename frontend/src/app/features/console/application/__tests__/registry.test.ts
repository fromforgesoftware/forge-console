import { describe, it, expect } from 'vitest';
import { plugins, pluginRoutes } from '../registry';

describe('plugin registry', () => {
	it('registers the talos, gjallarhorn and gleipnir plugins', () => {
		const ids = plugins.map((p) => p.serviceId);
		expect(ids).toEqual(expect.arrayContaining(['talos', 'gjallarhorn', 'gleipnir']));
	});

	it('does not compile-time register aegis (it resolves at runtime from its remote)', () => {
		const ids = plugins.map((p) => p.serviceId);
		expect(ids).not.toContain('aegis');
	});

	it('flattens plugin pages into authenticated routes', () => {
		const routes = pluginRoutes();
		const audit = routes.find((r) => r.path === '/talos/audit-events');
		expect(audit).toBeDefined();
		expect(audit?.meta?.requiresAuth).toBe(true);
	});

	it('mounts the audit timeline and the delivery board', () => {
		const routes = pluginRoutes();
		expect(routes.some((r) => r.path === '/talos/audit-events')).toBe(true);
		expect(routes.some((r) => r.path === '/gjallarhorn/notifications')).toBe(true);
	});
});
