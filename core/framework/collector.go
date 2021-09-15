package framework

import "github.com/prometheus/client_golang/prometheus"

type ServerCollector struct {
	ServerPanicCounter               prometheus.Counter
	ControllerErrorCounterVector     *prometheus.CounterVec
	ControllerLatencyHistogramVector *prometheus.HistogramVec
}

func NewCollector() *ServerCollector {
	panicCounter := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "escore_server_panic_total",
			Help: "es-core server panic total count",
		})
	errorCounterVec := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "escore_server_error_total",
			Help: "es-core server controller error count.",
		},
		[]string{"action"})
	latencyHistogramVec := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "escore_server_latency_seconds",
			Help:    "es-core server controller latency seconds.",
			Buckets: []float64{0, 500, 1000, 5000},
		},
		[]string{"action"})

	return &ServerCollector{
		ServerPanicCounter:               panicCounter,
		ControllerErrorCounterVector:     errorCounterVec,
		ControllerLatencyHistogramVector: latencyHistogramVec,
	}
}

func (collector *ServerCollector) Collect(ch chan<- prometheus.Metric) {
	collector.ServerPanicCounter.Collect(ch)
	collector.ControllerErrorCounterVector.Collect(ch)
	collector.ControllerLatencyHistogramVector.Collect(ch)
}

func (collector *ServerCollector) Describe(ch chan<- *prometheus.Desc) {
	collector.ServerPanicCounter.Describe(ch)
	collector.ControllerErrorCounterVector.Describe(ch)
	collector.ControllerLatencyHistogramVector.Describe(ch)
}
