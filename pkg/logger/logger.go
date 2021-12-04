package logger

import (
	"os"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//Logger type
type Logger struct {
	*zap.Logger
}

type Config struct {
	Environment string
}

var logger *zap.Logger = zap.NewExample()

func InitLogger(cf *Config) error {
	ec := zap.NewProductionEncoderConfig()
	ec.EncodeTime = zapcore.ISO8601TimeEncoder

	var cfg zap.Config
	if cf.Environment == "production" {
		cfg = zap.NewProductionConfig()
		cfg.OutputPaths = []string{"stdout"}
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.OutputPaths = []string{"stdout"}
	}

	cfg.EncoderConfig = ec

	l, err := cfg.Build()
	if err != nil {
		return errors.WithStack(err)
	}

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	logger = l.With(zap.String("hostname", hostname))
	return nil
}

func Info(msg string, fields ...zapcore.Field) {
	logger.WithOptions(zap.AddCallerSkip(1)).Info(msg, fields...)
}

func Debug(msg string, fields ...zapcore.Field) {
	logger.WithOptions(zap.AddCallerSkip(1)).Debug(msg, fields...)
}

func Fatal(msg string, fields ...zapcore.Field) {
	logger.WithOptions(zap.AddCallerSkip(1)).Fatal(msg, fields...)
}

func Error(msg string, fields ...zapcore.Field) {
	logger.WithOptions(zap.AddCallerSkip(1)).Error(msg, fields...)
}

func With(fields ...zapcore.Field) *zap.Logger {
	return logger.With(fields...)
}
