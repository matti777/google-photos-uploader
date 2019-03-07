// File operations

package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
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

// mustScanDir scans a directory for files and subdirectories; returns the
// lists of unprocessed files and subdirectories as well as any upload tokens
// not yet added to an album. Panics on error.
func mustScanDir(dir string, journal *pb.Journal,
	journalMap map[string]*pb.JournalEntry) ([]os.FileInfo,
	[]os.FileInfo, []string) {

	d, err := os.Open(dir)
	if err != nil {
		log.Fatalf("Failed to open directory '%v': %v", dir, err)
	}
	defer d.Close()

	infos, err := d.Readdir(0)
	if err != nil {
		log.Fatalf("Failed to read directory '%v': %v", dir, err)
	}

	files := []os.FileInfo{}
	dirs := []os.FileInfo{}
	unaddedUploadTokens := make([]string, 0)

	for _, info := range infos {
		name := info.Name()
		log.Debugf("Scan found name: %v", name)

		if info.Mode()&os.ModeSymlink != 0 {
			log.Debugf("Skipping symlink..")
			continue
		}

		entry := journalMap[name]

		if entry != nil {
			if entry.MediaItemCreated {
				log.Debugf("Image '%v' already uploaded", name)
				continue
			}

			if entry.UploadToken != "" {
				unaddedUploadTokens = append(unaddedUploadTokens,
					entry.UploadToken)
				log.Debugf("Image '%v' uploaded but not added to album", name)
			}
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

	log.Debugf("Directory '%v' scan completed.", dir)

	return files, dirs, unaddedUploadTokens
}

// getAlbumName forms the album name from the directory name
func getAlbumName(dirName string) string {
	albumName, err := replaceInString(dirName, nameSubstitutionTokens)
	if err != nil {
		log.Fatalf("Failed to replace in string: %v", err)
	}

	// log.Debugf("capitalize = %v", capitalize)

	if capitalize {
		albumName = strings.Title(albumName)
		log.Debugf("Capitalized album name: %v", albumName)
	}
	albumName = strings.Trim(albumName, " \n\r")

	return albumName
}

// getAlbum retrieves the matching existing album by its name or creates
// a new one if not found
func getAlbum(name string) (*photos.Album, error) {
	log.Debugf("Looking for album with title '%v'", name)

	// See if an existing album with such name exists
	for _, a := range albums {
		if a.Title == name {
			log.Debugf("Found album '%v'", name)
			return a, nil
		}
	}

	log.Debugf("Creating missing album: '%v'", name)

	if dryRun {
		log.Debugf("Dry run enabled; simulating album creation.")
		uuidv4, _ := uuid.NewRandom()
		return &photos.Album{ID: uuidv4.String(), Title: name}, nil
	}

	album, err := photosClient.CreateAlbum(name)
	if err != nil {
		log.Errorf("Failed to create Photos Album: %v", err)
		return nil, err
	}

	log.Debugf("Album created.")

	return album, nil
}

// Simulates the upload of a file.
// Returns image upload token or error.
func simulateUploadPhoto(path string, size int64, album *photos.Album,
	callback func(int64)) (string, error) {

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

	// Generate a random UUIDv4 to act as an upload token
	uuidv4, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	return uuidv4.String(), nil
}

// Synchronously uploads an image file (or simulates it). Manages a progress
// bar for the upload.
// Returns image upload token or error.
func upload(dir string, file os.FileInfo,
	padLength int, album *photos.Album) (string, error) {

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

// uploadAll uploads all photos in a given directory. It returns the
// list of upload tokens for the uploaded photos.
func uploadAll(dir, dirName string, album *photos.Album, journal *pb.Journal,
	journalMap map[string]*pb.JournalEntry, files []os.FileInfo) []string {

	uploadTokens := make([]string, 0, len(files))

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
			uploadToken, err := upload(dir, file, padLength, album)
			if err != nil {
				log.Fatalf("File upload failed: %v", err)
			} else {
				if uploadToken != "" {
					log.Debugf("Photo uploaded with token %v", uploadToken)
					uploadTokens = append(uploadTokens, uploadToken)
				} else {
					log.Debugf("Uploaded photo didnt receive upload token " +
						"-- it has already been uploaded with another token.")
				}

				if !dryRun {
					mustAddJournalEntry(dir, file.Name(), uploadToken,
						journal, &journalMap)
				}
			}
		})
	}

	// Wait for all the operations to complete
	q.GracefulShutdown()
	log.Debugf("All uploads finished.")
	uiprogress.Stop()

	return uploadTokens
}

// Processes all the entries in a single directory. Aborts as soon as
// an upload fails.
func mustProcessDir(dir string) {
	// Check that the diretory exists
	if exists, _ := directoryExists(dir); !exists {
		log.Fatalf("Directory '%v' does not exist!", dir)
	}

	log.Debugf("Processing directory %v ..", dir)

	dirName := filepath.Base(dir)
	albumName := getAlbumName(dirName)

	mustConfirm("About to upload directory '%v' as folder '%v'",
		dirName, albumName)

	// Get / create album by albumName
	album, err := getAlbum(albumName)
	if err != nil {
		log.Fatalf("Failed to get album: %v", err)
	}
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

	// Create a lookup map for faster access
	journalMap := newJournalMap(journal)

	// Find all the files & subdirectories
	files, dirs, unaddedUploadTokens := mustScanDir(dir, journal, journalMap)
	log.Debugf("unaddedUploadTokens: %v", unaddedUploadTokens)

	// Upload all the files in this directory
	uploadTokens := uploadAll(dir, dirName, album, journal, journalMap, files)

	// Add all previously uploaded but not added to any album -photos get
	// added to this album.
	uploadTokens = append(uploadTokens, unaddedUploadTokens...)

	if !dryRun {
		// We must split the tokens into groups of max MaxAddPhotosPerCall items
		chunks := chunked(uploadTokens, photos.MaxAddPhotosPerCall)
		for _, c := range chunks {
			// Create n media items at a time in the album
			if err := photosClient.AddToAlbum(album, c); err != nil {
				log.Fatalf("Failed to add photos to album: %v", err)
			}
		}

		// Update all the journal entries to reflect they
		// were successfully added
		for _, tok := range uploadTokens {
			for _, e := range journal.Entries {
				if e.UploadToken == tok {
					e.MediaItemCreated = true
					break
				}
			}
		}

		// Save the updated journal for this directory
		mustWriteJournalFile(dir, journal)
	}

	// If enabled, recurse into all the subdirectories
	for _, d := range dirs {
		if recurse {
			subDir := filepath.Join(dir, d.Name())
			mustProcessDir(subDir)
		} else {
			log.Debugf("Non-recursive; skipping directory: %v", d.Name())
		}
	}

	log.Debugf("Directory '%v' processed.", dirName)
}
