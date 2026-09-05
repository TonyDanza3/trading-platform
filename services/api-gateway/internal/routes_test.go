package internal

import "testing"

func TestMatchRoute(t *testing.T) {
	routes := Config{
		AuthURL:   "http://auth:8080",
		UserURL:   "http://user:8080",
		MarketURL: "http://market:8080",
	}.Routes()

	cases := []struct {
		path  string
		want  string
		found bool
	}{
		{path: "/api/v1/auth/login", want: "/api/v1/auth", found: true},
		{path: "/api/v1/users/me", want: "/api/v1/users", found: true},
		{path: "/api/v1/instruments", want: "/api/v1/instruments", found: true},
		{path: "/api/v1/instruments/abc", want: "/api/v1/instruments", found: true},
		{path: "/api/v1/orders", found: false},
		{path: "/health", found: false},
		{path: "/openapi.yaml", found: false},
	}

	for _, tc := range cases {
		got, ok := matchRoute(tc.path, routes)
		if ok != tc.found {
			t.Fatalf("%s: found=%v want %v", tc.path, ok, tc.found)
		}
		if ok && got.Prefix != tc.want {
			t.Fatalf("%s: prefix=%s want %s", tc.path, got.Prefix, tc.want)
		}
	}
}
