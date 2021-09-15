package dashboard

import (
	"github.com/SongOf/edge-storage-core/core"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"net/http/pprof"
)

type MountPoint struct {
	Path    string
	Handler http.Handler
}

type Panel struct {
	MountPoints []*MountPoint
}

func (mountPoint *MountPoint) Mount(mux *http.ServeMux) {
	mux.Handle(mountPoint.Path, mountPoint.Handler)
}

func NewPromPanel(reporters ...core.Reporter) *Panel {
	mountPoints := make([]*MountPoint, 0)

	for _, reporter := range reporters {
		prometheus.MustRegister(reporter.Collectors()...)
	}
	mountPoints = append(mountPoints, &MountPoint{
		Path:    "/metrics",
		Handler: promhttp.Handler(),
	})
	return &Panel{
		MountPoints: mountPoints,
	}
}

// NewPprofPanel create a Panel for pprof
// net/http/pprof init() register handler in default mux
// if we want to use pprof in new mux
// we have to follow what happens in pprof.init()
func NewPprofPanel() *Panel {
	surfaces := make([]*MountPoint, 0)

	surfaces = append(surfaces, &MountPoint{
		Path:    "/debug/pprof/",
		Handler: http.HandlerFunc(pprof.Index),
	})
	surfaces = append(surfaces, &MountPoint{
		Path:    "/debug/pprof/cmdline",
		Handler: http.HandlerFunc(pprof.Cmdline),
	})
	surfaces = append(surfaces, &MountPoint{
		Path:    "/debug/pprof/profile",
		Handler: http.HandlerFunc(pprof.Profile),
	})
	surfaces = append(surfaces, &MountPoint{
		Path:    "/debug/pprof/symbol",
		Handler: http.HandlerFunc(pprof.Symbol),
	})
	surfaces = append(surfaces, &MountPoint{
		Path:    "/debug/pprof/trace",
		Handler: http.HandlerFunc(pprof.Trace),
	})
	return &Panel{MountPoints: surfaces}
}
