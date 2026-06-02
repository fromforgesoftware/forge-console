import { describe, it, expect } from 'vitest';
import { parseServices } from '../services';

describe('parseServices', () => {
	it('parses id=url pairs and strips trailing slashes', () => {
		const got = parseServices('aegis=http://localhost:8080/,talos=http://h:8081');
		expect(got).toEqual({ aegis: 'http://localhost:8080', talos: 'http://h:8081' });
	});

	it('returns empty for undefined', () => {
		expect(parseServices(undefined)).toEqual({});
	});
});
