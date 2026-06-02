import {
	__federation_method_setRemote,
	__federation_method_getRemote,
	__federation_method_unwrapDefault,
} from 'virtual:__federation__';
import type { AppInfo } from '@/app/features/console/stores/apps';
import type { RemotePluginModule } from '@fromforgesoftware/forge-console-plugin';

// Runtime Module-Federation loader. The console is built as an MF *host* with
// an empty static remotes map (see vite.config.ts); remotes are discovered from
// /apps at runtime and registered here Grafana-style. We use
// @originjs/vite-plugin-federation's dynamic-remote API from the
// `virtual:__federation__` runtime:
//   - __federation_method_setRemote(name, { url, format, from }) registers it
//   - __federation_method_getRemote(name, './plugin') loads the exposed module
// setRemote is idempotent per name, so registering twice is harmless.

// MF remote names must be valid JS identifiers; slugs are already
// `[a-z0-9-]`, so normalise the only illegal char (`-`).
function remoteName(slug: string): string {
	return `forge_remote_${slug.replace(/[^a-zA-Z0-9_$]/g, '_')}`;
}

// importRemote is the RemoteImporter the package's loadConsolePlugins expects:
// given a moduleUri it registers the remote and returns its `./plugin` module
// (a { plugin } / { default } factory holder). Throws on failure so the caller
// can isolate it — loadConsolePlugins wraps each call in try/catch already.
export async function importRemote(uri: string): Promise<RemotePluginModule> {
	// Derive a stable remote name from the URL so the same remote registered
	// twice collapses to one entry.
	const name = remoteName(uri);
	__federation_method_setRemote(name, {
		url: uri,
		format: 'esm',
		from: 'vite',
	});
	const mod = await __federation_method_getRemote(name, './plugin');
	// vite-plugin-federation may wrap an ESM default; unwrap so a remote that
	// `export default () => plugin` is read the same as `export const plugin`.
	const resolved = (await __federation_method_unwrapDefault(mod)) as RemotePluginModule;
	return resolved;
}

// hasRemote reports whether an app has migrated to a runtime plugin remote.
// Empty moduleUri => still a compile-time bundled plugin (the 4-plugin status
// quo); the hybrid registry falls back to the bundled plugin for these.
export function hasRemote(app: AppInfo): boolean {
	return typeof app.moduleUri === 'string' && app.moduleUri.trim() !== '';
}
