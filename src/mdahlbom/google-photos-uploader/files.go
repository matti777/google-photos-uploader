// File operations

package main

import (
	"os"
	"path/filepath"
	"strings"

	"mdahlbom/google-photos-uploader/pb"
)

// Checks that a path is an existing directory
func directoryExists(dir string) (bool, error) {
	log.Debugf("Checking if %v exists..", dir)

	if info, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			log.Fatalf("Failed to stat '%v': %v", dir, err)
			return false, err
		}
	} else {
		return info.IsDir(), nil
	}
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

func uploadAll(dir, dirName string, journal *pb.Journal,
	journalMap map[string]*pb.JournalEntry, files, dirs []os.FileInfo) {

	// Calculate the common padding length from the longest filename
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
			mustProcessDir(subDir)
			mustAddJournalEntry(dir, d.Name(), true, journal, &journalMap)
		} else {
			log.Debugf("Non-recursive; skipping directory: %v", d.Name())
		}
	}
}

// Processes all the entries in a single directory. Aborts as soon as
// an upload fails.
func mustProcessDir(dir string) {
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

	// Find all the files & subdirectories and upload them
	files, dirs := mustScanDir(dir, journal, journalMap)
	uploadAll(dir, dirName, journal, journalMap, files, dirs)

	log.Debugf("Directory '%v' uploaded OK.", dirName)
}
