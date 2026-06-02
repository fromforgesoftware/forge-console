import { describe, it, expect } from 'vitest';
import { auditStreamURL, subscribeMessage, unsubscribeMessage, parseAuditMessage } from '../stream';

describe('auditStreamURL', () => {
	it('switches http→ws', () => {
		expect(auditStreamURL('http://localhost:8081')).toBe(
			'ws://localhost:8081/api/audit-events/stream',
		);
	});

	it('switches https→wss', () => {
		expect(auditStreamURL('https://hallmark.example.com')).toBe(
			'wss://hallmark.example.com/api/audit-events/stream',
		);
	});

	it('resolves a relative apiBase against the origin', () => {
		expect(auditStreamURL('/__services/hallmark', 'https://console.example.com')).toBe(
			'wss://console.example.com/__services/hallmark/api/audit-events/stream',
		);
	});
});

describe('subscribe/unsubscribe messages', () => {
	it('builds a subscribe frame with only non-empty filters', () => {
		const msg = JSON.parse(subscribeMessage('panel-1', { action: 'doc.updated', actorId: '' }));
		expect(msg).toEqual({
			type: 'subscribe',
			id: 'panel-1',
			topic: 'audit',
			data: { action: 'doc.updated' },
		});
	});

	it('includes replayFrom when provided', () => {
		const msg = JSON.parse(subscribeMessage('panel-1', {}, '2026-01-01T00:00:00Z'));
		expect(msg.data).toEqual({ replayFrom: '2026-01-01T00:00:00Z' });
	});

	it('omits replayFrom when blank', () => {
		const msg = JSON.parse(subscribeMessage('panel-1', { action: 'x' }, ''));
		expect(msg.data).toEqual({ action: 'x' });
	});

	it('builds an unsubscribe frame', () => {
		expect(JSON.parse(unsubscribeMessage('panel-1'))).toEqual({
			type: 'unsubscribe',
			id: 'panel-1',
		});
	});
});

describe('parseAuditMessage', () => {
	it('parses a tagged audit event frame', () => {
		const got = parseAuditMessage(
			JSON.stringify({
				type: 'message',
				topic: 'audit',
				subject: 'panel-1',
				sn: 3,
				data: { id: 'evt-1', action: 'doc.updated' },
			}),
		);
		expect(got?.subId).toBe('panel-1');
		expect(got?.sn).toBe(3);
		expect(got?.frame.id).toBe('evt-1');
		expect(got?.frame.action).toBe('doc.updated');
		expect(got?.frame.actorId).toBe('');
	});

	it('ignores non-event frames (welcome/ack/ping)', () => {
		expect(parseAuditMessage(JSON.stringify({ type: 'ack', id: 'panel-1' }))).toBeNull();
		expect(parseAuditMessage(JSON.stringify({ type: 'welcome' }))).toBeNull();
	});

	it('returns null for malformed data', () => {
		expect(parseAuditMessage('not json')).toBeNull();
		expect(
			parseAuditMessage(JSON.stringify({ type: 'message', topic: 'audit', data: {} })),
		).toBeNull();
	});
});
