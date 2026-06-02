import { describe, it, expect, vi, beforeEach } from 'vitest';
import { buildCreateBody, createResource, fetchCollection, postCommand } from '../jsonapi';

// The console helpers route through ts-kit's ApiClient (FetchAdapter → global
// fetch), so mocks return real Response objects with status + headers.
function jsonResponse(data: unknown, status = 200): Response {
	return new Response(JSON.stringify({ data }), {
		status,
		headers: { 'content-type': 'application/vnd.api+json' },
	});
}

describe('buildCreateBody', () => {
	it('wraps type + attributes in a JSON:API document', () => {
		expect(buildCreateBody('roles', { name: 'editor' })).toEqual({
			data: { type: 'roles', attributes: { name: 'editor' } },
		});
	});
});

describe('createResource', () => {
	beforeEach(() => vi.restoreAllMocks());

	it('POSTs the create envelope with a bearer token', async () => {
		const fetchMock = vi
			.fn()
			.mockResolvedValue(jsonResponse({ id: 'r1', type: 'roles', attributes: {} }));
		vi.stubGlobal('fetch', fetchMock);

		await createResource('http://aegis', 'roles', { name: 'editor' }, 'tok');

		const [url, init] = fetchMock.mock.calls[0];
		expect(url).toBe('http://aegis/api/roles');
		expect(init.method).toBe('POST');
		expect(init.headers.Authorization).toBe('Bearer tok');
		expect(JSON.parse(init.body)).toEqual({
			data: { type: 'roles', attributes: { name: 'editor' } },
		});
	});
});

describe('postCommand', () => {
	beforeEach(() => vi.restoreAllMocks());

	it('POSTs the envelope to the command path (path ≠ type)', async () => {
		const fetchMock = vi
			.fn()
			.mockResolvedValue(jsonResponse({ id: 'acc-1', type: 'accountBans', attributes: {} }));
		vi.stubGlobal('fetch', fetchMock);

		await postCommand(
			'http://aegis',
			'/api/accounts/ban',
			'accountBans',
			{ accountId: 'acc-1' },
			'tok',
		);

		const [url, init] = fetchMock.mock.calls[0];
		expect(url).toBe('http://aegis/api/accounts/ban');
		expect(JSON.parse(init.body)).toEqual({
			data: { type: 'accountBans', attributes: { accountId: 'acc-1' } },
		});
	});

	it('returns null on a 204 with no body', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(null, { status: 204 })));
		const out = await postCommand('http://aegis', '/api/accounts/unban', 'accountBans', {}, null);
		expect(out).toBeNull();
	});

	it('throws on a non-ok response', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(null, { status: 400 })));
		await expect(
			postCommand('http://aegis', '/api/accounts/ban', 'accountBans', {}, null),
		).rejects.toThrow();
	});
});

describe('fetchCollection', () => {
	beforeEach(() => vi.restoreAllMocks());

	it('returns data[] and omits Authorization without a token', async () => {
		const fetchMock = vi
			.fn()
			.mockResolvedValue(jsonResponse([{ id: 'a', type: 'realms', attributes: {} }]));
		vi.stubGlobal('fetch', fetchMock);

		const rows = await fetchCollection('http://aegis', 'realms', null);
		expect(rows).toHaveLength(1);
		expect(fetchMock.mock.calls[0][1].headers.Authorization).toBeUndefined();
	});
});
