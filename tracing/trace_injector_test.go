package tracing

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

var (
	dummyRequest = &http.Request{
		Header: http.Header{},
	}
)

func init() {
	dummyRequest.Header.Set(HeaderKey, generateNewTraceParent())

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("OK"))
	})
	go http.ListenAndServe(":9999", TraceMiddleware(http.DefaultServeMux))
}

func BenchmarkGenerateNewTraceParent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generateNewTraceParent()
	}
}

func BenchmarkLogWithTrace(b *testing.B) {
	for i := 0; i < b.N; i++ {
		LogWithTrace(dummyRequest, "%s", "Test")
	}
}

func TestLogWithTrace(t *testing.T) {
	r := &http.Request{
		Header: http.Header{},
	}
	r.Header.Set(HeaderKey, generateNewTraceParent())
	LogWithTrace(r, "%s", "test")
}

func TestAll(t *testing.T) {
	buf := bytes.NewBufferString("TEST")
	req, err := http.NewRequest("POST", "http://localhost:9999/", buf)
	if err != nil {
		t.Fatal(err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		t.Fatal(err)
		return
	}

	fmt.Printf("Response: %v\n", data)
}
