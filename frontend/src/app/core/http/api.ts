import { ApiClient } from '@fromforgesoftware/ts-kit/jsonapi-client';

// The single JSON:API client for the Forge SPA. Same-origin, cookie-session
// auth (`credentials: 'include'`); the kit defaults the vnd.api+json media-type
// headers and parses JSON:API errors into ApiError. EVERY Forge API call goes
// through this — no hand-rolled fetch or bespoke envelope building.
export const api = ApiClient.create({
	baseUrl: '',
	basePath: '/api',
	credentials: 'include',
});

export interface JsonApiResource {
	id: string;
	type: string;
	attributes: Record<string, unknown>;
}

// envelope builds the JSON:API request body for a write.
export function envelope(type: string, attributes: Record<string, unknown>) {
	return { data: { type, attributes } };
}

// one issues a request and returns the single resource from the document.
export async function one(
	method: string,
	path: string,
	type?: string,
	attributes?: Record<string, unknown>,
): Promise<JsonApiResource> {
	const res = await api.request({
		method,
		path,
		body: type ? envelope(type, attributes ?? {}) : undefined,
	});
	return (res.body as { data: JsonApiResource }).data;
}

// many issues a GET and returns the resource collection from the document.
export async function many(path: string): Promise<JsonApiResource[]> {
	const res = await api.request({ method: 'GET', path });
	return (res.body as { data?: JsonApiResource[] }).data ?? [];
}

// send issues a request with no resource body in the response (e.g. 204).
export async function send(
	method: string,
	path: string,
	type?: string,
	attributes?: Record<string, unknown>,
): Promise<void> {
	await api.request({
		method,
		path,
		body: type ? envelope(type, attributes ?? {}) : undefined,
	});
}
