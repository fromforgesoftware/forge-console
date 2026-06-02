<template>
	<div class="flex min-h-0 flex-1 flex-col">
		<div class="min-h-0 flex-1 space-y-4 overflow-y-auto p-4">
			<template v-if="!result">
				<FormField label="Name" for="sa-name">
					<Input id="sa-name" v-model="form.name" placeholder="ci-deployer" />
				</FormField>
				<p v-if="error" class="text-sm text-destructive">{{ error }}</p>
			</template>

			<template v-else>
				<Alert>
					<AlertTitle>Save these credentials now</AlertTitle>
					<AlertDescription>
						The client secret is shown only once and cannot be retrieved later.
					</AlertDescription>
				</Alert>
				<FormField label="Client ID" for="sa-client-id">
					<Input id="sa-client-id" :model-value="result.clientId" readonly />
				</FormField>
				<FormField label="Client secret" for="sa-client-secret">
					<div class="flex gap-2">
						<Input id="sa-client-secret" :model-value="result.clientSecret" readonly />
						<Button variant="outline" @click="copySecret">
							<Check v-if="copied" class="size-4" />
							<Copy v-else class="size-4" />
						</Button>
					</div>
				</FormField>
			</template>
		</div>

		<DrawerFooter>
			<template v-if="!result">
				<Button variant="outline" @click="emit('cancel')">Cancel</Button>
				<Button :disabled="busy || !form.name" @click="submit">Create</Button>
			</template>
			<template v-else>
				<Button @click="emit('saved')">Done</Button>
			</template>
		</DrawerFooter>
	</div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { Check, Copy } from '@lucide/vue';
import {
	Alert,
	AlertTitle,
	AlertDescription,
	Button,
	DrawerFooter,
	FormField,
	Input,
} from '@fromforgesoftware/vue-kit';
import { adminService, type NewServiceAccountResult } from '../../../services/admin.service';

const emit = defineEmits<{ saved: []; cancel: [] }>();

const form = ref({ name: '' });
const busy = ref(false);
const error = ref('');
const result = ref<NewServiceAccountResult | null>(null);
const copied = ref(false);

async function submit(): Promise<void> {
	busy.value = true;
	error.value = '';
	try {
		result.value = await adminService.createServiceAccount(form.value.name);
	} catch {
		error.value = 'Failed to create service account.';
	} finally {
		busy.value = false;
	}
}

async function copySecret(): Promise<void> {
	if (!result.value) return;
	try {
		await navigator.clipboard.writeText(result.value.clientSecret);
		copied.value = true;
		setTimeout(() => (copied.value = false), 1500);
	} catch {
		error.value = 'Could not copy to clipboard.';
	}
}
</script>
