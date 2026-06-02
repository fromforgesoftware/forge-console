<template>
	<Drawer v-model:open="openModel">
		<DrawerPanel>
			<DrawerHeader>
				<DrawerTitle>Edit app</DrawerTitle>
			</DrawerHeader>
			<AppForm v-if="open && app" :app="app" @saved="onSaved" @cancel="openModel = false" />
		</DrawerPanel>
	</Drawer>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { Drawer, DrawerPanel, DrawerHeader, DrawerTitle } from '@fromforgesoftware/vue-kit';
import AppForm from './AppForm.vue';
import type { App } from '../../../services/admin.service';

const props = defineProps<{ open: boolean; app: App | null }>();
const emit = defineEmits<{ 'update:open': [value: boolean]; saved: [] }>();

const openModel = computed({
	get: () => props.open,
	set: (v) => emit('update:open', v),
});

function onSaved(): void {
	emit('update:open', false);
	emit('saved');
}
</script>
