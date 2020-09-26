package logger

import (
	"context"
	"os"
	"sync"

	"github/maxidelgado/toolkit-go/pkg/ctxhelper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	log     *zap.Logger
	logOnce sync.Once
)

type Config struct {
	Level string `yaml:"level"`
}

func Logger(config ...*Config) *zap.Logger {
	logOnce.Do(func() {
		log = buildLogger(config[0])
	})

	return log
}

func WithContext(ctx context.Context) *zap.Logger {
	ch := ctxhelper.WithContext(ctx)
	return Logger().With(
		zap.String("rid", ch.GetRequestId()),
	)
}

// buildLogger function for init new zap logger instance
func buildLogger(config *Config) *zap.Logger {
	// Define log level
	level := zap.NewAtomicLevel()

	// Set log level from .env file
	switch config.Level {
	case "debug":
		level.SetLevel(zap.DebugLevel)
	case "warn":
		level.SetLevel(zap.WarnLevel)
	case "error":
		level.SetLevel(zap.ErrorLevel)
	case "fatal":
		level.SetLevel(zap.FatalLevel)
	case "panic":
		level.SetLevel(zap.PanicLevel)
	default:
		level.SetLevel(zap.InfoLevel)
	}

	// Create new zap logger config
	encoderCfg := zap.NewProductionEncoderConfig()

	// Formated timestamp in the output.
	encoderCfg.EncodeTime = zapcore.RFC3339TimeEncoder

	// Create new zap logger
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		level,
	))

	defer func() {
		_ = logger.Sync()
	}()

	return logger
}
