package logs

import (
	"errors"
	"io"

	sentry "github.com/getsentry/sentry-go"
)

type logTypeEnum int

const (
	eLogUnknown = iota
	eLogInfo
	eLogWarning
	eLogError
	eLogFatal
)

type Logger interface {
	io.Writer

	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Fatal(msg string)
}

type loggerImpl struct {
	logType       logTypeEnum
	useStacktrace bool
}

var loggerUniqueObject *loggerImpl

func NewLogger() Logger {
    return &loggerImpl{}
}

func NewLoggerWithStacktrace() Logger {
    return  &loggerImpl{useStacktrace: true}
}

func (self *loggerImpl) writeLog(msg string, level logTypeEnum) error {
	event := sentry.NewEvent()

    switch level {
    case eLogInfo:
	    event.Level = sentry.LevelInfo
    case eLogError:
	    event.Level = sentry.LevelError
    case eLogWarning:
	    event.Level = sentry.LevelWarning
    }

	switch level {
	case eLogFatal:
        err := sentry.CaptureException(errors.New(msg))

        if err != nil {
			panic(msg)
		} else {
			return errors.New("Please initize sentry first")
		}

	default:
	    event.Message = msg

	    if self.useStacktrace {
	    	event.Threads = []sentry.Thread{{
	    		Stacktrace: sentry.NewStacktrace(),
	    		Crashed:    false,
	    		Current:    true,
	    	}}
	    }

		if sentry.CaptureEvent(event) != nil {
			return nil
		} else {
			return errors.New("Please initize sentry first")
		}
	}

    return errors.New("Crashing!!")
}

func (self *loggerImpl) Info(msg string) {
    self.writeLog(msg, eLogInfo)
}

func (self *loggerImpl) Warn(msg string) {
    self.writeLog(msg, eLogWarning)
}

func (self *loggerImpl) Error(msg string) {
    self.writeLog(msg, eLogError)
}

func (self *loggerImpl) Fatal(msg string) {
    self.writeLog(msg, eLogFatal)
}

func (self *loggerImpl) Write(b []byte) (int, error) {
    err := self.writeLog(string(b), eLogInfo)
    if err != nil {
        return 0, err
    }

    return len(b), nil
}

