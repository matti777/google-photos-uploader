package main

import (
	"fmt"
	"os"
	"time"

	"github.com/matti777/google-photos-uploader/internal/exif"

	"github.com/sirupsen/logrus"
)

var (
	// Our local logger
	log *logrus.Logger
)

// Build me like so:
// go build github.com/matti777/google-photos-uploader/cmd/exiftest

func main() {
	log := logrus.New()
	log.SetOutput(os.Stdout)

	if len(os.Args) < 2 {
		fmt.Printf("Usage: %v <JPEG file path>\n", os.Args[0])
		os.Exit(1)
	}

	t := time.Date(1987, 4, 26, 0, 0, 0, 0, time.UTC)
	if err := exif.SetDateIfNone(os.Args[1], t); err != nil {
		log.Fatalf("Failed to set EXIF date: %v", err)
	}

	log.Debugf("Set EXIF date")
}
