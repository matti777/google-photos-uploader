package main

import (
	"fmt"
	"os"
	"time"

	"github.com/matti777/google-photos-uploader/internal/exif"
	"github.com/matti777/google-photos-uploader/internal/exiftool"
	"github.com/matti777/google-photos-uploader/internal/logging"

	"github.com/sirupsen/logrus"
)

func mustCheckExiftoolInstalled() {
	if !exiftool.IsInstalled() {
		fmt.Printf("This application requires the installation of exiftool.\n\n")
		fmt.Printf("To install the tool:\n\n")
		fmt.Printf("MacOS:\t\tbrew install exiftool\n")
		fmt.Printf("Debian:\t\tsudo apt-get install exiftool\n")
		fmt.Printf("Windows:\tSee https://exiftool.org/install.html\n\n")
		os.Exit(-1)
	}
}

func main() {
	log := logging.MustGetLogger()
	log.SetLevel(logrus.DebugLevel)
	log.SetOutput(os.Stdout)

	mustCheckExiftoolInstalled()

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
