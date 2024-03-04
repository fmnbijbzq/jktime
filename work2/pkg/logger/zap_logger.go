package logger

import "go.uber.org/zap"

type ZapLogger struct {
	l *zap.Logger
}

func NewZapLogger(log *zap.Logger) Logger {
	return &ZapLogger{
		l: log,
	}
}

// Debug implements Logger.
func (z *ZapLogger) Debug(msg string, args ...Field) {
	z.l.Debug(msg, z.toArgs(args...)...)
}

// Error implements Logger.
func (z *ZapLogger) Error(msg string, args ...Field) {
	z.l.Error(msg, z.toArgs(args...)...)
}

// Info implements Logger.
func (z *ZapLogger) Info(msg string, args ...Field) {
	z.l.Info(msg, z.toArgs(args...)...)
}

// Warn implements Logger.
func (z *ZapLogger) Warn(msg string, args ...Field) {
	z.l.Warn(msg, z.toArgs(args...)...)
}

func (z *ZapLogger) toArgs(args ...Field) []zap.Field {
	res := make([]zap.Field, 0, len(args))
	for _, val := range args {
		res = append(res, zap.Any(val.Key, val.Val))
	}
	return res
}
