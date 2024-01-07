package exiftool

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"time"
)

const (
	binaryName = "exiftool"
	dateFormat = "2006:01:02 15:04:05" // YYYY:MM:DD HH:mm:ss
)

func IsInstalled() bool {
	_, err := exec.LookPath(binaryName)
	return err == nil
}

func MustCheckExiftoolInstalled() {
	if !IsInstalled() {
		fmt.Printf("This application requires the installation of exiftool.\n\n")
		fmt.Printf("To install the tool:\n\n")
		fmt.Printf("MacOS:\t\tbrew install exiftool\n")
		fmt.Printf("Debian:\t\tsudo apt-get install exiftool\n")
		fmt.Printf("Windows:\tSee https://exiftool.org/install.html\n\n")
		os.Exit(-1)
	}
}

func SetAllDates(inFilePath, outFilePath string, exifDate time.Time) error {
	allDates := fmt.Sprintf("-AllDates=\"%s\"", exifDate.Format(dateFormat))
	args := []string{"-o", outFilePath, allDates, inFilePath}

	cmd := exec.Command(binaryName, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to call exiftool - stdout: %v, stderr: %v, error: %w",
			stdout.String(), stderr.String(), err)
	}

	return nil
}
