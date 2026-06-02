import type { RouteRecordRaw } from 'vue-router';

export const AUTH_ROUTES: RouteRecordRaw[] = [
	{ path: '', name: 'login', component: () => import('./LoginView.vue') },
];
