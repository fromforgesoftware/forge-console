<template>
	<div
		class="hidden items-center gap-3 rounded-md border border-border bg-card px-3 py-1.5 text-xs md:flex"
	>
		<span
			v-for="item in items"
			:key="item.label"
			class="inline-flex items-center gap-1.5 text-muted-foreground"
		>
			<span class="size-2 rounded-full" :style="{ backgroundColor: item.color }" />
			{{ item.label }}
		</span>
	</div>
</template>

<script setup lang="ts">
const items = [
	{ label: 'Running', color: 'var(--color-success)' },
	{ label: 'Pending', color: 'var(--color-info)' },
	{ label: 'Degraded', color: 'var(--color-destructive)' },
	{ label: 'Paused / not deployed', color: 'var(--color-border)' },
];
</script>
