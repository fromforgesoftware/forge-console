import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount, flushPromises } from '@vue/test-utils';
import { setActivePinia, createPinia } from 'pinia';
import ResourceCreateForm from './ResourceCreateForm.vue';

describe('ResourceCreateForm', () => {
	beforeEach(() => {
		setActivePinia(createPinia());
		vi.restoreAllMocks();
	});

	it('POSTs the create envelope on submit and emits created', async () => {
		const fetchMock = vi.fn().mockResolvedValue(
			new Response(JSON.stringify({ data: { id: 'realm-1', type: 'realms', attributes: {} } }), {
				status: 200,
				headers: { 'content-type': 'application/vnd.api+json' },
			}),
		);
		vi.stubGlobal('fetch', fetchMock);

		const wrapper = mount(ResourceCreateForm, {
			props: {
				apiBase: 'http://aegis',
				type: 'realms',
				title: 'New realm',
				fields: [{ name: 'name', label: 'Name', required: true }],
			},
		});
		await wrapper.find('input').setValue('acme');
		await wrapper.find('form').trigger('submit.prevent');
		await flushPromises();

		const [url, init] = fetchMock.mock.calls[0];
		expect(url).toBe('http://aegis/api/realms');
		expect(JSON.parse(init.body).data.attributes.name).toBe('acme');
		expect(wrapper.emitted('created')?.[0]).toEqual(['realm-1']);
	});

	it('blocks submit when a required field is empty', async () => {
		const fetchMock = vi.fn();
		vi.stubGlobal('fetch', fetchMock);

		const wrapper = mount(ResourceCreateForm, {
			props: {
				apiBase: 'http://aegis',
				type: 'realms',
				title: 'New realm',
				fields: [{ name: 'name', label: 'Name', required: true }],
			},
		});
		await wrapper.find('form').trigger('submit.prevent');
		await flushPromises();
		expect(fetchMock).not.toHaveBeenCalled();
	});
});
