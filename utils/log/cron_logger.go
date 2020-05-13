package log

type CronLogger interface {
	// Info logs routine messages about cron's operation.
	Info(msg string, keysAndValues ...interface{})
	// Error logs an error condition.
	Error(err error, msg string, keysAndValues ...interface{})
}

func NewCronLogger() CronLogger {
	return log
}

func (l logger) Info(msg string, keysAndValues ...interface{}) {
	Info(msg)
}

func (l logger) Error(err error, msg string, keysAndValues ...interface{}) {
	Error(err)
}
