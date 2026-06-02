import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { can as canMatch } from './permissions';
import { one, many, type JsonApiResource } from '@/app/core/http/api';

// User is the signed-in console administrator (never an application
// end-user — those live in Aegis realms).
export interface User {
	id: string;
	email: string;
	displayName: string;
	avatar?: string;
	isAdmin: boolean;
	roles: string[];
	permissions: string[];
	settings?: { theme: string };
}

// AuthProvider is an external OIDC identity provider the user may sign in
// through (e.g. a corporate IdP). Empty when only password login is enabled.
export interface AuthProvider {
	id: string;
	name: string;
}

// The auth store talks to the Forge backend via the shared JSON:API client.
// Sessions are an httpOnly cookie the backend sets on login, so there is no
// token in JS — the client sends `credentials: 'include'`.
export const useAuthStore = defineStore('auth', () => {
	const user = ref<User | null>(null);
	const initialized = ref(false);
	const isAuthenticated = computed(() => user.value !== null);

	// Legacy bearer slot for plugin clients; under cookie sessions the gateway
	// authenticates via the cookie, so app calls carry no token.
	const token = ref<string | null>(null);

	// JSON:API `users` resource → User (attributes carry the hydration payload).
	function toUser(res: JsonApiResource): User {
		const a = res.attributes;
		return {
			id: res.id,
			email: (a.email as string) ?? '',
			displayName: (a.displayName as string) ?? '',
			avatar: a.avatar as string | undefined,
			isAdmin: (a.isAdmin as boolean) ?? false,
			roles: (a.roles as string[]) ?? [],
			permissions: (a.permissions as string[]) ?? [],
			settings: a.settings as { theme: string } | undefined,
		};
	}

	function can(action: string): boolean {
		return canMatch(user.value?.permissions ?? [], action);
	}

	function hasRole(slug: string): boolean {
		return user.value?.roles?.includes(slug) ?? false;
	}

	async function login(email: string, password: string): Promise<boolean> {
		try {
			user.value = toUser(await one('POST', '/auth/login', 'sessions', { email, password }));
			return true;
		} catch {
			user.value = null;
			return false;
		}
	}

	// fetchMe restores the session on app load from the cookie.
	async function fetchMe(): Promise<boolean> {
		try {
			user.value = toUser(await one('GET', '/users/me'));
			return true;
		} catch {
			user.value = null;
			return false;
		} finally {
			initialized.value = true;
		}
	}

	async function providers(): Promise<AuthProvider[]> {
		try {
			return (await many('/auth/providers')).map((p) => ({
				id: p.id,
				name: (p.attributes.name as string) ?? '',
			}));
		} catch {
			return [];
		}
	}

	async function logout(): Promise<void> {
		user.value = null;
		// Full navigation so the browser follows the RP-initiated logout chain:
		// Forge clears its session, bounces through the IdP's end_session to
		// terminate the SSO session, then lands back on /login.
		window.location.assign('/api/auth/logout');
	}

	return {
		user,
		initialized,
		isAuthenticated,
		token,
		can,
		hasRole,
		login,
		fetchMe,
		providers,
		logout,
	};
});
