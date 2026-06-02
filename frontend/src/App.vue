<template>
	<RouterView />
	<CommandPaletteRoot />
	<Toaster />
</template>

<script setup lang="ts">
import { watch } from 'vue';
import { RouterView } from 'vue-router';
import { Toaster } from '@fromforgesoftware/vue-kit';
import { CommandPaletteRoot } from '@/app/core/command-palette';
import { useAuthStore } from '@/app/core/auth/store';
import { useThemeStore } from '@/app/core/theme/store';

const auth = useAuthStore();
const theme = useThemeStore();

watch(
	() => auth.user?.settings?.theme,
	(value) => theme.hydrate(value),
	{ immediate: true },
);
</script>
