package eslog

import (
	"testing"
)

func TestInfof(t *testing.T) {
	option := LogOption{
		LogFileName: "/tmp/test.log",
		MaxSize:     1,
		Environment: "Development",
	}
	Init(option)
	L().Infof("p1 %s, p2 %s", "a", "b")
}
