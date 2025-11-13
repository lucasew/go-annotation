package annotation

import (
	"log"
	"net/http"
	"time"
)

// i18nMiddleware adds the appropriate localizer to the request context
func i18nMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		localizer := GetLocalizerFromRequest(r)
		ctx := WithLocalizer(r.Context(), localizer)
		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}

func HTTPLogger(handler http.Handler) http.Handler {
    return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
        initialTime := time.Now()
        method := r.Method
        path := r.URL.String()
        wr := NewStatusCodeRecorderResponseWriter(w)
        handler.ServeHTTP(wr, r)
        finalTime := time.Now()
        statusCode := wr.Status
        log.Printf("http: time:%dms %d %s %s", finalTime.Sub(initialTime) / time.Millisecond , statusCode, method, path, )
    })
}

type StatusCodeRecorderResponseWriter struct {
    http.ResponseWriter
    Status int
}

func (r *StatusCodeRecorderResponseWriter) WriteHeader(status int) {
    r.Status = status
    r.ResponseWriter.WriteHeader(status)
}

func NewStatusCodeRecorderResponseWriter(w http.ResponseWriter) *StatusCodeRecorderResponseWriter {
    return &StatusCodeRecorderResponseWriter{ResponseWriter: w, Status: 200}
}
