import type { NavigationGuardNext, RouteLocationNormalized } from 'vue-router';
import { useAuthStore } from '@/app/core/auth/store';
import { useAppsStore } from '@/app/features/console/stores/apps';
import { useRealmStore } from '@/app/features/console/stores/realm';

// On the first navigation we restore the session from the cookie and, when
// authed, load the apps + realms the shell needs. A deep-link refresh
// therefore resolves auth before deciding whether to bounce to /login.
export async function authGuard(
	to: RouteLocationNormalized,
	_from: RouteLocationNormalized,
	next: NavigationGuardNext,
) {
	const auth = useAuthStore();

	if (!auth.initialized) {
		await auth.fetchMe();
		if (auth.isAuthenticated) {
			await Promise.all([useAppsStore().load(), useRealmStore().load()]);
		}
	}

	const required = to.meta.requiredPermission as string | undefined;

	if (to.meta.requiresAuth && !auth.isAuthenticated) {
		next({ name: 'login', query: { redirect: to.fullPath } });
	} else if (to.name === 'login' && auth.isAuthenticated) {
		next({ name: 'home' });
	} else if (required && auth.isAuthenticated && !auth.can(required)) {
		next({ name: 'home' });
	} else {
		next();
	}
}
