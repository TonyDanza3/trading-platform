package internal

import "strings"

type Route struct {
	Prefix string
	Target string
}

func matchRoute(path string, routes []Route) (Route, bool) {
	best := Route{}
	found := false
	for _, route := range routes {
		if path == route.Prefix || strings.HasPrefix(path, route.Prefix+"/") {
			if !found || len(route.Prefix) > len(best.Prefix) {
				best = route
				found = true
			}
		}
	}
	return best, found
}
