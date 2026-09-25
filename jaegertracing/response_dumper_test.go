package jaegertracing

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type partialResponseWriter struct {
	http.ResponseWriter
	err error
	n   int
}

func (w partialResponseWriter) Write([]byte) (int, error) {
	return w.n, w.err
}

func TestResponseDumperWrite(t *testing.T) {
	failure := errors.New("write failed")
	for _, tc := range []struct {
		name     string
		err      error
		wantErr  error
		n        int
		wantBody string
	}{
		{name: "success", n: 4, wantBody: "body"},
		{name: "short write", n: 2, wantErr: io.ErrShortWrite},
		{name: "error", n: 2, err: failure, wantErr: failure},
		{name: "full write with error", n: 4, err: failure, wantErr: failure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			writer := partialResponseWriter{ResponseWriter: httptest.NewRecorder(), n: tc.n, err: tc.err}
			d := newResponseDumper(writer)
			n, err := d.Write([]byte("body"))
			if n != tc.n || err != tc.wantErr || d.GetResponse() != tc.wantBody {
				t.Fatalf("Write = (%d, %v), body = %q", n, err, d.GetResponse())
			}

			if d.Unwrap() != writer {
				t.Fatal("underlying writer changed")
			}
		})
	}
}

func BenchmarkResponseDumper(b *testing.B) {
	writer := partialResponseWriter{ResponseWriter: httptest.NewRecorder(), n: 4}
	body := []byte("body")
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		d := newResponseDumper(writer)
		if _, err := d.Write(body); err != nil {
			b.Fatal(err)
		}

		_ = d.GetResponse()
	}
}
