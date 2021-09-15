package recovery

import (
	"github.com/SongOf/edge-storage-core/core"
	"github.com/SongOf/edge-storage-core/pkg/eslog"
	"runtime/debug"
)

type Delegate func(*core.Context) error
type PanicHandler func()

func Recover(ctx *core.Context, handler func()) {
	if r := recover(); r != nil {
		eslog.C(ctx).Error("Recover from panic", eslog.Field("Panic", r), eslog.Field("Stack", string(debug.Stack())))
		if handler != nil {
			handler()
		}
	}
}

func Process(ctx *core.Context, delegate Delegate, panicHandler PanicHandler) error {
	defer Recover(ctx, panicHandler)
	return delegate(ctx)
}
