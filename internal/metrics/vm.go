package metrics

import (
	"io"
	"time"

	"github.com/VictoriaMetrics/metrics"
)

var RequestsTotal *metrics.Counter

func InitMetrics() {
	RequestsTotal = metrics.NewCounter("requests_total")
	metrics.InitPush("http://vmagent-vm-stack-victoria-metrics-k8s-stack.monitoring.svc.cluster.local:8429/api/v1/import/prometheus", 10*time.Second, `instance="http1"`, false)
}

func WriteMetrics(w io.Writer) {
	metrics.WritePrometheus(w, false)
}
