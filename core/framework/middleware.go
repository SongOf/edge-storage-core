package framework

import (
	"github.com/SongOf/edge-storage-core/core"
	"github.com/SongOf/edge-storage-core/pkg/eslog"
	"time"
)

type latencyMiddleware struct {
	collector *ServerCollector
}

func NewLatencyMiddleware(collector *ServerCollector) core.Middleware {
	return &latencyMiddleware{collector: collector}
}

func (middleware *latencyMiddleware) Run(ctx *core.Context) error {
	trackTime := time.Now()
	err := ctx.Next()
	elapsed := time.Since(trackTime) / time.Millisecond
	middleware.collector.ControllerLatencyHistogramVector.WithLabelValues(ctx.Action).Observe(float64(elapsed))
	return err
}

type resultMiddleware struct {
	controller Controller
	resp       *ServerResponse
	collector  *ServerCollector
}

func NewResultMiddleware(
	controller Controller, resp *ServerResponse, collector *ServerCollector) core.Middleware {

	return &resultMiddleware{controller: controller, resp: resp, collector: collector}
}

func (middleware *resultMiddleware) Run(ctx *core.Context) error {
	result, rawerr := middleware.controller.Entry(ctx)
	resp := middleware.resp
	if rawerr == nil {
		resp.WithResult(result).Reply()
		return ctx.Next()
	} else {
		if middleware.collector != nil {
			// error count ++
			middleware.collector.ControllerErrorCounterVector.WithLabelValues(ctx.Action).Inc()
		}
		resp.WithError(rawerr).Reply()
		return ctx.Next()
	}
}

type loadUserInfoMiddleware struct{}

func NewLoadUserInfoMiddleware() *loadUserInfoMiddleware {
	return &loadUserInfoMiddleware{}
}

func (middleware loadUserInfoMiddleware) Run(ctx *core.Context) error {
	if err := ctx.LoadUserInfo(ctx.Params); err != nil {
		eslog.C(ctx).Error("LoadUserInfo failed", eslog.Err(err))
	}
	return ctx.Next()
}
