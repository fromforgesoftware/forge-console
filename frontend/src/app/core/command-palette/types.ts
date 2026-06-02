import type { Component } from 'vue';
import type { Router, RouteLocationNormalizedLoaded } from 'vue-router';

export type CommandGroup = 'navigation' | 'action';

export interface CommandContext {
	route: RouteLocationNormalizedLoaded;
	router: Router;
}

export interface Command {
	id: string;
	title: string;
	group: CommandGroup;
	icon: Component;
	/** Synonyms for matching. */
	keywords?: string[];
	/** Breadcrumb shown before the title, e.g. the app name. */
	parent?: string;
	handler: (ctx: CommandContext) => void | Promise<void>;
}

export type CommandDispose = () => void;
