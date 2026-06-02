<template>
	<div class="space-y-8">
		<section class="space-y-1">
			<h1 class="text-2xl font-semibold">Preferences</h1>
			<p class="text-sm text-muted-foreground">Customize how the console looks and feels.</p>
		</section>

		<section class="space-y-3">
			<h2 class="text-xs font-medium uppercase tracking-wide text-muted-foreground">
				Interface and theme
			</h2>
			<div class="rounded-xl border border-border bg-card divide-y divide-border">
				<SettingsRow title="Theme" description="Choose your interface color scheme.">
					<Select :model-value="theme.theme" @update:model-value="onThemeChange">
						<SelectTrigger class="w-56"><SelectValue /></SelectTrigger>
						<SelectContent>
							<SelectItem v-for="option in THEME_OPTIONS" :key="option.value" :value="option.value">
								{{ option.label }}
							</SelectItem>
						</SelectContent>
					</Select>
				</SettingsRow>
			</div>
		</section>
	</div>
</template>

<script setup lang="ts">
import {
	Select,
	SelectTrigger,
	SelectValue,
	SelectContent,
	SelectItem,
} from '@fromforgesoftware/vue-kit';
import { useThemeStore, type Theme } from '@/app/core/theme/store';
import SettingsRow from '../SettingsRow.vue';

const theme = useThemeStore();

const THEME_OPTIONS: { value: Theme; label: string }[] = [
	{ value: 'light', label: 'Light' },
	{ value: 'dark', label: 'Dark' },
	{ value: 'system', label: 'System' },
];

function onThemeChange(value: unknown): void {
	void theme.setTheme(value as Theme);
}
</script>
