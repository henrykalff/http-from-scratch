package main

import (
	"context"
	"fmt"
	"httpServer/internal/http1"
	"httpServer/internal/metrics"
	"io"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const addr = ":8080"

func main() {
	ctx := context.Background()
	tp, _ := metrics.InitTracer(ctx)
	defer tp.Shutdown(ctx)

	metrics.InitMetrics()

	http1.HandleFunc("GET /index.html", func(w http.ResponseWriter, r *http.Request) {
		reqCtx, rootSpan := metrics.Tracer.Start(r.Context(), "request-index")
		defer rootSpan.End()

		_, childSpan := metrics.Tracer.Start(reqCtx, "calculate-payload",
			trace.WithAttributes(attribute.String("custom.tag", "sample-work")),
		)

		data := []byte(`<html>
		<head>
		<title>Hello World</title>
		</head>
		<body>
		Hello World
		</body>
		</html>`)

		time.Sleep(1 * time.Second)
		childSpan.End()
		time.Sleep(10 * time.Millisecond)

		w.Header().Set("Server", "nginx/1.15.8")
		w.Header().Set("Date", "Fri, 09 Apr 2021 13:34:03 GMT")
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
		w.Write(data)
	})

	http1.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
		reqStart := time.Now()

		// 1. Common trace ID for this entire request lifecycle
		traceID := metrics.NewID(16)
		rootSpanID := metrics.NewID(8)

		// --- Inner / Child Operation ---
		childSpanID := metrics.NewID(8)
		childStart := time.Now()
		time.Sleep(30 * time.Millisecond) // Simulated database or compute work
		metrics.SendTraceEvent(metrics.TraceEvent{
			Msg:        "Payload calculation fininshed",
			TraceID:    traceID,
			SpanID:     childSpanID,
			ParentID:   rootSpanID,
			Service:    "http1server",
			SpanName:   "calculate-payload",
			DurationMs: int64(time.Since(childStart)),
			CustomTag:  "sample-work",
		})

		fmt.Fprintf(w, "hello")
		metrics.SendTraceEvent(metrics.TraceEvent{
			Msg:        "Finished GET /hello",
			TraceID:    traceID,
			SpanID:     rootSpanID,
			Service:    "http1Server",
			DurationMs: int64(time.Since(reqStart)),
		})

	})
	http1.HandleFunc("PUT /echo", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Body)
		_, _ = io.Copy(w, r.Body)
	})
	http1.HandleFunc("GET /metrics", func(w http.ResponseWriter, req *http.Request) {
		metrics.WriteMetrics(w)
	})
	if err := http1.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}

}
