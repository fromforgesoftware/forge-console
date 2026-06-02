import { Cable } from '@lucide/vue';
import type { ForgeConsolePlugin } from '@/app/features/console/domain/plugin';
import { apiBaseFor } from '@/app/core/http/services';
import ResourceListView from '@/app/features/console/views/components/ResourceListView.vue';
import ResourceCreateForm from '@/app/features/console/views/components/ResourceCreateForm.vue';

// The Gleipnir console plugin: the read-only provider catalog, the owner's
// authorized connections (with lifecycle status), and a form to authorize a
// new connection. Credential intake is API/CLI-only (it carries secrets).
export function gleipnirPlugin(): ForgeConsolePlugin {
	const apiBase = apiBaseFor('gleipnir');
	return {
		serviceId: 'gleipnir',
		title: 'Gleipnir',
		basePath: '/gleipnir',
		apiBase,
		icon: Cable,
		order: 4,
		pages: [
			{
				path: 'connections',
				name: 'Connections',
				component: ResourceListView,
				props: {
					apiBase,
					type: 'connections',
					title: 'Connections',
					columns: ['owner', 'connector', 'status', 'expiresAt'],
				},
			},
			{
				path: 'connections/new',
				name: 'New connection',
				component: ResourceCreateForm,
				props: {
					apiBase,
					type: 'connections',
					title: 'Authorize a connection',
					fields: [
						{ name: 'owner', label: 'Owner', required: true },
						{
							name: 'connector',
							label: 'Connector',
							type: 'select',
							options: [
								{ value: 'alpaca', label: 'Alpaca' },
								{ value: 'binance', label: 'Binance' },
								{ value: 'coinbase', label: 'Coinbase' },
							],
							required: true,
						},
					],
				},
			},
			{
				path: 'connectors',
				name: 'Connectors',
				component: ResourceListView,
				props: {
					apiBase,
					type: 'connectors',
					title: 'Connector catalog',
					columns: ['name', 'authType', 'rateLimit'],
				},
			},
		],
	};
}
