// File operations

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"

	photos "mdahlbom/google-photos-uploader/googlephotos"
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
			log.Debugf("Image '%v' already uploaded", name)
			continue
		}

		if info.IsDir() {
			dirs = append(dirs, info)
		} else {
			ext := filepath.Ext(info.Name())
			ext = strings.ToLower(strings.TrimLeft(ext, "."))

			for _, e := range imageExtensions {
				if ext == e {
					files = append(files, info)
					break
				}
			}
		}
	}

	return files, dirs
}

// Forms the album name and retrieves the matching Feed Entry or nil
// if an album by that name is not found
func getAlbum(dirName string) *photos.Album {
	albumName, err := replaceInString(dirName, nameSubstitutionTokens)
	if err != nil {
		log.Fatalf("Failed to replace in string: %v", err)
	}

	log.Debugf("capitalize = %v", capitalize)

	if capitalize {
		albumName = strings.Title(albumName)
		log.Debugf("Capitalized album name: %v", albumName)
	}
	albumName = strings.Trim(albumName, " \n\r")

	log.Debugf("Looking for album with title '%v'", albumName)

	// We will need to have an existing album to upload to.
	// See if it exists
	for _, a := range albums {
		if a.Title == albumName {
			log.Debugf("Found album '%v'", albumName)
			return a
		}
	}

	fmt.Printf("Missing album: %v\n", albumName)

	return nil
}

// Simulates the upload of a file.
func simulateUploadPhoto(path string, size int64, album *photos.Album,
	callback func(int64)) error {

	log.Debugf("Simulating uploading file: %v", path)

	const steps = 10
	const duration = 2.0
	const sleep = float32(time.Second) * (float32(duration) / steps)

	remaining := size
	sent := int64(0)
	perStep := remaining / steps

	for i := 0; i < steps; i++ {
		time.Sleep(time.Duration(sleep))

		if remaining < perStep {
			sent += remaining
		} else {
			sent += perStep
		}

		callback(sent)

		remaining -= perStep
	}

	return nil
}

// Synchronously uploads an image file (or simulates it). Manages a progress
// bar for the upload.
func upload(dir string, file os.FileInfo,
	padLength int, album *photos.Album) error {

	log.Debugf("Uploading file '%v'", file.Name())
	paddedName := strutil.PadRight(file.Name(), padLength, ' ')

	bar := uiprogress.AddBar(int(file.Size())).PrependElapsed().
		AppendCompleted()
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return paddedName
	})
	bar.Fill = '#'
	bar.Head = '#'
	bar.Empty = ' '

	filePath := filepath.Join(dir, file.Name())

	progressCallback := func(count int64) {
		bar.Set(int(count))
	}

	if dryRun {
		return simulateUploadPhoto(filePath, file.Size(),
			album, progressCallback)
	} else {
		return photosClient.UploadPhoto(filePath, album, progressCallback)
	}
}

func uploadAll(dir, dirName string, journal *pb.Journal,
	journalMap map[string]*pb.JournalEntry, files, dirs []os.FileInfo) {

	album := getAlbum(dirName)

	// Only process files in this directory if there is something left to do
	// and the album already exists
	if album != nil {
		log.Debugf("Got album: %+v", album)
		mustConfirm("About to upload directory '%v' as folder '%v'",
			dirName, album.Title)

		// Calculate the common padding length from the longest filename
		padLength := findLongestName(files)

		// Create a concurrency execution queue for the uploads
		q, err := NewOperationQueue(maxConcurrency, 100)
		if err != nil {
			log.Fatalf("Failed to create operation queue: %v", err)
		}

		uiprogress.Start()

		// Process the files in this directory firs
		for _, f := range files {
			file := f

			q.Add(func() {
				if err := upload(dir, file, padLength, album); err != nil {
					log.Fatalf("File upload failed: %v", err)
				} else {
					mustAddJournalEntry(dir, file.Name(), journal, &journalMap)
				}
			})
		}

		// Wait for all the operations to complete
		q.GracefulShutdown()
		log.Debugf("All uploads finished.")
		uiprogress.Stop()
	}

	// Then (possibly recursively) process the subdirectories
	for _, d := range dirs {
		if recurse {
			subDir := filepath.Join(dir, d.Name())
			mustProcessDir(subDir)
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

	log.Debugf("Processing directory %v ..", dir)

	journal := newEmptyJournal()

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

	log.Debugf("Directory '%v' processed.", dirName)
}
