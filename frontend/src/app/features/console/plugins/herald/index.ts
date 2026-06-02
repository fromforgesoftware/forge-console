import { BellRing } from '@lucide/vue';
import type { ForgeConsolePlugin } from '@/app/features/console/domain/plugin';
import { apiBaseFor } from '@/app/core/http/services';
import ResourceListView from '@/app/features/console/views/components/ResourceListView.vue';
import ResourceCreateForm from '@/app/features/console/views/components/ResourceCreateForm.vue';
import ActionForm from '@/app/features/console/views/components/ActionForm.vue';

// The Herald console plugin: a delivery-status board, a test-send form (now
// covering the webhook channel + scheduledAt), and per-recipient channel
// opt-out preferences.
export function heraldPlugin(): ForgeConsolePlugin {
	const apiBase = apiBaseFor('herald');
	return {
		serviceId: 'herald',
		title: 'Herald',
		basePath: '/herald',
		apiBase,
		icon: BellRing,
		order: 3,
		pages: [
			{
				path: 'notifications',
				name: 'Delivery board',
				component: ResourceListView,
				props: {
					apiBase,
					type: 'notifications',
					title: 'Delivery board',
					columns: ['recipient', 'channel', 'status', 'subject'],
				},
			},
			{
				path: 'notifications/new',
				name: 'Test send',
				component: ResourceCreateForm,
				props: {
					apiBase,
					type: 'notifications',
					title: 'Test send',
					fields: [
						{ name: 'recipient', label: 'Recipient', required: true },
						{
							name: 'channel',
							label: 'Channel',
							type: 'select',
							options: [
								{ value: 'EMAIL', label: 'Email' },
								{ value: 'WEBHOOK', label: 'Webhook' },
							],
							required: true,
						},
						{ name: 'subject', label: 'Subject' },
						{ name: 'body', label: 'Body' },
						{ name: 'template', label: 'Template' },
						{ name: 'realmId', label: 'Realm ID' },
						{ name: 'scheduledAt', label: 'Scheduled at (RFC3339, blank = now)' },
					],
				},
			},
			{
				path: 'preferences',
				name: 'Set preference',
				component: ActionForm,
				props: {
					apiBase,
					path: '/api/notification-preferences',
					type: 'notification-preferences',
					title: 'Set notification preference',
					submitLabel: 'Save',
					fields: [
						{ name: 'recipient', label: 'Recipient', required: true },
						{
							name: 'channel',
							label: 'Channel',
							type: 'select',
							options: [
								{ value: 'EMAIL', label: 'Email' },
								{ value: 'WEBHOOK', label: 'Webhook' },
							],
							required: true,
						},
						{ name: 'realmId', label: 'Realm ID' },
						{ name: 'suppressed', label: 'Suppressed (mute this channel)', type: 'checkbox' },
					],
				},
			},
		],
	};
}
