package exiftool

import "os/exec"

const (
	binaryName = "exiftool"
)

func IsInstalled() bool {
	_, err := exec.LookPath(binaryName)
	return err == nil
}
