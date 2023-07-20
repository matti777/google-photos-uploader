package logging

import (
	"os"
	"sync"

	gologging "github.com/op/go-logging"
)

var (
	logger     *gologging.Logger
	loggerOnce sync.Once
)

func InitLogging(level gologging.Level) {
	var format = gologging.MustStringFormatter("%{color}%{time:15:04:05.000} " +
		"%{shortfunc} ▶ %{level} " +
		"%{color:reset} %{message}")
	backend := gologging.NewLogBackend(os.Stderr, "", 0)
	formatter := gologging.NewBackendFormatter(backend, format)
	gologging.SetBackend(formatter)
	gologging.SetLevel(level, "uploader")
}

func MustGetLogger() *gologging.Logger {
	loggerOnce.Do(func() {
		logger = gologging.MustGetLogger("uploader")
	})

	return logger
}
