package accesslog

import "net/http"

type AccessLogResponseWriter struct {
	http.ResponseWriter
	ContentLength int
	StatusCode    int
}

func (rw *AccessLogResponseWriter) Write(data []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(data)
	rw.ContentLength += n

	return n, err
}

func (rw *AccessLogResponseWriter) WriteHeader(code int) {
	rw.StatusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
