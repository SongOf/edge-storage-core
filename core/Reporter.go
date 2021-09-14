package core

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Reporter interface {
	Collectors() []prometheus.Collector
}
