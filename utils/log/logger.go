package log

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"github.com/figment-networks/polkadothub-indexer/config"
	"strings"
)

var (
	log logger
)

type LoggerI interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
}

type logger struct {
	log *zap.Logger
}

func init() {
	logConfig := zap.Config{
		OutputPaths: []string{config.LogOutput()},
		Encoding:    "json",
		Level:       zap.NewAtomicLevelAt(getLevel()),
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:   "msg",
			LevelKey:     "level",
			TimeKey:      "time",
			EncodeLevel:  zapcore.LowercaseLevelEncoder,
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}

	var err error
	if log.log, err = logConfig.Build(); err != nil {
		panic(err)
	}
}

func getLevel() zapcore.Level {
	switch strings.ToLower(config.LogLevel()) {
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

func GetLogger() LoggerI {
	return log
}

func (l logger) Printf(format string, v ...interface{}) {
	if len(v) == 0 {
		Info(format)
	} else {
		Info(fmt.Sprintf(format, v...))
	}
}

func (l logger) Print(v ...interface{}) {
	Info(fmt.Sprintf("%v", v))
}

func Field(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

func Info(msg string, tags ...zap.Field) {
	log.log.Info(msg, tags...)
	log.log.Sync()
}

func Debug(msg string, tags ...zap.Field) {
	log.log.Debug(msg, tags...)
	log.log.Sync()
}

func Warn(msg string, tags ...zap.Field) {
	log.log.Warn(msg, tags...)
	log.log.Sync()
}

func Error(err error, tags ...zap.Field) {
	msg := fmt.Sprintf("[ERROR: %v]", err)
	log.log.Error(msg, tags...)
	log.log.Sync()
}
