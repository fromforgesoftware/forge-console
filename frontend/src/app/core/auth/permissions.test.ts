import { describe, it, expect } from 'vitest';
import { permissionMatches, can } from './permissions';

describe('permissionMatches', () => {
	it('bare * matches everything', () => {
		expect(permissionMatches('*', 'users.read')).toBe(true);
		expect(permissionMatches('*', 'app:aegis.write')).toBe(true);
	});

	it('*.* matches everything', () => {
		expect(permissionMatches('*.*', 'users.read')).toBe(true);
		expect(permissionMatches('*.*', 'app:aegis.write')).toBe(true);
	});

	it('matches exact', () => {
		expect(permissionMatches('users.write', 'users.write')).toBe(true);
		expect(permissionMatches('users.write', 'users.read')).toBe(false);
	});

	it('verb wildcard matches any verb of the resource', () => {
		expect(permissionMatches('users.*', 'users.write')).toBe(true);
		expect(permissionMatches('users.*', 'users.read')).toBe(true);
		expect(permissionMatches('users.*', 'roles.read')).toBe(false);
	});

	it('resource-type wildcard matches any resource', () => {
		expect(permissionMatches('*.read', 'app:aegis.read')).toBe(true);
		expect(permissionMatches('*.read', 'users.read')).toBe(true);
		expect(permissionMatches('*.read', 'users.write')).toBe(false);
	});

	it('colon resource type matches exactly', () => {
		expect(permissionMatches('app:aegis.read', 'app:aegis.read')).toBe(true);
		expect(permissionMatches('app:aegis.read', 'app:herald.read')).toBe(false);
		expect(permissionMatches('app:aegis.read', 'app:aegis.write')).toBe(false);
	});

	it('mismatches', () => {
		expect(permissionMatches('users.read', 'roles.read')).toBe(false);
		expect(permissionMatches('roles.write', 'users.write')).toBe(false);
	});
});

describe('can', () => {
	it('is true when any granted pattern matches', () => {
		expect(can(['users.read', 'roles.read'], 'roles.read')).toBe(true);
		expect(can(['*.*'], 'app:aegis.write')).toBe(true);
	});

	it('is false when none match', () => {
		expect(can(['users.read'], 'users.write')).toBe(false);
		expect(can([], 'users.read')).toBe(false);
	});
});
