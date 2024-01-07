package main

import (
	"fmt"
	"os"
	"time"

	"github.com/matti777/google-photos-uploader/internal/exiftool"
	"github.com/matti777/google-photos-uploader/internal/logging"

	"github.com/sirupsen/logrus"
)

func main() {
	log := logging.MustGetLogger()
	log.SetLevel(logrus.DebugLevel)
	log.SetOutput(os.Stdout)

	exiftool.MustCheckExiftoolInstalled()

	if len(os.Args) < 2 {
		fmt.Printf("Usage: %v <JPEG file path>\n", os.Args[0])
		os.Exit(1)
	}

	outputPath := "/tmp/test.jpg"

	t := time.Date(1987, 4, 26, 10, 11, 12, 0, time.UTC)
	if err := exiftool.SetAllDates(os.Args[1], outputPath, t); err != nil {
		log.Fatalf("failed to set EXIF date: %v", err)
	}

	log.Debugf("Set EXIF date and wrote output to %v", outputPath)
}
