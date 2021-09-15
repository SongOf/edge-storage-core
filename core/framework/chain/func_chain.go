package chain

import (
	"github.com/SongOf/edge-storage-core/core"
	"github.com/SongOf/edge-storage-core/core/eserrors"
	"github.com/SongOf/edge-storage-core/core/recovery"
	"github.com/SongOf/edge-storage-core/pkg/eslog"
	"reflect"
	"runtime"
	"strings"
)

type Function func(*core.Context) error

type Unit struct {
	ForwardFunc  Function
	RollbackFunc Function
}

func NewUnit(forward, rollback Function) Unit {
	return Unit{
		ForwardFunc:  forward,
		RollbackFunc: rollback,
	}
}

type FunctionChain []Unit

func NewChain(units ...Unit) FunctionChain {
	chain := FunctionChain{}
	return chain.AddUnitList(units...)
}

func (chain FunctionChain) AddUnitList(units ...Unit) FunctionChain {
	ch := chain
	for _, unit := range units {
		ch = ch.AddUnit(unit)
	}
	return ch
}

func (chain FunctionChain) AddUnit(unit Unit) FunctionChain {
	if unit.ForwardFunc == nil && unit.RollbackFunc == nil {
		return chain
	}
	return append(chain, unit)
}

func (chain FunctionChain) Run(ctx *core.Context) error {
	chainLength := len(chain)

	if chainLength == 0 {
		eslog.C(ctx).Info("Function chain length is zero, no need to execute.")
		return nil
	}

	var failedIndex = 0
	var failedError error
	for index, unit := range chain {
		// forward
		if unit.ForwardFunc != nil {
			err := runUnitFunc(ctx, unit.ForwardFunc)
			if err != nil {
				failedIndex, failedError = index, err
				eslog.C(ctx).Error("Run forward function failed.",
					eslog.Field("Function", GetFunctionName(unit.ForwardFunc)),
					eslog.Err(err))
				break
			}
		}
	}

	if failedError != nil {
		// rollback
		for i := failedIndex - 1; i >= 0; i-- {
			unit := chain[i]
			if unit.RollbackFunc != nil {
				// ignore rollback error
				rollbackErr := runUnitFunc(ctx, unit.RollbackFunc)

				if rollbackErr != nil {
					eslog.C(ctx).Error("Run rollback function failed.",
						eslog.Field("Function", GetFunctionName(unit.RollbackFunc)),
						eslog.Err(rollbackErr))
				} else {
					eslog.C(ctx).Infof(
						"Run rollback function success. [%s]", GetFunctionName(unit.RollbackFunc))
				}
			}
		}
		eslog.C(ctx).Infof("func chain finish with error: %+v", failedError)
		return failedError
	}

	eslog.C(ctx).Info("func chain finish")
	return nil
}

func runUnitFunc(ctx *core.Context, unitFunc Function) (err error) {
	defer recovery.Recover(ctx, func() {
		eslog.C(ctx).Warn("Unit Function raise panic!")
		err = eserrors.InternalError()
	})

	err = unitFunc(ctx)
	return
}

func GetFunctionName(f Function) string {
	name := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	s := strings.Split(name, "/")
	if len(s) > 0 {
		return s[len(s)-1]
	}
	return name
}
