package tracing

import (
	"bytes"
	"net/http"
)

type TraceableResponse struct {
	http.ResponseWriter

	Original     http.ResponseWriter
	ResponseCopy bytes.Buffer
	Status       int
}

func (tr *TraceableResponse) Header() http.Header {
	return tr.Original.Header()
}

func (tr *TraceableResponse) Write(data []byte) (int, error) {
	if tr.Status == 0 {
		tr.Status = http.StatusOK
	}

	n, err := tr.Original.Write(data)
	tr.ResponseCopy.Write(data)
	return n, err
}

func (tr *TraceableResponse) WriteHeader(statusCode int) {
	tr.Original.WriteHeader(statusCode)
	tr.Status = statusCode
}
