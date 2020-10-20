package logger

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/rollbar/rollbar-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Log Logger
)

func Init(cfg *config.Config) error {
	logConfig := zap.Config{
		OutputPaths: []string{cfg.LogOutput},
		Encoding:    "json",
		Level:       zap.NewAtomicLevelAt(getLevel(cfg.LogLevel)),
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:   "msg",
			LevelKey:     "level",
			TimeKey:      "time",
			EncodeLevel:  zapcore.LowercaseLevelEncoder,
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}

	log, err := logConfig.Build()
	if err != nil {
		return err
	}

	Log.Logger = log

	return nil
}

type Logger struct {
	*zap.Logger
}

func Field(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

func Info(msg string, tags ...zap.Field) {
	Log.Info(msg, tags...)
	Log.Sync()
}

func Debug(msg string, tags ...zap.Field) {
	Log.Debug(msg, tags...)
	Log.Sync()
}

func DebugJSON(msg interface{}, tags ...zap.Field) {
	jsonMsg, _ := json.Marshal(msg)
	Log.Debug(string(jsonMsg), tags...)
	Log.Sync()
}

func Warn(msg string, tags ...zap.Field) {
	Log.Warn(msg, tags...)
	Log.Sync()
}

func Error(err error, tags ...zap.Field) {
	msg := fmt.Sprintf("[ERROR: %v]", err)
	Log.Error(msg, tags...)
	Log.Sync()
	rollbar.Error(msg)
}

func getLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "panic":
		return zap.PanicLevel
	case "fatal":
		return zap.FatalLevel
	case "error":
		return zap.ErrorLevel
	case "warn", "warning":
		return zap.WarnLevel
	case "info":
		return zap.InfoLevel
	case "debug":
		return zap.DebugLevel
	default:
		return zap.InfoLevel
	}
}
