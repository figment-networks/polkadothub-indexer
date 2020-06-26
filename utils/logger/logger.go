package logger

import (
	"encoding/json"
	"fmt"
	"github.com/figment-networks/polkadothub-indexer/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"strings"
)

var (
	Log Logger
)

func Init(cfg *config.Config) (*Logger, error) {
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
		return nil, err
	}

	Log.log = log

	return &Logger{
		log: log,
	}, nil
}

type Logger struct {
	log *zap.Logger
}

func Field(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

func Info(msg string, tags ...zap.Field) {
	Log.log.Info(msg, tags...)
	Log.log.Sync()
}

func Debug(msg string, tags ...zap.Field) {
	Log.log.Debug(msg, tags...)
	Log.log.Sync()
}

func DebugJSON(msg interface{}, tags ...zap.Field) {
	jsonMsg, _ := json.Marshal(msg)
	Log.log.Debug(string(jsonMsg), tags...)
	Log.log.Sync()
}

func Warn(msg string, tags ...zap.Field) {
	Log.log.Warn(msg, tags...)
	Log.log.Sync()
}

func Error(err error, tags ...zap.Field) {
	msg := fmt.Sprintf("[ERROR: %v]", err)
	Log.log.Error(msg, tags...)
	Log.log.Sync()
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