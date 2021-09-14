package storage

import "github.com/prometheus/client_golang/prometheus"

var (
	defaultCollector = NewCollector()
)

func DefaultCollector() prometheus.Collector {
	return defaultCollector
}

func NewCollector() *Collector {
	databaseErrorCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "escore_database_error_total",
		Help: "escore database error total count",
	})

	cacheErrorCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "escore_cache_error_total",
		Help: "escore cache error total count",
	})

	return &Collector{
		DatabaseErrorCounter: databaseErrorCounter,
		CacheErrorCounter:    cacheErrorCounter,
	}
}

func DatabaseErrorInc() {
	defaultCollector.DatabaseErrorInc()
}

func CacheErrorInc() {
	defaultCollector.CacheErrorInc()
}

type Collector struct {
	DatabaseErrorCounter prometheus.Counter
	CacheErrorCounter    prometheus.Counter
}

func (collector *Collector) CacheErrorInc() {
	collector.CacheErrorCounter.Inc()
}

func (collector *Collector) DatabaseErrorInc() {
	collector.DatabaseErrorCounter.Inc()
}

func (collector *Collector) Collect(ch chan<- prometheus.Metric) {
	collector.DatabaseErrorCounter.Collect(ch)
	collector.CacheErrorCounter.Collect(ch)
}

func (collector *Collector) Describe(ch chan<- *prometheus.Desc) {
	collector.DatabaseErrorCounter.Describe(ch)
	collector.CacheErrorCounter.Describe(ch)
}
