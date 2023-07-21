package logging

import (
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	logger     *logrus.Logger
	loggerOnce sync.Once
)

func MustGetLogger() *logrus.Logger {
	loggerOnce.Do(func() {
		logger = logrus.New()
	})

	return logger
}
