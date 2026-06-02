// permissionMatches mirrors the backend matcher: a bare "*" matches everything;
// otherwise both pattern and action split on their LAST "." into
// [resourceType, verb] (resourceType may contain ":", e.g. "app:aegis.read"),
// and each segment matches iff it is "*" or equal.
export function permissionMatches(pattern: string, action: string): boolean {
	if (pattern === '*') return true;
	const [pRes, pVerb] = splitLast(pattern);
	const [aRes, aVerb] = splitLast(action);
	return segmentMatches(pRes, aRes) && segmentMatches(pVerb, aVerb);
}

function splitLast(s: string): [string, string] {
	const i = s.lastIndexOf('.');
	if (i < 0) return [s, ''];
	return [s.slice(0, i), s.slice(i + 1)];
}

function segmentMatches(pattern: string, value: string): boolean {
	return pattern === '*' || pattern === value;
}

export function can(granted: string[], action: string): boolean {
	return granted.some((p) => permissionMatches(p, action));
}
