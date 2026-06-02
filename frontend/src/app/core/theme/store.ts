import { defineStore } from 'pinia';
import { ref } from 'vue';
import { send } from '@/app/core/http/api';

export type Theme = 'light' | 'dark' | 'system';

function resolveDark(theme: Theme): boolean {
	if (theme === 'system') {
		return window.matchMedia('(prefers-color-scheme: dark)').matches;
	}
	return theme === 'dark';
}

function apply(theme: Theme): void {
	document.documentElement.classList.toggle('dark', resolveDark(theme));
}

// useThemeStore owns the user's interface theme. `setTheme` persists to the
// backend and applies immediately; `hydrate` is called once on boot from the
// user's stored settings. A `system` choice tracks the OS preference live.
export const useThemeStore = defineStore('theme', () => {
	const theme = ref<Theme>('system');
	const media = window.matchMedia('(prefers-color-scheme: dark)');

	media.addEventListener('change', () => {
		if (theme.value === 'system') apply('system');
	});

	function hydrate(value: string | undefined): void {
		theme.value = (value as Theme) || 'system';
		apply(theme.value);
	}

	async function setTheme(value: Theme): Promise<void> {
		theme.value = value;
		apply(value);
		await send('PUT', '/users/me/settings', 'user-settings', { theme: value });
	}

	return { theme, hydrate, setTheme };
});
