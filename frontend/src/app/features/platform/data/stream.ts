import { ref, type Ref } from 'vue';
import { useWebSocket } from '@vueuse/core';
import type { Topology, TopologyNode, TopologyEdge, ClusterInfo } from '../domain/topology';

// The platform stream is forge-native and same-origin, so the websocket
// handshake carries the session cookie automatically. It reuses the kit
// envelope: a `subscribe` frame (subject = subscription id) opens a feed and the
// server pushes `message` frames tagged with that subject.
function streamURL(): string {
	const u = new URL('/api/platform/stream', window.location.origin);
	u.protocol = u.protocol === 'https:' ? 'wss:' : 'ws:';
	return u.toString();
}

interface Envelope {
	type?: string;
	topic?: string;
	subject?: string;
	data?: unknown;
}

function parse(data: unknown): Envelope | null {
	if (typeof data !== 'string') return null;
	try {
		return JSON.parse(data) as Envelope;
	} catch {
		return null;
	}
}

// usePlatformStream keeps the topology live: it subscribes to the `topology`
// topic and invokes onTopology with each informer-driven snapshot.
export function usePlatformStream(onTopology: (t: Topology) => void) {
	const { status, send } = useWebSocket(streamURL(), {
		immediate: true,
		autoReconnect: { retries: -1, delay: 2000 },
		heartbeat: false,
		onConnected() {
			send(JSON.stringify({ type: 'subscribe', topic: 'topology', subject: 'topo' }));
		},
		onMessage(_ws, ev) {
			const msg = parse(ev.data);
			if (!msg || msg.type !== 'message' || msg.topic !== 'topology') return;
			const d = (msg.data ?? {}) as {
				cluster?: ClusterInfo;
				nodes?: TopologyNode[];
				edges?: TopologyEdge[];
			};
			onTopology({
				cluster: d.cluster ?? { name: '', version: '', nodeCount: 0, available: false },
				nodes: d.nodes ?? [],
				edges: d.edges ?? [],
			});
		},
	});
	return { status };
}

export interface LogTarget {
	namespace: string;
	pod: string;
	container?: string;
}

// useLogStream opens its own short-lived connection for the detail drawer's
// Logs tab, streaming a single pod's log lines until stopped.
export function useLogStream(): {
	lines: Ref<string[]>;
	start: (t: LogTarget) => void;
	stop: () => void;
} {
	const lines = ref<string[]>([]);
	let ws: ReturnType<typeof useWebSocket> | null = null;

	function start(target: LogTarget) {
		stop();
		lines.value = [];
		ws = useWebSocket(streamURL(), {
			immediate: true,
			autoReconnect: false,
			heartbeat: false,
			onConnected() {
				ws?.send(
					JSON.stringify({ type: 'subscribe', topic: 'logs', subject: 'logs', data: target }),
				);
			},
			onMessage(_w, ev) {
				const msg = parse(ev.data);
				if (!msg || msg.topic !== 'logs') return;
				const line = (msg.data as { line?: string })?.line;
				if (typeof line === 'string') {
					lines.value.push(line);
					if (lines.value.length > 1000) lines.value.splice(0, lines.value.length - 1000);
				}
			},
		});
	}

	function stop() {
		ws?.close();
		ws = null;
	}

	return { lines, start, stop };
}
