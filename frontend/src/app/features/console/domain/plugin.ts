import type { Component } from 'vue';

// ForgeConsolePage is one navigable screen a plugin contributes. props are
// passed straight to the component (the generic renderer reads type/apiBase).
export interface ForgeConsolePage {
	path: string;
	name: string;
	component: Component;
	props?: Record<string, unknown>;
}

// ForgeConsolePlugin is a compile-time registered service surface. serviceId
// keys into service discovery for apiBase; pages mount under basePath. The
// manifest carries its own `icon` (a lucide/vue component) so adding a plugin
// is a single self-contained file — the sidebar, dashboard, and command
// palette all read it, with no per-service icon map to edit.
export interface ForgeConsolePlugin {
	serviceId: string;
	title: string;
	basePath: string;
	apiBase: string;
	icon: Component;
	pages: ForgeConsolePage[];
	/** Sidebar sort order (ascending). Unset sorts after numbered ones. */
	order?: number;
}
