package logs

import (
	"errors"
	"io"
	"log"
	"sync"

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
	mu sync.Mutex // ensures atomic writes; protects the following fields

	logType       logTypeEnum
	useStacktrace bool
}

var loggerUniqueObject *loggerImpl

func GetLogger() Logger {
	if loggerUniqueObject == nil {
		loggerUniqueObject = &loggerImpl{
			mu: sync.Mutex{},
		}
	}

	log.SetOutput(loggerUniqueObject)
	return loggerUniqueObject
}

func GetLoggerWithStacktrace() Logger {
	if loggerUniqueObject == nil {
		loggerUniqueObject = &loggerImpl{
			mu: sync.Mutex{},
		}
	}

	loggerUniqueObject.useStacktrace = true

	log.SetOutput(loggerUniqueObject)
	return loggerUniqueObject
}

func (self *loggerImpl) Write(p []byte) (n int, err error) {
	sentry.CaptureMessage(string(p))

	switch self.logType {
	case eLogWarning:
		event := sentry.NewEvent()
		event.Level = sentry.LevelWarning
		event.Message = string(p)

		if sentry.CaptureEvent(event) != nil {
			return len(p), nil
		} else {
			return 0, errors.New("Please initize sentry first")
		}

	case eLogError:
		event := sentry.NewEvent()
		event.Level = sentry.LevelError
		event.Message = string(p)

		if self.useStacktrace {
			event.Threads = []sentry.Thread{{
				Stacktrace: sentry.NewStacktrace(),
				Crashed:    false,
				Current:    true,
			}}
		}

		if sentry.CaptureEvent(event) != nil {
			return len(p), nil
		} else {
			return 0, errors.New("Please initize sentry first")
		}

	case eLogFatal:
		if sentry.CaptureException(errors.New(string(p))) != nil {
			panic(string(p))
		} else {
			return 0, errors.New("Please initize sentry first")
		}

	default:
		event := sentry.NewEvent()
		event.Level = sentry.LevelInfo
		event.Message = string(p)

		if sentry.CaptureEvent(event) != nil {
			return len(p), nil
		} else {
			return 0, errors.New("Please initize sentry first")
		}
	}
}

func (self *loggerImpl) Infof(format string, args ...interface{}) {
	defer self.mu.Unlock()

	self.mu.Lock()
	self.logType = eLogInfo

	log.Printf(format + "\n", args)
}

func (self *loggerImpl) Warnf(format string, args ...interface{}) {
	defer self.mu.Unlock()

	self.mu.Lock()
	self.logType = eLogWarning

	log.Printf(format + "\n", args)
}

func (self *loggerImpl) Errorf(format string, args ...interface{}) {
	defer self.mu.Unlock()

	self.mu.Lock()
	self.logType = eLogError

	log.Printf(format + "\n", args)
}

func (self *loggerImpl) Fatalf(format string, args ...interface{}) {
	defer self.mu.Unlock()

	self.mu.Lock()
	self.logType = eLogFatal

	log.Printf(format + "\n", args)
}
