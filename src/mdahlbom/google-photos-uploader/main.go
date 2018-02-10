package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mdahlbom/google-photos-uploader/pb"

	"github.com/golang/protobuf/ptypes"
	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
	logging "github.com/op/go-logging"
	"github.com/urfave/cli"
)

// Our local logger
var log = logging.MustGetLogger("uploader")

var (
	// Filename extensions to consider as images
	imageExtensions = []string{"jpg", "jpeg"}

	// Directory name -> photos folder substitution CSV string; should
	// be formatted as old1,new1,old2,new2, ... where new1 replaces old1 etc
	nameSubstitutionTokens = ""

	// Whether to skip (assume Yes) all confirmations)
	skipConfirmation = false

	// Whether doing a 'dry run', ie not actually sending anything.
	dryRun = false
)

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
func simulateUpload(progressBar *uiprogress.Bar, dir, dirName string,
	file os.FileInfo) error {

	const steps = 5

	remaining := file.Size()
	sent := int64(0)
	perStep := remaining / steps

	for i := 0; i < steps; i++ {
		time.Sleep(time.Millisecond * 250)

		if remaining < perStep {
			sent += remaining
		} else {
			sent += perStep
		}

		progressBar.Set(int(sent))

		remaining -= perStep
	}

	return nil
}

func upload(dir, dirName string, file os.FileInfo,
	padLength int) error {

	log.Debugf("Uploading '%v' for dirName '%v'", file.Name(), dirName)

	paddedName := strutil.PadRight(file.Name(), padLength, ' ')

	progress := uiprogress.New()
	bar := progress.AddBar(int(file.Size())).PrependElapsed().AppendCompleted()
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return paddedName
	})
	bar.Fill = '#'
	bar.Head = '#'
	bar.Empty = ' '

	progress.Start()
	defer progress.Stop()

	return simulateUpload(bar, dir, dirName, file)
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
			ext := filepath.Ext(info.Name())
			ext = strings.ToLower(strings.TrimLeft(ext, "."))
			log.Debugf("File ext is: %v", ext)

			for _, e := range imageExtensions {
				if ext == e {
					log.Debugf("Ext matches %v - is image", e)
					files = append(files, info)
					break
				}
			}
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

	dirName := filepath.Base(dir)
	folderName, err := replaceInString(dirName, nameSubstitutionTokens)
	if err != nil {
		log.Fatalf("Invalid replacement pattern: %v", nameSubstitutionTokens)
	}

	mustConfirm("About to upload directory '%v' as folder '%v' - continue?",
		dirName, folderName)

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
	padLength := findLongestName(files)

	// Process the files in this directory firs
	for _, f := range files {
		if err := upload(dir, dirName, f, padLength); err != nil {
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

	log.Debugf("Directory '%v' uploaded OK.", dirName)
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

	disregardJournal := GlobalBoolT(c, "disregard-journal")
	if disregardJournal {
		log.Debugf("Disregarding reading journal files..")
	}

	recursive := GlobalBoolT(c, "recursive")
	log.Debugf("Recurse into subdirectories: %v", recursive)

	skipConfirmation = GlobalBoolT(c, "yes")
	dryRun = GlobalBoolT(c, "dry-run")

	exts := c.String("extensions")
	if exts != "" {
		s := strings.Split(exts, ",")
		imageExtensions = make([]string, len(s))
		for i, item := range s {
			imageExtensions[i] = strings.ToLower(strings.Trim(item, " "))
		}
	}

	log.Debugf("Using image extensions: %v", imageExtensions)

	nameSubstitutionTokens = c.String("folder-name-substitutions")
	log.Debugf("Using folder name substitution tokens: %v",
		nameSubstitutionTokens)

	mustProcessDir(baseDir, recursive, disregardJournal)

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
	app.ArgsUsage = "[directory]"
	app.Usage = "A command line Google Photos upload utility"
	app.UsageText = fmt.Sprintf("%v [options] directory", appname)
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
		cli.BoolTFlag{
			Name:  "recursive, r",
			Usage: "Process subdirectories recursively",
		},
		cli.BoolTFlag{
			Name:  "yes, y",
			Usage: "Answer Yes to all confirmations",
		},
		cli.BoolTFlag{
			Name:  "dry-run",
			Usage: "Specify to just scan, not actually upload anything",
		},
		cli.StringFlag{
			Name: "folder-name-substitutions, s",
			Usage: "Directory name -> Photos Folder substition " +
				"tokens, default is no substitution. " +
				"The format is CSV like so: old1,new1,old2,new2 where token " +
				"new1 would replace token old1 etc. For example to replace " +
				"all underscores with spaces and add spaces around " +
				"all dashes, specify -s \"_, ,-, - \"",
		},
		cli.StringFlag{
			Name: "extensions, e",
			Usage: "File extensions to consider as uploadable images;  " +
				"eg. \"jpg, jpeg, png\". Default is \"jpg, jpeg\"",
		},
	}

	//app.Commands = commands
	app.Run(os.Args)
}
