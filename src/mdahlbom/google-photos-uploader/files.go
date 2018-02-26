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

	photos "mdahlbom/google-photos-uploader/google_photos"
	"mdahlbom/google-photos-uploader/pb"
	//"github.com/golang/protobuf/ptypes"
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

func getAlbum(dirName string) *photos.FeedEntry {

	//mustUpload := false

	/*
		if album.AlbumId == "" {
			// Check if there are files to upload; if so, must create an album
			for _, e := range journal.Entries {
				completed, err := ptypes.Timestamp(e.Completed)
				if err != nil {
					log.Fatalf("Failed to convert timestamp: %v", err)
				}

				if !e.IsDirectory && completed.IsZero() {
					// Found file that's not been uploaded; we must upload it.
					mustUpload = true
					break
				}
			}
		}
	*/

	//	if mustUpload {
	albumName, err := replaceInString(dirName, nameSubstitutionTokens)
	if err != nil {
		log.Fatalf("Failed to replace in string: %v", err)
	}

	if capitalize {
		albumName = strings.Title(albumName)
	}
	albumName = strings.Trim(albumName, " \n\r")

	log.Debugf("Looking for album with title '%v'", albumName)

	// We will need to have an existing album to upload to.
	// See if it exists
	for _, e := range albumFeed.Entries {
		if e.Title == albumName {
			log.Debugf("Found album '%v'", e.Title)
			return &e
		}
	}

	fmt.Printf("Missing album: %v", albumName)
	//	}

	return nil
}

// Simulates the upload of a file.
func simulateUploadPhoto(path string, size int64, album *photos.FeedEntry,
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

func upload(dir string, file os.FileInfo,
	padLength int, album *photos.FeedEntry) error {

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

	filePath := filepath.Join(dir, file.Name())

	progressCallback := func(count int64) {
		bar.Set(int(count))
	}

	if dryRun {
		return simulateUploadPhoto(filePath, file.Size(),
			album, progressCallback)
	} else {
		log.Debugf("Uploading file: %v", filePath)
		return photosClient.UploadPhoto(filePath, album, progressCallback)
	}
}

func uploadAll(dir, dirName string, journal *pb.Journal,
	journalMap map[string]*pb.JournalEntry, files, dirs []os.FileInfo) {

	album := getAlbum(dirName)

	// Only process files in this directory if there is something left to do
	// and the album already exists
	if album != nil {
		// Calculate the common padding length from the longest filename
		padLength := findLongestName(files)

		// Process the files in this directory firs
		for _, f := range files {
			if err := upload(dir, f, padLength, album); err != nil {
				log.Fatalf("File upload failed: %v", err)
			} else {
				// Uploaded file successfully
				//mustAddJournalEntry(dir, f.Name(), false, journal, &journalMap)
				mustAddJournalEntry(dir, f.Name(), journal, &journalMap)
			}
		}
	}

	// Then (possibly recursively) process the subdirectories
	for _, d := range dirs {
		if recurse {
			subDir := filepath.Join(dir, d.Name())
			mustProcessDir(subDir)
			//mustAddJournalEntry(dir, d.Name(), true, journal, &journalMap)
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
