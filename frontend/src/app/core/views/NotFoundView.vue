<template>
	<div class="min-h-screen w-full flex items-center justify-center p-4">
		<div class="flex flex-col items-center gap-4 text-center">
			<p class="text-6xl font-bold tracking-tight text-muted-foreground">404</p>
			<h1 class="text-xl font-semibold">Page not found</h1>
			<p class="text-sm text-muted-foreground">The page you are looking for does not exist.</p>
			<Button @click="router.push('/')">Back to home</Button>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import { Button } from '@fromforgesoftware/vue-kit';

const router = useRouter();
</script>
