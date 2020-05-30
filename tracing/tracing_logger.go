package tracing

import (
	"github.com/DoubleChins/framework/cache"
	"github.com/golang/glog"
	"net/http"
)

var c = cache.NewMemoryPointerCache()

type Trace struct {
	traceId  string
	parentId string
}

func LogWithTrace(r *http.Request, format string, arguments ...interface{}) {
	// check cache for trace ids or compute new trace id
	trace := c.GetOrCompute(r, func(v interface{}) interface{} {
		return parseTraceID(v.(*http.Request))
	}).(*Trace)

	// move all arguments down to the end so we have the first two slots in array free
	arguments = append(arguments, nil, nil)
	for i := len(arguments) - 3; i >= 0; i-- {
		arguments[i+2], arguments[i] = arguments[i], arguments[i+2]
	}

	// Set trace and parent id in 0 and 1 index
	arguments[0] = trace.traceId
	arguments[1] = trace.parentId

	// Print to log
	glog.Infof("[T:%s] [P:%s] "+format+"\n", arguments...)
}

func parseTraceID(r *http.Request) *Trace {
	// Quick checks
	headerValue := r.Header.Get(HeaderKey)
	if len(headerValue) != 55 {
		return &Trace{
			traceId:  "ERR",
			parentId: "ERR",
		}
	}

	// Get trace and parent id from header
	return &Trace{
		traceId:  headerValue[3:35],
		parentId: headerValue[36:52],
	}
}
