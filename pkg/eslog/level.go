package eslog

import (
	"go.uber.org/zap"
	"log"
)

func (es *EsLogger) Debug(msg string, fields ...zap.Field) {
	if es.logger != nil {
		es.logger.Debug(msg, fields...)
	} else {
		log.Print(msg, fields)
	}
}

// Debugf log a template message with args
func (es *EsLogger) Debugf(template string, args ...interface{}) {
	if es.sugar != nil {
		es.sugar.Debugf(template, args...)
	} else {
		log.Printf(template, args...)
	}
}

func (es *EsLogger) Warn(msg string, fields ...zap.Field) {
	if es.logger != nil {
		es.logger.Warn(msg, fields...)
	} else {
		log.Print(msg, fields)
	}
}

// Warnf log a template message with args
func (es *EsLogger) Warnf(template string, args ...interface{}) {
	if es.sugar != nil {
		es.sugar.Warnf(template, args...)
	} else {
		log.Printf(template, args...)
	}
}

func (es *EsLogger) Info(msg string, fields ...zap.Field) {
	if es.logger != nil {
		es.logger.Info(msg, fields...)
	} else {
		log.Print(msg, fields)
	}
}

// Infof log a template message with args
func (es *EsLogger) Infof(template string, args ...interface{}) {
	if es.sugar != nil {
		es.sugar.Infof(template, args...)
	} else {
		log.Printf(template, args...)
	}
}

func (es *EsLogger) DPanic(msg string, fields ...zap.Field) {
	if es.logger != nil {
		es.logger.DPanic(msg, fields...)
	} else {
		log.Print(msg, fields)
		panic(msg)
	}
}

// DPanicf log a template message with args
func (es *EsLogger) DPanicf(template string, args ...interface{}) {
	if es.sugar != nil {
		es.sugar.DPanicf(template, args...)
	} else {
		log.Printf(template, args...)
	}
}

func (es *EsLogger) Panic(msg string, fields ...zap.Field) {
	if es.logger != nil {
		es.logger.Panic(msg, fields...)
	} else {
		log.Print(msg, fields)
		panic(msg)
	}
}

// Panicf log a template message with args
func (es *EsLogger) Panicf(template string, args ...interface{}) {
	if es.sugar != nil {
		es.sugar.Panicf(template, args...)
	} else {
		log.Printf(template, args...)
	}
}

func (es *EsLogger) Error(msg string, fields ...zap.Field) {
	if es.logger != nil {
		es.logger.Error(msg, fields...)
	} else {
		log.Print(msg, fields)
	}
}

// Errorf log a template message with args
func (es *EsLogger) Errorf(template string, args ...interface{}) {
	if es.sugar != nil {
		es.sugar.Errorf(template, args...)
	} else {
		log.Printf(template, args...)
	}
}

func (es *EsLogger) Fatal(msg string, fields ...zap.Field) {
	if es.logger != nil {
		es.logger.Fatal(msg, fields...)
	} else {
		log.Print(msg, fields)
	}
}

// Fetalf log a template message with args
func (es *EsLogger) Fetalf(template string, args ...interface{}) {
	if es.sugar != nil {
		es.sugar.Fatalf(template, args...)
	} else {
		log.Printf(template, args...)
	}
}
