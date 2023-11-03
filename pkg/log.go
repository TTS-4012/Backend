package pkg

import (
	"errors"
	"github.com/sirupsen/logrus"
	"io"
	"ocontest/pkg/configs"
	"os"
)

type Logger struct {
	*logrus.Logger
	config configs.SectionLog
}

var (
	Log *Logger
)

func InitLog(config configs.SectionLog) {
	l := stdoutInit()
	logger := &Logger{l, config}
	err := logger.SetLevel(config.Level)
	if err != nil {
		_ = logger.SetLevel("info")
	}
	logger.SetReportCaller(true)
	Log = logger
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
