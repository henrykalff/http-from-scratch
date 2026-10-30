package metrics

import (
	"io"

	"github.com/VictoriaMetrics/metrics"
)

var RequestsTotal *metrics.Counter
var RequestProcessingDuration *metrics.Summary

func InitMetrics() {
	RequestsTotal = metrics.NewCounter("requests_total")
	RequestProcessingDuration = metrics.NewSummary("requests_processing_time_ms")

	// metrics.InitPush("http://vmagent-vm-stack-victoria-metrics-k8s-stack.monitoring.svc.cluster.local:8429/api/v1/import/prometheus", 10*time.Second, `instance="http1"`, false)

}

func WriteMetrics(w io.Writer) {
	metrics.WritePrometheus(w, false)
}
