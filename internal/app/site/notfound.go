package site

import (
	"bufio"
	"net"
	"net/http"
)

// WithNotFound wraps an http.Handler so plain mux 404 responses render the
// site not-found page (avoids registering GET /{path...} which conflicts
// with the static file prefix).
func (f *Feature) WithNotFound(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(rec, r)
		if rec.code == http.StatusNotFound && !rec.hijacked {
			f.handleNotFound(w, r)
		}
	})
}

type statusRecorder struct {
	http.ResponseWriter
	code     int
	wrote    bool
	hijacked bool
}

func (r *statusRecorder) WriteHeader(code int) {
	r.code = code
	if code == http.StatusNotFound {
		return
	}
	r.wrote = true
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.code == http.StatusNotFound {
		return len(b), nil
	}
	if !r.wrote {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(b)
}

func (r *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	r.hijacked = true
	return h.Hijack()
}

func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

var _ http.ResponseWriter = (*statusRecorder)(nil)
