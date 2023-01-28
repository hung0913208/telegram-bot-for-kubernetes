package logs

import (
	"errors"
	"io"
    "fmt"

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

	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
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

func (self *loggerImpl) Infof(format string, args ...interface{}) {
    self.writeLog(fmt.Sprintf(format, args), eLogInfo)
}

func (self *loggerImpl) Warnf(format string, args ...interface{}) {
    self.writeLog(fmt.Sprintf(format, args), eLogWarning)
}

func (self *loggerImpl) Errorf(format string, args ...interface{}) {
    self.writeLog(fmt.Sprintf(format, args), eLogError)
}

func (self *loggerImpl) Fatalf(format string, args ...interface{}) {
    self.writeLog(fmt.Sprintf(format, args), eLogFatal)
}

func (self *loggerImpl) Write(b []byte) (int, error) {
    err := self.writeLog(string(b), eLogInfo)
    if err != nil {
        return 0, err
    }

    return len(b), nil
}

