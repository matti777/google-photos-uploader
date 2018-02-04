package main

import (
	"fmt"
	"os"
	"path/filepath"

	"mdahlbom/google-photos-uploader/pb"

	logging "github.com/op/go-logging"
	"github.com/urfave/cli"
)

// Our local logger
var log = logging.MustGetLogger("uploader")

func setupLogging() {
	var format = logging.MustStringFormatter("%{color}%{time:15:04:05.000} " +
		"%{shortfunc} ▶ %{level} " +
		"%{color:reset} %{message}")
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	formatter := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(formatter)
	if enableDebugLogging {
		logging.SetLevel(logging.DEBUG, "uploader")
	} else {
		logging.SetLevel(logging.INFO, "uploader")
	}
}

func processDir(dir string, recursive bool) {
	// Check that the diretory exists
	if exists, _ := directoryExists(dir); !exists {
		log.Fatalf("Directory '%v' does not exist!", dir)
	}

	// Read the journal file of the directory
	journal, err := readJournalFile(dir)
	if err != nil {
		log.Fatalf("Error reading Journal file: %v", err)
	}

	if journal == nil {
		// There is no journal file; create new one
		journal = &pb.Journal{}
	}

	log.Debugf("Read journal file: %+v", journal)

	// Read all the files in this directory
	d, err := os.Open(dir)
	if err != nil {
		log.Fatalf("Failed to open directory '%v': %v", dir, err)
	}

	infos, err := d.Readdir(0)
	if err != nil {
		log.Fatalf("Failed to read directory '%v': %v", dir, err)
	}

	for _, info := range infos {
		log.Debugf("Name: %v", info.Name())
		if info.Mode()&os.ModeSymlink != 0 {
			log.Debugf("Skipping symlink..")
			continue
		}

		if info.IsDir() {
			log.Debugf("It is a directory.")
		}
	}

	//TODO recurse
}

func defaultAction(c *cli.Context) error {
	log.Debugf("Running Default action..")

	baseDir := c.Args().Get(0)
	if baseDir == "" {
		log.Debugf("No base dir defined, using the CWD..")
		dir, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get CWD: %v", err)
		}
		baseDir = dir
	}

	log.Debugf("baseDir: %v", baseDir)

	// Resolve the base dir
	baseDir, err := filepath.Abs(baseDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path for '%v': %v", err)
	}
	log.Debugf("Cleaned baseDir: %v", baseDir)

	processDir(baseDir, false)

	return nil
}

func main() {
	// Set up logging
	setupLogging()

	appname := os.Args[0]
	log.Debugf("main(): running %v..", appname)

	// Setup CLI app framework
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Name = appname
	app.Description = fmt.Sprintf("Command-line utility for uploading "+
		"photos to Google Photos from a local disk directory. "+
		"For help, run '%v help'", appname)
	app.Copyright = "(c) 2018 Matti Dahlbom"
	app.Version = "0.0.1-alpha"
	app.Action = defaultAction
	//app.Commands = commands
	app.Run(os.Args)
}
