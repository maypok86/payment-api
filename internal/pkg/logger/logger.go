package logger

import (
	"io"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func levelFromString(l string) zapcore.Level {
	switch strings.ToLower(l) {
	case "debug":
		return zapcore.DebugLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func New(w io.Writer, level string) *zap.Logger {
	cfg := zap.NewProductionConfig()
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(cfg.EncoderConfig),
		zapcore.AddSync(w),
		levelFromString(level),
	)

	return zap.New(core)
}
