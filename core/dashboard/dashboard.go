package dashboard

import (
	"context"
	"fmt"
	"github.com/SongOf/edge-storage-core/pkg/eslog"
	"net/http"
	"time"
)

type Dashboard struct {
	option *Option
	server *http.Server
	panels []*Panel
}

type Option struct {
	ListenPort int
}

func NewDashboard(panels []*Panel, option *Option) *Dashboard {
	return &Dashboard{
		panels: panels,
		option: option,
		server: &http.Server{},
	}
}

func (dashboard *Dashboard) Run() error {
	mux := http.NewServeMux()

	for _, panel := range dashboard.panels {
		for _, mountPoint := range panel.MountPoints {
			mountPoint.Mount(mux)
		}
	}

	addr := fmt.Sprintf("0.0.0.0:%d", dashboard.option.ListenPort)

	dashboard.server.Handler = mux
	dashboard.server.Addr = addr

	if err := dashboard.server.ListenAndServe(); err != http.ErrServerClosed {
		eslog.L().Warn("Dashboard server stop serving with error.", eslog.Err(err))
		return err
	}
	eslog.L().Info("Dashboard server stop serving.")
	return nil
}

func (dashboard *Dashboard) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := dashboard.server.Shutdown(ctx); err != nil {
		eslog.L().Warn("Shutdown Dashboard server failed.", eslog.Err(err))
		return err
	} else {
		eslog.L().Info("shutdown Dashboard server success.")
	}
	return nil
}
