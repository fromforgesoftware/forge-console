<script setup lang="ts">
import { reactive, ref } from 'vue';
import {
	Alert,
	AlertDescription,
	Button,
	FormField,
	Input,
	Select,
	SelectTrigger,
	SelectValue,
	SelectContent,
	SelectItem,
} from '@fromforgesoftware/vue-kit';
import { useAuthStore } from '@/app/core/auth/store';
import { createResource } from '@/app/core/http/jsonapi';
import { bindingAttributes } from '@/app/features/console/application/forms';

// Binding editor: grant a subject (account or group) a role on a resource.
const props = defineProps<{ apiBase: string }>();

const auth = useAuthStore();
const form = reactive({ resourceId: '', roleId: '', subjectType: 'ACCOUNT', subjectId: '' });
const error = ref<string | null>(null);
const createdId = ref<string | null>(null);
const submitting = ref(false);

async function submit() {
	error.value = null;
	if (!form.resourceId || !form.roleId || !form.subjectId) {
		error.value = 'resource, role and subject are required';
		return;
	}
	submitting.value = true;
	try {
		const created = await createResource(
			props.apiBase,
			'bindings',
			bindingAttributes(form),
			auth.token,
		);
		createdId.value = created.id;
	} catch (e) {
		error.value = e instanceof Error ? e.message : 'create failed';
	} finally {
		submitting.value = false;
	}
}
</script>

<template>
	<section class="mx-auto w-full max-w-2xl space-y-4">
		<h1 class="text-xl font-semibold">New binding</h1>

		<Alert v-if="createdId" variant="success">
			<AlertDescription>
				Granted binding <span class="font-mono">{{ createdId }}</span
				>.
			</AlertDescription>
		</Alert>

		<form class="space-y-4" @submit.prevent="submit">
			<FormField label="Resource ID" for="binding-resource">
				<Input id="binding-resource" v-model="form.resourceId" />
			</FormField>
			<FormField label="Role ID" for="binding-role">
				<Input id="binding-role" v-model="form.roleId" />
			</FormField>
			<FormField label="Subject type" for="binding-subject-type">
				<Select
					:model-value="form.subjectType"
					@update:model-value="(v) => (form.subjectType = v as string)"
				>
					<SelectTrigger id="binding-subject-type"><SelectValue /></SelectTrigger>
					<SelectContent>
						<SelectItem value="ACCOUNT">ACCOUNT</SelectItem>
						<SelectItem value="ACTOR_SET">ACTOR_SET (group)</SelectItem>
					</SelectContent>
				</Select>
			</FormField>
			<FormField label="Subject ID" for="binding-subject">
				<Input id="binding-subject" v-model="form.subjectId" />
			</FormField>

			<Alert v-if="error" variant="destructive">
				<AlertDescription>{{ error }}</AlertDescription>
			</Alert>

			<div class="flex justify-end">
				<Button type="submit" :disabled="submitting">
					{{ submitting ? 'Saving…' : 'Grant binding' }}
				</Button>
			</div>
		</form>
	</section>
</template>
