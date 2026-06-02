package app

import "testing"

func TestPermissionMatches(t *testing.T) {
	cases := []struct {
		name    string
		pattern string
		action  string
		want    bool
	}{
		{"bare star matches everything", "*", "users.read", true},
		{"star dot star matches everything", "*.*", "app:aegis.read", true},
		{"star dot star matches plain rt", "*.*", "users.write", true},
		{"star dot star matches colon write", "*.*", "app:aegis.write", true},
		{"exact match", "users.read", "users.read", true},
		{"exact mismatch verb", "users.read", "users.write", false},
		{"exact mismatch rt", "users.read", "roles.read", false},
		{"verb wildcard matches", "users.*", "users.write", true},
		{"verb wildcard mismatch rt", "users.*", "roles.write", false},
		{"rt wildcard matches colon rt", "*.read", "app:aegis.read", true},
		{"rt wildcard mismatch verb", "*.read", "app:aegis.write", false},
		{"colon rt exact match", "app:aegis.read", "app:aegis.read", true},
		{"colon rt mismatch slug", "app:aegis.read", "app:forge.read", false},
		{"colon rt verb wildcard", "app:aegis.*", "app:aegis.write", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := permissionMatches(tc.pattern, tc.action); got != tc.want {
				t.Fatalf("permissionMatches(%q, %q) = %v, want %v", tc.pattern, tc.action, got, tc.want)
			}
		})
	}
}

func TestSplitPermission(t *testing.T) {
	cases := []struct {
		in   string
		rt   string
		verb string
	}{
		{"users.read", "users", "read"},
		{"app:aegis.read", "app:aegis", "read"},
		{"*.*", "*", "*"},
		{"*", "*", ""},
	}
	for _, tc := range cases {
		rt, verb := splitPermission(tc.in)
		if rt != tc.rt || verb != tc.verb {
			t.Fatalf("splitPermission(%q) = (%q, %q), want (%q, %q)", tc.in, rt, verb, tc.rt, tc.verb)
		}
	}
}
