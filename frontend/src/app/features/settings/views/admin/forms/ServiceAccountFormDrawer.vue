<template>
	<Drawer v-model:open="openModel">
		<DrawerPanel>
			<DrawerHeader>
				<DrawerTitle>New service account</DrawerTitle>
			</DrawerHeader>
			<ServiceAccountForm v-if="open" @saved="onSaved" @cancel="openModel = false" />
		</DrawerPanel>
	</Drawer>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { Drawer, DrawerPanel, DrawerHeader, DrawerTitle } from '@fromforgesoftware/vue-kit';
import ServiceAccountForm from './ServiceAccountForm.vue';

const props = defineProps<{ open: boolean }>();
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
