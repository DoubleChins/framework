//
// This implements tracing according to https://www.w3.org/TR/trace-context/
//
package tracing

import (
	"bytes"
	"encoding/hex"
	"github.com/DoubleChins/framework/logging"
	"github.com/DoubleChins/framework/util"
	"github.com/golang/glog"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

const HeaderKey = "traceparent"

type TraceHandler struct {
	http.Handler

	Original http.Handler
}

func (handler *TraceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Store current header value
	currentHeader := r.Header.Get(HeaderKey)
	if currentHeader == "" {
		currentHeader = generateNewTraceParent()
		if currentHeader != "" {
			r.Header.Set(HeaderKey, currentHeader)
		}
	}

	var response http.ResponseWriter

	// Handle the request
	if glog.V(logging.TraceDebug) {
		// Log request
		logRequest(r)

		// Create traceable response
		response = &TraceableResponse{
			Original: w,
		}
	} else {
		response = w
	}

	// Handle the original request
	handler.Original.ServeHTTP(response, r)

	// Log response
	if glog.V(logging.TraceDebug) {
		logResponse(r, response.(*TraceableResponse))
	}

	// Invalidate caching
	c.Delete(r)
}

func logResponse(r *http.Request, w *TraceableResponse) {
	// Indicate what we dump
	LogWithTrace(r, "--- RESPONSE")
	LogWithTrace(r, "")

	// Log status
	LogWithTrace(r, "%d", w.Status)

	// Log body
	LogWithTrace(r, "")

	dump := hex.Dump(w.ResponseCopy.Bytes())
	lines := strings.Split(dump, "\n")
	for _, line := range lines {
		LogWithTrace(r, line)
	}
}

func logRequest(r *http.Request) {
	// Indicate what we dump
	LogWithTrace(r, "--- REQUEST")
	LogWithTrace(r, "")

	// Log request line
	LogWithTrace(r, "%s %s %s", r.Method, r.RequestURI, r.Proto)
	LogWithTrace(r, "")

	// Log Headers
	for key, values := range r.Header {
		for _, value := range values {
			LogWithTrace(r, "%s: %s", key, value)
		}
	}

	// Copy body
	var err error
	var copyBody io.ReadCloser

	copyBody, r.Body, err = drainBody(r.Body)
	if err != nil {
		LogWithTrace(r, "can't copy body: %v", err)
	} else {
		buf, err := ioutil.ReadAll(copyBody)
		defer copyBody.Close()

		if err != nil {
			LogWithTrace(r, "can't read body: %v", err)
		} else {
			// Log body
			dump := hex.Dump(buf)
			LogWithTrace(r, "")

			lines := strings.Split(dump, "\n")
			for _, line := range lines {
				LogWithTrace(r, line)
			}
		}
	}
}

func drainBody(b io.ReadCloser) (r1, r2 io.ReadCloser, err error) {
	if b == http.NoBody {
		// No copying needed. Preserve the magic sentinel meaning of NoBody.
		return http.NoBody, http.NoBody, nil
	}

	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, b, err
	}

	if err = b.Close(); err != nil {
		return nil, b, err
	}

	return ioutil.NopCloser(&buf), ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

func generateNewTraceParent() string {
	// Generate new trace context id
	id, err := uuid.NewUUID()
	if err != nil {
		glog.V(logging.TraceInformation).Infof("could not generate uuid for tracing: %v", err)
		return ""
	}

	// Hex the binary form
	idData, err := id.MarshalBinary()
	if err != nil {
		glog.V(logging.TraceInformation).Infof("could not marshal uuid for tracing: %v", err)
		return ""
	}

	// Generate new parent id
	parentId := generateNewParent()

	// Now return new traceparent id
	return "00-" + hex.EncodeToString(idData) + "-" + parentId + "-00"
}

func generateNewParent() string {
	randomString := util.RandomString(16, []byte(util.HEX))
	return randomString
}

func TraceMiddleware(handler http.Handler) http.Handler {
	// Return trace wrapped handler
	return &TraceHandler{
		Original: handler,
	}
}
