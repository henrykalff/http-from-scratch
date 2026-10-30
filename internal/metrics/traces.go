package metrics

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

var Tracer = otel.Tracer("http1server")

func InitTracer(ctx context.Context) (*sdktrace.TracerProvider, error) {
	// VictoriaTraces OTLP HTTP receiver on port 10428
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint("vtsingle-vm-stack-victoria-metrics-k8s-stack.monitoring.svc.cluster.local:10428"),
		otlptracehttp.WithInsecure(),
		// Standard OTLP path accepted by VictoriaTraces
		otlptracehttp.WithURLPath("/insert/opentelemetry/v1/traces"),
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("http1server"),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(1*time.Second)),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

func NewID(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}

type TraceEvent struct {
	Msg        string `json:"_msg"`
	Service    string `json:"service"`
	TraceID    string `json:"trace_id"`
	SpanID     string `json:"span_id"`
	ParentID   string `json:"parent_id,omitempty"` // empty if it's the root event
	SpanName   string `json:"span_name"`
	DurationMs int64  `json:"duration_ms"`
	CustomTag  string `json:"custom_tag,omitempty"`
}

func SendTraceEvent(event TraceEvent) {
	body, _ := json.Marshal(event)
	url := "http://vlsingle-vm-stack-victoria-metrics-k8s-stack.monitoring.svc.cluster.local:9428/insert/jsonline"

	go func() {
		_, _ = http.Post(url, "application/json", bytes.NewReader(body))
	}()
}
