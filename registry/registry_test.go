package registry

import (
	"api_gateway/config"
	"api_gateway/route"
	"testing"
)

func RoutesAreEqual(expected, actual route.Route) bool {
	c1 := expected.Method == actual.Method
	c2 := expected.Prefix == actual.Prefix
	c3 := expected.ProxyTarget == actual.ProxyTarget
	c4 := expected.RedirectTarget == actual.RedirectTarget
	c5 := expected.FixedPath == actual.FixedPath
	return c1 && c2 && c3 && c4 && c5
}

func DomainRoutesAreEqual(expected, actual route.DomainRoute) bool {
	c1 := expected.Domain == actual.Domain
	c2 := expected.ProxyTarget == actual.ProxyTarget
	return c1 && c2
}

func TestRouteParsing(t *testing.T) {
	cfg, err := config.LoadConfig("./route_config.yaml", "yaml")
	if err != nil {
		t.Logf("error: %v", err)
	}
	rr := &RouteRegistry{}
	rr.FromConfig(cfg)

	route1 := route.Route{
		Prefix:       "/foo",
		RelativePath: "/foo/*path",
		Method:       "POST",
		ProxyTarget:  "https://bar.com",
	}

	route2 := route.Route{
		Prefix:       "/foo",
		RelativePath: "/foo/docs/todos/*path",
		Method:       "GET",
		ProxyTarget:  "https://bar.com",
		FixedPath:    "/docs/todos",
	}

	route3 := route.Route{
		Prefix:       "/foo",
		RelativePath: "/foo/docs/templates/*path",
		Method:       "PUT",
		ProxyTarget:  "https://bar.com",
		FixedPath:    "/docs/templates",
	}

	route4 := route.Route{
		Prefix:         "/foobar",
		RedirectTarget: "https://xyzzy.com",
	}

	route5 := route.Route{
		Prefix:         "/thud",
		RelativePath:   "/thud/foo",
		RedirectTarget: "https://foo.com",
		FixedPath:      "/foo",
	}

	route6 := route.Route{
		Prefix:         "/thud",
		RelativePath:   "/thud/bar",
		RedirectTarget: "https://bar.com",
		FixedPath:      "/bar",
	}

	domainRoute1 := route.DomainRoute{
		Domain:      "www.example.com",
		ProxyTarget: "https://dummy.com",
	}

	domainRoute2 := route.DomainRoute{
		Domain:      "www.test.com",
		ProxyTarget: "https://tower.com",
	}

	expectedRoutes := []route.Route{route1, route2, route3, route4, route5, route6}

	expectedDomainRoutes := []route.DomainRoute{domainRoute1, domainRoute2}

	for idx, expected := range expectedRoutes {
		actual := rr.Routes[idx]
		if !RoutesAreEqual(expected, actual) {
			t.Errorf("Structs are not equal:\nExpected: %+v\nActual: %+v", expected, actual)
		}
	}

	for idx, expected := range expectedDomainRoutes {
		actual := rr.DomainRoutes[idx]
		if !DomainRoutesAreEqual(expected, actual) {
			t.Errorf("Structs are not equal:\nExpected: %+v\nActual: %+v", expected, actual)
		}
	}
}

func TestMiddlewareParsing(t *testing.T) {
	cfg, _ := config.LoadConfig("./route_config.yaml", "yaml")

	rr := &RouteRegistry{}
	rr.FromConfig(cfg)

	expectedRouteMiddleware := [8]int{1, 2, 1, 0, 0, 0}

	expectedDomainMiddleware := [2]int{1, 2}

	for idx, r := range rr.Routes {
		if len(r.Middleware) != expectedRouteMiddleware[idx] {
			t.Errorf("middleware number is not correct:\nExpected: %+v\nActual: %+v", len(r.Middleware), expectedRouteMiddleware[idx])
		}
	}

	for idx, dr := range rr.DomainRoutes {
		if len(dr.Middleware) != expectedDomainMiddleware[idx] {
			t.Errorf("middleware number is not correct:\nExpected: %+v\nActual: %+v", len(dr.Middleware), expectedDomainMiddleware[idx])
		}

	}

	// for idx, expected := range expectedMiddlewareNumbers {
	// if len(rr.Routes[idx].Middleware) != expected {
	//
	// }
	// }

	// fw1 := config.RateLimitConfig{
	// 	Ttl:             8 * time.Hour,
	// 	CleanupInterval: 2 * time.Hour,
	// 	Algorithm:       "fixed_window_counter",
	// 	Limit:           10,
	// 	WindowSize:      5 * time.Second,
	// }

	// tb1 := config.RateLimitConfig{
	// 	Ttl:             8 * time.Hour,
	// 	CleanupInterval: 2 * time.Hour,
	// 	Algorithm:       "token_bucket",
	// 	Capacity:        10,
	// 	RefillTokens:    10,
	// 	RefillInterval:  5 * time.Second,
	// }

	// algo, rl := ParseRateLimitCfg(tb1)

	// handler := middleware.NewRateLimitMiddleware(algo, rl)
}
