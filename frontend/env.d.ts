/// <reference types="vite/client" />

declare module '*.vue' {
	import type { DefineComponent } from 'vue';
	const component: DefineComponent<object, object, unknown>;
	export default component;
}

// The SystemJS browser loader is a side-effecting script that installs the
// global `System` loader; it ships no type declarations, so declare the
// side-effect-only import the host bootstrap uses (see system.ts).
declare module 'systemjs/dist/system.js';

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
