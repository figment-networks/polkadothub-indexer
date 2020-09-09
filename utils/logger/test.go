package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type levelEnabler bool

func (o levelEnabler) Enabled(level zapcore.Level) bool {
	return bool(o)
}

func GetTestLogger() *zap.Logger {
	core, _ := observer.New(levelEnabler(false))
	return zap.New(core)
}

func InitTest() {
	Log = Logger{Logger: GetTestLogger()}
}
