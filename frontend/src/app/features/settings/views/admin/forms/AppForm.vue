<template>
	<div class="flex min-h-0 flex-1 flex-col">
		<div class="min-h-0 flex-1 space-y-4 overflow-y-auto p-4">
			<FormField label="Name" for="app-name">
				<Input id="app-name" v-model="form.name" />
			</FormField>
			<FormField label="Kind" for="app-kind">
				<Input id="app-kind" v-model="form.kind" />
			</FormField>
			<FormField label="Admin base URL" for="app-url">
				<Input id="app-url" v-model="form.adminBaseURL" placeholder="https://..." />
			</FormField>
			<FormField label="Enabled" description="Show this app in the console.">
				<Switch v-model:checked="form.enabled" />
			</FormField>

			<p v-if="error" class="text-sm text-destructive">{{ error }}</p>
		</div>

		<DrawerFooter>
			<Button variant="outline" @click="emit('cancel')">Cancel</Button>
			<Button :disabled="busy" @click="submit">Save</Button>
		</DrawerFooter>
	</div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import { Button, DrawerFooter, FormField, Input, Switch } from '@fromforgesoftware/vue-kit';
import { adminService, type App } from '../../../services/admin.service';

const props = defineProps<{ app: App }>();
const emit = defineEmits<{ saved: []; cancel: [] }>();

const form = ref<App>({ ...props.app });
const busy = ref(false);
const error = ref('');

watch(
	() => props.app,
	(p) => (form.value = { ...p }),
);

async function submit(): Promise<void> {
	busy.value = true;
	error.value = '';
	try {
		const { slug, ...rest } = form.value;
		await adminService.updateApp(slug, rest);
		emit('saved');
	} catch {
		error.value = 'Failed to save app.';
	} finally {
		busy.value = false;
	}
}
</script>
