import { createRouter, createWebHistory } from 'vue-router';
import { authGuard } from './guards/authGuard';
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

router.beforeEach(authGuard);

export default router;
