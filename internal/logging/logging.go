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

func setupLogging() {
	var format = gologging.MustStringFormatter("%{color}%{time:15:04:05.000} " +
		"%{shortfunc} ▶ %{level} " +
		"%{color:reset} %{message}")
	backend := gologging.NewLogBackend(os.Stderr, "", 0)
	formatter := gologging.NewBackendFormatter(backend, format)
	gologging.SetBackend(formatter)
	if enableDebugLogging {
		gologging.SetLevel(gologging.DEBUG, "uploader")
	} else {
		gologging.SetLevel(gologging.INFO, "uploader")
	}
}

func MustGetLogger() *gologging.Logger {
	loggerOnce.Do(func() {
		setupLogging()
		logger = gologging.MustGetLogger("uploader")
	})

	return logger
}
