package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

func NewLogger(level string) *zap.Logger {

	atomicLvl := zap.NewAtomicLevel()

	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.Lock(os.Stdout),
		atomicLvl,
	))

	switch level {
	case "DEBUG":
		atomicLvl.SetLevel(zap.DebugLevel)
	case "WARN":
		atomicLvl.SetLevel(zap.WarnLevel)
	case "ERROR":
		atomicLvl.SetLevel(zap.ErrorLevel)
	case "PANIC":
		atomicLvl.SetLevel(zap.PanicLevel)
	default:
		atomicLvl.SetLevel(zap.InfoLevel)
	}

	return logger
}
