package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"mdahlbom/google-photos-uploader/pb"

	"github.com/golang/protobuf/ptypes"
	logging "github.com/op/go-logging"
	"github.com/urfave/cli"
)

// Our local logger
var log = logging.MustGetLogger("uploader")

// Filename extensions to consider as images
var (
	imageExtensions = []string{"jpg", "jpeg"}
)

// Configures the local logger
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

func mustAddJournalEntry(dir string, name string, isDir bool,
	journal *pb.Journal, journalMap *map[string]*pb.JournalEntry) {

	// Make sure there isnt already such an entry (sanity check)
	for _, e := range journal.Entries {
		if e.Name == name {
			log.Fatalf("Already found journal entry '%v' in journal", name)
		}
	}
	if (*journalMap)[name] != nil {
		log.Fatalf("Already found journal map entry '%v' in journal", name)
	}

	entry := &pb.JournalEntry{Name: name, IsDirectory: isDir,
		Completed: ptypes.TimestampNow()}

	journal.Entries = append(journal.Entries, entry)
	(*journalMap)[name] = entry

	// Save the journal
	mustWriteJournalFile(dir, journal)
	log.Debugf("Added journal entry: %+v", entry)
}

// Creates a map of the journal's file name entries for faster access
func newJournalMap(journal *pb.Journal) map[string]*pb.JournalEntry {
	m := map[string]*pb.JournalEntry{}

	if journal.Entries == nil {
		return m
	}

	for _, e := range journal.Entries {
		m[e.Name] = e
	}

	return m
}

// Simulates the upload of a file.
func simulateUpload(file *os.FileInfo) error {
	for i := 0; i < 5; i++ {
		time.Sleep(time.Millisecond * 250)
		log.Debugf("Uploading..")
	}

	return nil
}

// Processes all the entries in a single directory. Aborts as soon as
// an upload fails.
func processDir(dir string, recurse bool) {
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

	// Create a lookup map for faster access
	journalMap := newJournalMap(journal)

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
		name := info.Name()
		log.Debugf("Name: %v", name)

		if info.Mode()&os.ModeSymlink != 0 {
			log.Debugf("Skipping symlink..")
			continue
		}

		entry := journalMap[name]
		log.Debugf("entry: %+v", entry)

		if entry != nil {
			log.Debugf("Already uploaded, skipping..")
			continue
		}

		entryFileName := filepath.Join(dir, name)
		log.Debugf("entryFileName: %v", entryFileName)

		if info.IsDir() {
			if recurse {
				log.Debugf("Recursing into directory: %v", name)
				processDir(entryFileName, recurse)
				mustAddJournalEntry(dir, name, true, journal, &journalMap)
			} else {
				log.Debugf("Non-recursive; skipping directory: %v", name)
			}
		} else {
			log.Debugf("Uploading file '%v'..", name)
			if err := simulateUpload(&info); err != nil {
				log.Fatalf("File upload failed: %v", err)
			} else {
				// Uploaded file successfully
				log.Debugf("Passing journalMap: %v, %v",
					journalMap, &journalMap)

				mustAddJournalEntry(dir, name, false, journal, &journalMap)
			}
		}
	}

	log.Debugf("Directory '%v' uploaded OK.")
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
