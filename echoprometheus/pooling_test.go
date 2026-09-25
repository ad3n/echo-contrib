package echoprometheus

import (
	"fmt"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/ad3n/echo/v5"
	"github.com/prometheus/client_golang/prometheus"
)

func BenchmarkMiddlewareLabels(b *testing.B) {
	for _, custom := range []bool{false, true} {
		b.Run(fmt.Sprintf("custom=%t", custom), func(b *testing.B) {
			conf := MiddlewareConfig{Registerer: prometheus.NewRegistry()}
			if custom {
				conf.LabelFuncs = map[string]LabelValueFunc{"tenant": func(c *echo.Context, _ error) string { return c.Request().Header.Get("Tenant") }}
			}

			handler := NewMiddlewareWithConfig(conf)(func(c *echo.Context) error { return nil })
			c := echo.New().NewContext(httptest.NewRequest("GET", "/", nil), httptest.NewRecorder())
			c.Request().Header.Set("Tenant", "one")
			if err := handler(c); err != nil {
				b.Fatal(err)
			}

			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				if err := handler(c); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func TestMiddlewareConcurrentLabels(t *testing.T) {
	registry := prometheus.NewRegistry()
	handler := NewMiddlewareWithConfig(MiddlewareConfig{
		Registerer: registry,
		LabelFuncs: map[string]LabelValueFunc{
			"tenant": func(c *echo.Context, _ error) string { return c.Request().Header.Get("Tenant") },
		},
	})(func(c *echo.Context) error { return nil })
	var wg sync.WaitGroup
	for i := range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()

			c := echo.New().NewContext(httptest.NewRequest("GET", "/", nil), httptest.NewRecorder())
			c.Request().Header.Set("Tenant", fmt.Sprint(i))
			for range 100 {
				if err := handler(c); err != nil {
					t.Error(err)
					return
				}
			}
		}()
	}

	wg.Wait()
	families, err := registry.Gather()
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, family := range families {
		if family.GetName() != "echo_requests_total" {
			continue
		}

		found = true
		tenants := make(map[string]bool)
		for _, metric := range family.Metric {
			if metric.GetCounter().GetValue() != 100 {
				t.Fatalf("counter = %v, want 100", metric.GetCounter().GetValue())
			}

			for _, label := range metric.Label {
				if label.GetName() == "tenant" {
					tenants[label.GetValue()] = true
				}
			}
		}

		if len(tenants) != 16 {
			t.Fatalf("tenants = %d, want 16", len(tenants))
		}
	}

	if !found {
		t.Fatal("request counter missing")
	}
}
