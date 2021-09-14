package eslog

import (
	"context"
	"github.com/SongOf/edge-storage-core/core"
	"go.uber.org/zap"
	"gorm.io/gorm/logger"
	"time"
)

type ModelLogger struct {
	logger *EsLogger
}

func NewModelLogger(option LogOption) *ModelLogger {
	innerLogger := EsLogger{}
	innerLogger.logger = zap.New(buildZapCore(&option),
		zap.AddStacktrace(zap.ErrorLevel),
		zap.AddCaller(),
		zap.AddCallerSkip(4))
	innerLogger.sugar = innerLogger.logger.Sugar()
	innerLogger.option = &option
	innerLogger.Info("model logger init success.")
	return &ModelLogger{logger: &innerLogger}
}

func (m *ModelLogger) LogMode(logger.LogLevel) logger.Interface {
	return m
}

func convertToZapFields(fields ...interface{}) []zap.Field {
	var zapFields []zap.Field
	for _, field := range fields {
		if zapField, ok := field.(zap.Field); ok {
			zapFields = append(zapFields, zapField)
		}
	}
	return zapFields
}

func (m *ModelLogger) c(ctx context.Context) *EsLogger {
	if coreCtx := core.Cast(ctx); coreCtx != nil {
		ctxModelLogger := coreCtx.Value(ctxModelLoggerName)
		if ctxModelLogger == nil {
			newLogger := m.logger.copyWithField(coreCtx.LogFields)
			coreCtx.Set(ctxModelLoggerName, newLogger)
			return newLogger
		}
		return ctxModelLogger.(*EsLogger)
	} else {
		return m.logger
	}
}

func (m *ModelLogger) Info(ctx context.Context, msg string, fields ...interface{}) {
	m.c(ctx).Info(msg, convertToZapFields(fields)...)
}

func (m *ModelLogger) Warn(ctx context.Context, msg string, fields ...interface{}) {
	m.c(ctx).Warn(msg, convertToZapFields(fields)...)
}

func (m *ModelLogger) Error(ctx context.Context, msg string, fields ...interface{}) {
	m.c(ctx).Error(msg, convertToZapFields(fields)...)
}

func (m *ModelLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	m.c(ctx).Info("TraceSql",
		Field("SQLError", err),
		Field("Duration", float64(elapsed.Nanoseconds())/1e6),
		Field("Rows", rows),
		Field("Sql", sql))
}
