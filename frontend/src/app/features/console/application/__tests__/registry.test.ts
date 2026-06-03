import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { AppInfo } from '@/app/features/console/stores/apps';

// The hybrid loader pulls plugin code from app remotes via runtime.importModule
// (SystemJS). Mock the runtime so these unit tests exercise the resolver's
// empty-bundled + graceful-skip behaviour without loading real remotes.
const importModule = vi.fn();
vi.mock('../runtime', () => ({
	importModule: (uri: string) => importModule(uri),
	hasRemote: (app: AppInfo) => typeof app.moduleUri === 'string' && app.moduleUri.trim() !== '',
}));

import { plugins, pluginRoutes, resolvePlugins } from '../registry';

const app = (over: Partial<AppInfo> = {}): AppInfo => ({
	slug: 'talos',
	name: 'Talos',
	kind: 'app',
	moduleUri: '',
	...over,
});

describe('plugin registry', () => {
	it('ships NO compile-time bundled plugins (every app resolves from its remote)', () => {
		expect(plugins).toEqual([]);
	});

	it('flattens to no routes when the bundled registry is empty', () => {
		expect(pluginRoutes()).toEqual([]);
	});
});

describe('hybrid resolvePlugins (no bundled fallback)', () => {
	beforeEach(() => {
		importModule.mockReset();
	});

	it('skips apps without a remote module (empty moduleUri, no bundled plugin)', async () => {
		const resolved = await resolvePlugins([
			app({ slug: 'talos' }),
			app({ slug: 'gjallarhorn' }),
			app({ slug: 'gleipnir' }),
		]);
		expect(resolved).toEqual([]);
		expect(importModule).not.toHaveBeenCalled();
	});

	it('resolves purely from the remote module when an app exposes one', async () => {
		importModule.mockResolvedValue({
			default: {
				serviceId: 'talos',
				title: 'Talos',
				basePath: '/talos',
				apiBase: '',
				pages: [{ path: 'audit-events', name: 'Audit timeline', component: {} }],
			},
		});

		const resolved = await resolvePlugins([
			app({ slug: 'talos', moduleUri: 'https://example/talos/module.js' }),
		]);

		expect(importModule).toHaveBeenCalledWith('https://example/talos/module.js');
		expect(resolved).toHaveLength(1);
		expect(resolved[0]?.serviceId).toBe('talos');
		expect(resolved[0]?.type).toBe('app');
	});

	it('gracefully skips an app whose remote module fails to load (no bundled fallback)', async () => {
		importModule.mockRejectedValue(new Error('404'));
		const err = vi.spyOn(console, 'error').mockImplementation(() => {});

		const resolved = await resolvePlugins([
			app({ slug: 'talos', moduleUri: 'https://example/talos/module.js' }),
		]);

		expect(resolved).toEqual([]);
		expect(err).toHaveBeenCalled();
		err.mockRestore();
	});
});
