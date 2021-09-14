package eslog

import (
	"context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"songof.com/edge-storage-core/core"
)

var esLogger EsLogger

const (
	Production = "Production"
)

const ctxLoggerName = "ctx-logger"
const ctxModelLoggerName = "ctx-model-logger"

type EsLogger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
	option *LogOption
}

type LogOption struct {
	LogFileName       string
	MaxSize           int //megabytes
	KeepTime          int //days
	Environment       string
	DisableStackTrace bool
}

func buildZapCore(option *LogOption) zapcore.Core {
	syncer := zapcore.AddSync(&lumberjack.Logger{
		Filename: option.LogFileName,
		MaxSize:  option.MaxSize,
		MaxAge:   option.KeepTime,
	})

	var cfg zap.Config
	if option.Environment == Production {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	zapCore := zapcore.NewCore(zapcore.NewJSONEncoder(cfg.EncoderConfig), syncer, cfg.Level)
	if option.Environment == Production {
		return zapCore
	} else {
		stdSyncer := zapcore.Lock(os.Stdout)
		multiZapCore := zapcore.NewTee(
			zapCore,
			zapcore.NewCore(zapcore.NewConsoleEncoder(cfg.EncoderConfig), stdSyncer, cfg.Level))
		return multiZapCore
	}
}

func Init(option LogOption) {
	esLogger.logger = zap.New(buildZapCore(&option),
		zap.AddCaller(),
		zap.AddCallerSkip(1))
	if !option.DisableStackTrace {
		esLogger.logger.WithOptions(zap.AddStacktrace(zap.ErrorLevel))
	}
	esLogger.sugar = esLogger.logger.Sugar()
	esLogger.option = &option
	esLogger.Info("logger init success")
}

func Flush() {
	_ = esLogger.logger.Sync()
}

func (es *EsLogger) Raw() *zap.Logger {
	return es.logger
}

func L() *EsLogger {
	return &esLogger
}

func (es *EsLogger) copyWithField(logFields map[string]interface{}) *EsLogger {
	if es.logger == nil {
		return &EsLogger{option: es.option}
	}

	var fields []zap.Field
	for key, value := range logFields {
		fields = append(fields, Field(key, value))
	}
	cloneLogger := es.logger.With(fields...)
	return &EsLogger{
		logger: cloneLogger,
		sugar:  cloneLogger.Sugar(),
	}
}

func C(ctx context.Context) *EsLogger {
	if ctx == nil {
		return &esLogger
	}

	ctxLogger := ctx.Value(ctxLoggerName)
	if ctxLogger == nil {
		if coreCtx := core.Cast(ctx); coreCtx != nil {
			newLogger := esLogger.copyWithField(coreCtx.LogFields)
			coreCtx.Set(ctxLoggerName, newLogger)
			return newLogger
		}
		return &esLogger
	}
	return ctxLogger.(*EsLogger)
}

func Field(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

func Err(err error) zap.Field {
	return Field("Error", err)
}
