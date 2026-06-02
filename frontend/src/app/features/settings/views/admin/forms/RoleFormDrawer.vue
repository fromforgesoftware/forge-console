<template>
	<Drawer v-model:open="openModel">
		<DrawerPanel>
			<DrawerHeader>
				<DrawerTitle>New role</DrawerTitle>
			</DrawerHeader>
			<RoleForm
				v-if="open"
				:permissions="permissions"
				@saved="onSaved"
				@cancel="openModel = false"
			/>
		</DrawerPanel>
	</Drawer>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { Drawer, DrawerPanel, DrawerHeader, DrawerTitle } from '@fromforgesoftware/vue-kit';
import RoleForm from './RoleForm.vue';
import type { Permission } from '../../../services/admin.service';

const props = defineProps<{ open: boolean; permissions: Permission[] }>();
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
