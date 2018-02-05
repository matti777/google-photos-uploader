package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"mdahlbom/google-photos-uploader/pb"

	"github.com/golang/protobuf/ptypes"
	"github.com/gosuri/uiprogress"
	//	"github.com/gosuri/uiprogress/util/strutil"
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
func simulateUpload(dir string, file os.FileInfo) error {
	const steps = 5

	log.Debugf("Uploading file %v (%v bytes) ..", file.Name(), file.Size())

	bar := uiprogress.AddBar(steps).PrependElapsed().AppendCompleted()
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return file.Name()
	})
	bar.Fill = '#'
	bar.Head = ' '
	bar.Empty = ' '

	for i := 0; i < steps; i++ {
		time.Sleep(time.Millisecond * 250)
		bar.Set(i + 1)
	}

	return nil
}

// Finds the longest file name
func findLongestName(infos []os.FileInfo) int {
	longest := 0

	for _, info := range infos {
		if len(info.Name()) > longest {
			longest = len(info.Name())
		}
	}

	return longest
}

// Scans a directory for files and subdirectories; returns the lists of
// unprocessed files and subdirectories. Panics on error.
func mustScanDir(dir string, journal *pb.Journal,
	journalMap map[string]*pb.JournalEntry) ([]os.FileInfo, []os.FileInfo) {

	d, err := os.Open(dir)
	if err != nil {
		log.Fatalf("Failed to open directory '%v': %v", dir, err)
	}

	infos, err := d.Readdir(0)
	if err != nil {
		log.Fatalf("Failed to read directory '%v': %v", dir, err)
	}

	files := []os.FileInfo{}
	dirs := []os.FileInfo{}

	for _, info := range infos {
		name := info.Name()
		log.Debugf("Scan found name: %v", name)

		if info.Mode()&os.ModeSymlink != 0 {
			log.Debugf("Skipping symlink..")
			continue
		}

		entry := journalMap[name]

		if entry != nil {
			log.Debugf("Already uploaded, skipping..")
			continue
		}

		if info.IsDir() {
			dirs = append(dirs, info)
		} else {
			files = append(files, info)
		}
	}

	return files, dirs
}

// Processes all the entries in a single directory. Aborts as soon as
// an upload fails.
func mustProcessDir(dir string, recurse, disregardJournal bool) {
	// Check that the diretory exists
	if exists, _ := directoryExists(dir); !exists {
		log.Fatalf("Directory '%v' does not exist!", dir)
	}

	log.Debugf("Processing directory %v ..", dir)

	journal := &pb.Journal{}

	if !disregardJournal {
		// Read the journal file of the directory
		j, err := readJournalFile(dir)
		if err != nil {
			log.Fatalf("Error reading Journal file: %v", err)
		} else if j != nil {
			journal = j
		}
	}

	log.Debugf("Read journal file: %+v", journal)

	// Create a lookup map for faster access
	journalMap := newJournalMap(journal)

	files, dirs := mustScanDir(dir, journal, journalMap)

	// Process the files in this directory firs
	for _, f := range files {
		if err := simulateUpload(dir, f); err != nil {
			log.Fatalf("File upload failed: %v", err)
		} else {
			// Uploaded file successfully
			mustAddJournalEntry(dir, f.Name(), false, journal, &journalMap)
		}
	}

	// Then (possibly recursively) process the subdirectories
	for _, d := range dirs {
		if recurse {
			subDir := filepath.Join(dir, d.Name())
			mustProcessDir(subDir, recurse, disregardJournal)
			mustAddJournalEntry(dir, d.Name(), true, journal, &journalMap)
		} else {
			log.Debugf("Non-recursive; skipping directory: %v", d.Name())
		}
	}

	/*
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
					processDir(entryFileName, recurse, disregardJournal)
					mustAddJournalEntry(dir, name, true, journal, &journalMap)
				} else {
					log.Debugf("Non-recursive; skipping directory: %v", name)
				}
			} else {
				log.Debugf("Uploading file '%v'..", name)
				if err := simulateUpload(info); err != nil {
					log.Fatalf("File upload failed: %v", err)
				} else {
					// Uploaded file successfully
					log.Debugf("Passing journalMap: %v, %v",
						journalMap, &journalMap)

					mustAddJournalEntry(dir, name, false, journal, &journalMap)
				}
			}
		}
	*/
	log.Debugf("Directory '%v' uploaded OK.", dir)
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

	disregardJournal := c.BoolT("disregard-journal")
	if disregardJournal {
		log.Debugf("Disregarding reading journal files..")
	}

	uiprogress.Start()
	mustProcessDir(baseDir, false, disregardJournal)
	uiprogress.Stop()

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
	app.Flags = []cli.Flag{
		cli.BoolTFlag{
			Name:  "disregard-journal, d",
			Usage: "Disregard reading journal files; re-upload everything",
		},
	}

	//app.Commands = commands
	app.Run(os.Args)
}
