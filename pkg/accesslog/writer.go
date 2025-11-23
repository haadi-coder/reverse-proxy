package accesslog

import "net/http"

type ResponseWriter struct {
	http.ResponseWriter
	ContentLength int
	StatusCode    int
}

func (rw *ResponseWriter) Write(data []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(data)
	rw.ContentLength += n

	return n, err
}

func (rw *ResponseWriter) WriteHeader(code int) {
	rw.StatusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
