import { createRouter, createWebHistory } from 'vue-router';
import { authGuard } from './guards/authGuard';
import { makePluginGuard } from './guards/pluginGuard';
import { CONSOLE_ROUTES } from '@/app/features/console/views/ConsoleRoutes';
import { AUTH_ROUTES } from '@/app/features/auth/views/AuthRoutes';
import { SETTINGS_ROUTES } from '@/app/features/settings/views/SettingsRoutes';
import AppLayout from '@/app/core/layouts/AppLayout.vue';
import AuthLayout from '@/app/core/layouts/AuthLayout.vue';
import PlainLayout from '@/app/core/layouts/PlainLayout.vue';

const router = createRouter({
	history: createWebHistory(import.meta.env.BASE_URL),
	routes: [
		{
			path: '/',
			name: 'app-shell',
			component: AppLayout,
			meta: { requiresAuth: true },
			children: CONSOLE_ROUTES,
		},
		...SETTINGS_ROUTES,
		{
			path: '/login',
			component: AuthLayout,
			children: AUTH_ROUTES,
		},
		{
			path: '/:pathMatch(.*)*',
			component: PlainLayout,
			children: [
				{
					path: '',
					name: 'not-found',
					component: () => import('@/app/core/views/NotFoundView.vue'),
				},
			],
		},
	],
});

// authGuard resolves auth + loads /apps on the first navigation; pluginGuard
// then lazily resolves+registers a remote-backed plugin's routes the first time
// its section is visited. Order matters: pluginGuard depends on apps being
// loaded, which authGuard guarantees for authenticated navigations.
router.beforeEach(authGuard);
router.beforeEach(makePluginGuard(router));

export default router;
