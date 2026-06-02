import type { AppInfo } from '@/app/features/console/stores/apps';
import type { ConsolePluginModule } from '@fromforgesoftware/forge-console-plugin';
import { systemImport } from './system';

// Runtime plugin loader, Grafana-style via SystemJS (NOT Module Federation).
// Each migrated app exposes a single SystemJS `module.js` whose default export
// is its ForgeConsolePlugin (or a zero-arg factory). The host registers the
// shared singletons in a SystemJS import map at bootstrap (see system.ts), then
// `System.import()`s a plugin's module.js so its externalised imports (vue,
// pinia, the kits, this contract) resolve to the host's live instances. The
// moduleUri comes from the app's /apps entry; an empty moduleUri means the app
// has not migrated and the hybrid registry falls back to its bundled plugin.

// importModule is the loader the package's contract expects (SystemImporter):
// given a moduleUri it `System.import()`s the plugin module and returns its
// loaded namespace ({ default } / { plugin } holding the plugin or a factory).
// Throws on failure so the caller can isolate it — resolvePlugins wraps each
// call in try/catch already.
export async function importModule(uri: string): Promise<ConsolePluginModule> {
	return (await systemImport(uri)) as ConsolePluginModule;
}

// hasRemote reports whether an app has migrated to a runtime SystemJS plugin
// module. Empty moduleUri => still a compile-time bundled plugin (the 4-plugin
// status quo); the hybrid registry falls back to the bundled plugin for these.
export function hasRemote(app: AppInfo): boolean {
	return typeof app.moduleUri === 'string' && app.moduleUri.trim() !== '';
}
