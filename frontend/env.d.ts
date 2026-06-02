/// <reference types="vite/client" />

declare module '*.vue' {
	import type { DefineComponent } from 'vue';
	const component: DefineComponent<object, object, unknown>;
	export default component;
}

// @originjs/vite-plugin-federation injects a `virtual:__federation__` runtime
// at build time exposing the dynamic-remote API. The runtime loader imports it
// to register + load app plugin remotes discovered from /apps at runtime.
declare module 'virtual:__federation__' {
	export interface FederationRemote {
		url: string | (() => Promise<string>);
		format?: 'esm' | 'systemjs' | 'var';
		from?: 'vite' | 'webpack';
	}
	export function __federation_method_setRemote(name: string, remote: FederationRemote): void;
	export function __federation_method_getRemote(name: string, exposed: string): Promise<unknown>;
	export function __federation_method_unwrapDefault(mod: unknown): Promise<unknown>;
}

interface ImportMetaEnv {
	readonly VITE_AEGIS_BASE?: string;
	readonly VITE_AEGIS_CLIENT_ID?: string;
	readonly VITE_FORGE_SERVICES?: string;
	readonly VITE_ENVIRONMENT?: string;
	readonly VITE_API_URL?: string;
}
interface ImportMeta {
	readonly env: ImportMetaEnv;
}
