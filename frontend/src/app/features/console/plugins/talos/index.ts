import { ScrollText } from '@lucide/vue';
import type { ForgeConsolePlugin } from '@fromforgesoftware/forge-console-plugin';
import { ResourceListView } from '@fromforgesoftware/forge-console-plugin/ui';
import { apiBaseFor } from '@/app/core/http/services';
import LiveAuditTail from '@/app/features/console/views/components/LiveAuditTail.vue';

// The Talos console plugin: the audit timeline over Talos's read-only
// JSON:API, plus the live tail over Talos's WebSocket stream
// (/api/audit-events/stream).
export function talosPlugin(): ForgeConsolePlugin {
	const apiBase = apiBaseFor('talos');
	return {
		serviceId: 'talos',
		type: 'app',
		title: 'Talos',
		basePath: '/talos',
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
