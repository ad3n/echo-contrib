package helpers

import (
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/ad3n/echo/v5"
)

func TestDefaultStatusResolver(t *testing.T) {
	for _, tc := range []struct {
		name      string
		err       error
		committed int
		want      int
	}{
		{name: "uncommitted"},
		{name: "committed", committed: 201, want: 201},
		{name: "plain error", err: errors.New("failure"), want: 500},
		{name: "committed error", err: errors.New("failure"), committed: 202, want: 202},
		{name: "HTTP error", err: echo.NewHTTPError(400, "failure"), committed: 201, want: 400},
		{name: "wrapped HTTP error", err: fmt.Errorf("wrapped: %w", echo.NewHTTPError(403, "failure")), want: 403},
		{name: "joined HTTP error", err: errors.Join(errors.New("failure"), echo.NewHTTPError(404, "failure")), want: 404},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := echo.New().NewContext(httptest.NewRequest("GET", "/", nil), httptest.NewRecorder())
			if tc.committed != 0 {
				c.Response().WriteHeader(tc.committed)
			}

			if got := DefaultStatusResolver(c, tc.err); got != tc.want {
				t.Fatalf("status = %d, want %d", got, tc.want)
			}
		})
	}
}

func BenchmarkDefaultStatusResolver(b *testing.B) {
	for _, tc := range []struct {
		name string
		err  error
	}{
		{name: "success"},
		{name: "HTTPError", err: echo.NewHTTPError(400, "failure")},
		{name: "wrapped", err: fmt.Errorf("wrapped: %w", echo.NewHTTPError(403, "failure"))},
	} {
		b.Run(tc.name, func(b *testing.B) {
			c := echo.New().NewContext(httptest.NewRequest("GET", "/", nil), httptest.NewRecorder())
			c.Response().WriteHeader(200)
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				DefaultStatusResolver(c, tc.err)
			}
		})
	}
}
