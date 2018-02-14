// Logging configuration for google_photos
package google_photos

import (
	"os"

	logging "github.com/op/go-logging"
)

// Our local logger
var log = logging.MustGetLogger("google_photos")

// Configures the local logger
func setupLogging() {
	var format = logging.MustStringFormatter("%{color}%{time:15:04:05.000} " +
		"%{shortfunc} ▶ %{level} " +
		"%{color:reset} %{message}")
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	formatter := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(formatter)
	if enableDebugLogging {
		logging.SetLevel(logging.DEBUG, "google_photos")
	} else {
		logging.SetLevel(logging.INFO, "google_photos")
	}
}

func init() {
	setupLogging()
}
