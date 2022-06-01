package logger

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Logger(names ...string) logr.Logger {
	config := zap.NewProductionConfig()

	config.DisableCaller = false
	config.DisableStacktrace = true
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	config.Encoding = "console"

	l, _ := config.Build()
	for _, name := range names {
		l = l.Named(name)
	}
	return zapr.NewLogger(l)

}
