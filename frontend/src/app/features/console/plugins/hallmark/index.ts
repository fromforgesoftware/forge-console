import { ScrollText } from '@lucide/vue';
import type { ForgeConsolePlugin } from '@/app/features/console/domain/plugin';
import { apiBaseFor } from '@/app/core/http/services';
import ResourceListView from '@/app/features/console/views/components/ResourceListView.vue';
import LiveAuditTail from '@/app/features/console/views/components/LiveAuditTail.vue';

// The Hallmark console plugin: the audit timeline over Hallmark's read-only
// JSON:API, plus the live tail over Hallmark's WebSocket stream
// (/api/audit-events/stream).
export function hallmarkPlugin(): ForgeConsolePlugin {
	const apiBase = apiBaseFor('hallmark');
	return {
		serviceId: 'hallmark',
		title: 'Hallmark',
		basePath: '/hallmark',
		apiBase,
		icon: ScrollText,
		order: 2,
		pages: [
			{
				path: 'audit-events',
				name: 'Audit timeline',
				component: ResourceListView,
				props: {
					apiBase,
					type: 'audit-events',
					title: 'Audit timeline',
					columns: ['timestamp', 'action', 'actorId', 'resourceType', 'resourceId'],
				},
			},
			{
				path: 'live',
				name: 'Live tail',
				component: LiveAuditTail,
				props: { apiBase, title: 'Live audit tail' },
			},
		],
	};
}
