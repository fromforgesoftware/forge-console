import { describe, it, expect } from 'vitest';
import { roleAttributes, bindingAttributes } from '../forms';

describe('roleAttributes', () => {
	it('includes the selected permissions for atomic role seeding', () => {
		const attrs = roleAttributes(
			{ realmId: 'r', name: 'editor', resourceType: 'doc', kind: 'CUSTOM' },
			['doc.read', 'doc.write'],
		);
		expect(attrs).toEqual({
			realmId: 'r',
			name: 'editor',
			resourceType: 'doc',
			kind: 'CUSTOM',
			permissions: ['doc.read', 'doc.write'],
		});
	});
});

describe('bindingAttributes', () => {
	it('maps subject/role/resource into binding attributes', () => {
		expect(
			bindingAttributes({
				resourceId: 'res',
				roleId: 'role',
				subjectType: 'ACCOUNT',
				subjectId: 'acc',
			}),
		).toEqual({ resourceId: 'res', roleId: 'role', subjectType: 'ACCOUNT', subjectId: 'acc' });
	});
});
