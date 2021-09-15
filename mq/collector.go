package mq

import "github.com/prometheus/client_golang/prometheus"

var (
	defaultCollector = NewCollector()
)

func DefaultCollector() prometheus.Collector {
	return defaultCollector
}

func NewCollector() *Collector {
	errorCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "escore_mq_error_total",
		Help: "escore mq error total count",
	})
	return &Collector{
		ErrorCounter: errorCounter,
	}
}

func ErrorInc() {
	defaultCollector.ErrorInc()
}

type Collector struct {
	ErrorCounter prometheus.Counter
}

func (collector *Collector) ErrorInc() {
	collector.ErrorCounter.Inc()
}

func (collector *Collector) Collect(ch chan<- prometheus.Metric) {
	collector.ErrorCounter.Collect(ch)
}

func (collector *Collector) Describe(ch chan<- *prometheus.Desc) {
	collector.ErrorCounter.Describe(ch)
}
