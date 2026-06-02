<template>
	<Drawer v-model:open="openModel">
		<DrawerPanel>
			<DrawerHeader>
				<DrawerTitle>Roles</DrawerTitle>
				<p v-if="user" class="text-sm text-muted-foreground">
					Assign roles to {{ user.displayName }}.
				</p>
			</DrawerHeader>
			<UserRolesForm
				v-if="open && user"
				:user="user"
				:roles="roles"
				:can-write="canWrite"
				@saved="onSaved"
				@cancel="openModel = false"
			/>
		</DrawerPanel>
	</Drawer>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { Drawer, DrawerPanel, DrawerHeader, DrawerTitle } from '@fromforgesoftware/vue-kit';
import UserRolesForm from './UserRolesForm.vue';
import type { AdminUser, Role } from '../../../services/admin.service';

const props = defineProps<{
	open: boolean;
	user: AdminUser | null;
	roles: Role[];
	canWrite: boolean;
}>();
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
