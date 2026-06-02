<script setup lang="ts">
import { reactive, ref } from 'vue';
import {
	Alert,
	AlertDescription,
	Button,
	Checkbox,
	FormField,
	Input,
	Select,
	SelectTrigger,
	SelectValue,
	SelectContent,
	SelectItem,
} from '@fromforgesoftware/vue-kit';
import { useAuthStore } from '@/app/core/auth/store';
import { postCommand } from '@/app/core/http/jsonapi';
import type { FieldSpec } from '@/app/features/console/views/components/ResourceCreateForm.vue';

// Field-spec-driven form for JSON:API *command* endpoints whose path differs
// from the resource type (ban, unban, merge, set-preference, …). Mirrors
// ResourceCreateForm but targets an explicit path and shows the result.
const props = defineProps<{
	apiBase: string;
	path: string;
	type: string;
	title: string;
	submitLabel?: string;
	fields: FieldSpec[];
}>();

const auth = useAuthStore();
const form = reactive<Record<string, string | boolean>>(
	Object.fromEntries(props.fields.map((f) => [f.name, f.type === 'checkbox' ? false : ''])),
);
const submitting = ref(false);
const error = ref<string | null>(null);
const done = ref(false);

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
	done.value = false;
	const missing = props.fields.find((f) => f.required && !form[f.name]);
	if (missing) {
		error.value = `${missing.label} is required`;
		return;
	}
	submitting.value = true;
	try {
		await postCommand(props.apiBase, props.path, props.type, attributes(), auth.token);
		done.value = true;
	} catch (e) {
		error.value = e instanceof Error ? e.message : 'request failed';
	} finally {
		submitting.value = false;
	}
}
</script>

<template>
	<section class="mx-auto w-full max-w-2xl space-y-4">
		<h1 class="text-xl font-semibold">{{ title }}</h1>

		<Alert v-if="done" variant="success">
			<AlertDescription>Done.</AlertDescription>
		</Alert>

		<form class="space-y-4" @submit.prevent="submit">
			<template v-for="f in fields" :key="f.name">
				<!-- Checkbox fields render as a single inline control -->
				<div v-if="f.type === 'checkbox'" class="flex items-center gap-2">
					<Checkbox
						:id="`field-${f.name}`"
						:checked="form[f.name] as boolean"
						@update:checked="(v) => (form[f.name] = v === true)"
					/>
					<label :for="`field-${f.name}`" class="text-sm font-medium">{{ f.label }}</label>
				</div>

				<FormField v-else :label="f.label" :for="`field-${f.name}`">
					<Select
						v-if="f.type === 'select'"
						:model-value="form[f.name] as string"
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
						:model-value="form[f.name] as string"
						:type="f.type === 'number' ? 'number' : 'text'"
						@update:model-value="(v) => (form[f.name] = v as string)"
					/>
				</FormField>
			</template>

			<Alert v-if="error" variant="destructive">
				<AlertDescription>{{ error }}</AlertDescription>
			</Alert>

			<div class="flex justify-end">
				<Button type="submit" :disabled="submitting">
					{{ submitting ? 'Submitting…' : (submitLabel ?? 'Submit') }}
				</Button>
			</div>
		</form>
	</section>
</template>
