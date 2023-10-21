package main

import (
	"fmt"
	"os"
	"time"

	"github.com/matti777/google-photos-uploader/internal/exif"
	"github.com/matti777/google-photos-uploader/internal/logging"

	"github.com/sirupsen/logrus"
)

func main() {
	log := logging.MustGetLogger()
	log.SetLevel(logrus.DebugLevel)
	log.SetOutput(os.Stdout)

	if len(os.Args) < 2 {
		fmt.Printf("Usage: %v <JPEG file path>\n", os.Args[0])
		os.Exit(1)
	}

	outputPath := "/tmp/test.jpg"

	t := time.Date(1987, 4, 26, 0, 0, 0, 0, time.UTC)
	if _, err := exif.WriteImageDate(os.Args[1], t, outputPath); err != nil {
		log.Fatalf("failed to set EXIF date: %v", err)
	}

	log.Debugf("Set EXIF date and wrote output to %v", outputPath)
}
