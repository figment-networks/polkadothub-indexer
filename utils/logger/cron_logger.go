package logger

import "fmt"

type CronLogger interface {
	// Info logs routine messages about cron's operation.
	Info(msg string, keysAndValues ...interface{})
	// Error logs an error condition.
	Error(err error, msg string, keysAndValues ...interface{})
	// Print prints message
	Print(v ...interface{})
	// Printf prints formatted message
	Printf(format string, v ...interface{})
}

func NewCronLogger() CronLogger {
	return &cronLogger{}
}

type cronLogger struct {}

func (c cronLogger) Info(msg string, keysAndValues ...interface{}) {
	Info(msg)
}

func (c cronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	Error(err)
}

func (c cronLogger) Printf(format string, v ...interface{}) {
	if len(v) == 0 {
		Info(format)
	} else {
		Info(fmt.Sprintf(format, v...))
	}
}

func (c cronLogger) Print(v ...interface{}) {
	Info(fmt.Sprintf("%v", v))
}

