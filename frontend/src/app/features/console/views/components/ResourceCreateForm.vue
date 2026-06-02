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

// Field-spec-driven JSON:API create form. A plugin declares the fields; this
// builds the attributes and POSTs them, emitting `created` on success.
export interface FieldSpec {
	name: string;
	label: string;
	type?: 'text' | 'number' | 'select' | 'checkbox';
	options?: { value: string; label: string }[];
	required?: boolean;
}

const props = defineProps<{
	apiBase: string;
	type: string;
	title: string;
	fields: FieldSpec[];
}>();

const emit = defineEmits<{ created: [id: string] }>();

const auth = useAuthStore();
const form = reactive<Record<string, string>>(
	Object.fromEntries(props.fields.map((f) => [f.name, ''])),
);
const submitting = ref(false);
const error = ref<string | null>(null);
const createdId = ref<string | null>(null);

function attributes(): Record<string, unknown> {
	const out: Record<string, unknown> = {};
	for (const f of props.fields) {
		const raw = form[f.name];
		if (raw === '' || raw === undefined) continue;
		out[f.name] = f.type === 'number' ? Number(raw) : raw;
	}
	return out;
}

async function submit() {
	error.value = null;
	const missing = props.fields.find((f) => f.required && !form[f.name]);
	if (missing) {
		error.value = `${missing.label} is required`;
		return;
	}
	submitting.value = true;
	try {
		const created = await createResource(props.apiBase, props.type, attributes(), auth.token);
		createdId.value = created.id;
		emit('created', created.id);
	} catch (e) {
		error.value = e instanceof Error ? e.message : 'create failed';
	} finally {
		submitting.value = false;
	}
}
</script>

<template>
	<section class="mx-auto w-full max-w-2xl space-y-4">
		<h1 class="text-xl font-semibold">{{ title }}</h1>

		<Alert v-if="createdId" variant="success">
			<AlertDescription>
				Created <span class="font-mono">{{ createdId }}</span
				>.
			</AlertDescription>
		</Alert>

		<form class="space-y-4" @submit.prevent="submit">
			<FormField v-for="f in fields" :key="f.name" :label="f.label" :for="`field-${f.name}`">
				<Select
					v-if="f.type === 'select'"
					:model-value="form[f.name]"
					@update:model-value="(v) => (form[f.name] = v as string)"
				>
					<SelectTrigger :id="`field-${f.name}`"><SelectValue placeholder="—" /></SelectTrigger>
					<SelectContent>
						<SelectItem v-for="o in f.options" :key="o.value" :value="o.value">{{
							o.label
						}}</SelectItem>
					</SelectContent>
				</Select>
				<Input
					v-else
					:id="`field-${f.name}`"
					v-model="form[f.name]"
					:type="f.type === 'number' ? 'number' : 'text'"
				/>
			</FormField>

			<Alert v-if="error" variant="destructive">
				<AlertDescription>{{ error }}</AlertDescription>
			</Alert>

			<div class="flex justify-end">
				<Button type="submit" :disabled="submitting">
					{{ submitting ? 'Saving…' : 'Create' }}
				</Button>
			</div>
		</form>
	</section>
</template>
