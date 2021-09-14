package eslog

import (
	"testing"
)

func TestLumberjack(t *testing.T) {
	option := LogOption{
		LogFileName: "/tmp/test.log",
		MaxSize:     1,
		//Environment: "Production",
		Environment: "Development",
	}
	Init(option)

	for i := 0; i < 100000; i++ {
		L().Info("test info", Field("info", i))
		L().Error("graceful restart error", Field("Error", i))
	}
}
