package pkg

import (
	"errors"
	"io"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
	config *OContestConf
}

var (
	logOnce sync.Once
	Log     *Logger
)

func initLog(config *OContestConf) {
	logOnce.Do(func() {
		Log = defaultLogger(config)
	})
}

func NewLog(config *OContestConf) *Logger {
	l := stdoutInit()
	logger := &Logger{l, config}
	err := logger.SetLevel(config.Log.Level)
	if err != nil {
		_ = logger.SetLevel("info")
	}
	logger.SetReportCaller(config.Log.ReportCaller)
	return logger
}

func (l *Logger) GetLogger(loggerName string) *logrus.Entry {
	return l.WithFields(logrus.Fields{"logger": loggerName})
}

func (l *Logger) SetLevel(lvl string) error {
	level, err := logrus.ParseLevel(lvl)
	if err != nil {
		err = errors.New("failed to parse level")
		return err
	}
	l.Logger.Level = level
	return nil
}

func defaultLogger(config *OContestConf) *Logger {
	return NewLog(config)
}

func stdoutInit() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})

	var logWriter io.Writer = os.Stdout
	logger.SetOutput(logWriter)
	logger.SetNoLock()

	return logger
}
