<template>
	<div class="px-3 py-3">
		<div
			:class="
				orientation === 'vertical'
					? 'flex min-h-10 flex-col gap-2'
					: 'flex min-h-10 items-center justify-between gap-1'
			"
		>
			<div class="flex min-w-0 flex-grow flex-col gap-0.5">
				<div class="flex items-center gap-1.5">
					<p class="text-sm/normal font-medium">{{ title }}</p>
					<Tooltip v-if="tooltip" :content="tooltip" side="right">
						<Info class="size-3.5 text-muted-foreground cursor-help shrink-0" />
					</Tooltip>
				</div>
				<p v-if="description" class="text-sm text-muted-foreground max-w-prose">
					{{ description }}
				</p>
			</div>
			<slot />
		</div>
		<div v-if="$slots.footer" class="mt-3">
			<slot name="footer" />
		</div>
	</div>
</template>

<script setup lang="ts">
import { Tooltip } from '@fromforgesoftware/vue-kit';
import { Info } from '@lucide/vue';

withDefaults(
	defineProps<{
		title: string;
		description?: string;
		tooltip?: string;
		orientation?: 'horizontal' | 'vertical';
	}>(),
	{
		orientation: 'horizontal',
	},
);
</script>
