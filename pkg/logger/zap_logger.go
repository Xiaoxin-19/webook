package logger

import "go.uber.org/zap"

type ZapLogger struct {
	zapLog *zap.Logger
}

func NewZapLogger(l *zap.Logger) *ZapLogger {
	return &ZapLogger{
		zapLog: l,
	}
}

func (z *ZapLogger) toArgs(args []Field) []zap.Field {
	res := make([]zap.Field, 0, len(args))
	for _, arg := range args {
		res = append(res, zap.Any(arg.Key, arg.Val))
	}
	return res
}

func (z *ZapLogger) Debug(format string, args ...Field) {
	z.zapLog.Debug(format, z.toArgs(args)...)
}

func (z *ZapLogger) Info(format string, args ...Field) {
	z.zapLog.Info(format, z.toArgs(args)...)
}

func (z *ZapLogger) Warn(format string, args ...Field) {
	z.zapLog.Warn(format, z.toArgs(args)...)
}

func (z *ZapLogger) Error(format string, args ...Field) {
	z.zapLog.Error(format, z.toArgs(args)...)
}
