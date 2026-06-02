import type { RouteRecordRaw } from 'vue-router';
import { plugins } from '@/app/features/console/application/registry';

// Every plugin page becomes an authenticated child route mounted under the
// app shell. Keying RouterView by fullPath (in the layout) remounts the generic
// resource views so they refetch when navigating between siblings.
const pluginRoutes: RouteRecordRaw[] = plugins.flatMap((plugin) =>
	plugin.pages.map((page) => ({
		path: `${plugin.basePath}/${page.path}`.replace(/^\//, ''),
		name: `${plugin.serviceId}:${page.name}`,
		component: page.component,
		props: page.props,
	})),
);

export const CONSOLE_ROUTES: RouteRecordRaw[] = [
	{ path: '', name: 'home', component: () => import('./HomeView.vue') },
	{
		path: 'platform/topology',
		name: 'platform:topology',
		component: () => import('@/app/features/platform/views/PlatformTopologyView.vue'),
		meta: { requiredPermission: 'platform.read' },
	},
	...pluginRoutes,
];
