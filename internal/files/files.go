package files

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"

	"github.com/matti777/google-photos-uploader/internal/config"
	photos "github.com/matti777/google-photos-uploader/internal/googlephotos"
	"github.com/matti777/google-photos-uploader/internal/logging"
	"github.com/matti777/google-photos-uploader/internal/pb"
	"github.com/matti777/google-photos-uploader/internal/util"
)

var (
	log      = logging.MustGetLogger()
	settings = config.MustGetSettings()
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
				continue
			}
		}

		if info.IsDir() {
			dirs = append(dirs, info)
		} else {
			ext := filepath.Ext(info.Name())
			ext = strings.ToLower(strings.TrimLeft(ext, "."))

			for _, e := range settings.ImageExtensions {
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
	albumName, err := util.ReplaceInString(dirName, settings.NameSubstitutionTokens)
	if err != nil {
		log.Fatalf("Failed to replace in string: %v", err)
	}

	// log.Debugf("capitalize = %v", capitalize)

	if settings.Capitalize {
		albumName = strings.Title(albumName)
		log.Debugf("Capitalized album name: %v", albumName)
	}
	albumName = strings.Trim(albumName, " \n\r")

	return albumName
}

// getAlbum retrieves the matching existing album by its name or creates
// a new one if not found
func getAlbum(name string) (*photos.Album, error) {
	if settings.DryRun {
		log.Fatalf("getAlbum called with dryRun enabled")
	}

	log.Debugf("Looking for album with title '%v'", name)

	// See if an existing album with such name exists
	for _, a := range settings.Albums {
		if a.Title == name {
			log.Debugf("Found album '%v'", name)
			return a, nil
		}
	}

	log.Debugf("Creating missing album: '%v'", name)

	album, err := photos.MustGetClient().CreateAlbum(name)
	if err != nil {
		log.Errorf("Failed to create Photos Album: %v", err)
		return nil, err
	}

	log.Debugf("Album created.")

	return album, nil
}

// Simulates the upload of a file.
// Returns image upload token or error.
func simulateUploadPhoto(path string, size int64,
	callback func(int64)) (string, error) {

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
func upload(progress *uiprogress.Progress, dir string, file os.FileInfo,
	padLength int) (string, error) {

	paddedName := strutil.PadRight(file.Name(), padLength, ' ')

	bar := progress.AddBar(int(file.Size())).PrependElapsed().
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

	if settings.DryRun {
		return simulateUploadPhoto(filePath, file.Size(), progressCallback)
	}

	return photos.MustGetClient().UploadPhoto(filePath, progressCallback)
}

// uploadAll uploads all photos in a given directory. It returns the
// list of upload tokens for the uploaded photos.
func uploadAll(dir, dirName string, journal *pb.Journal,
	journalMap map[string]*pb.JournalEntry, files []os.FileInfo) []string {

	uploadTokens := make([]string, 0, len(files))

	// Calculate the common padding length from the longest filename
	padLength := util.FindLongestName(files)

	// Create a concurrency execution queue for the uploads
	q, err := util.NewOperationQueue(settings.MaxConcurrency, 100)
	if err != nil {
		log.Fatalf("Failed to create operation queue: %v", err)
	}

	progress := uiprogress.New()
	progress.Start()

	// Process the files in this directory firs
	for _, f := range files {
		file := f

		q.Add(func() {
			uploadToken, err := upload(progress, dir, file, padLength)
			if err != nil {
				log.Fatalf("File upload failed: %v", err)
			} else {
				if uploadToken != "" {
					// log.Debugf("Photo uploaded with token %v", uploadToken)
					uploadTokens = append(uploadTokens, uploadToken)
				} else {
					log.Debugf("Uploaded photo didn't receive upload token " +
						"-- it has already been uploaded with another token.")
				}

				if !settings.DryRun {
					mustAddJournalEntry(dir, file.Name(), uploadToken,
						journal, &journalMap)
				}
			}
		})
	}

	// Wait for all the operations to complete
	q.GracefulShutdown()
	log.Debugf("All uploads finished.")
	progress.Stop()
	progress.Bars = nil

	return uploadTokens
}

// Processes all the entries in a single directory. Aborts as soon as
// an upload fails.
func MustProcessDir(dir string) {
	// Check that the diretory exists
	if exists, _ := directoryExists(dir); !exists {
		log.Fatalf("Directory '%v' does not exist!", dir)
	}

	log.Debugf("Processing directory %v ..", dir)

	dirName := filepath.Base(dir)
	albumName := getAlbumName(dirName)

	journal := newEmptyJournal()

	if !settings.DisregardJournal {
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

	uploadTokens := make([]string, 0)

	if len(files) > 0 {
		// Ask the user whether to continue uploading to this album
		util.MustConfirm("About to upload directory '%v' (%v photos) as album '%v'",
			dirName, len(files), albumName)

		// Upload all the files in this directory
		uploadTokens = uploadAll(dir, dirName, journal, journalMap, files)
	}

	// Add all previously uploaded but not added to any album -photos get
	// added to this album.
	uploadTokens = append(uploadTokens, unaddedUploadTokens...)

	var album *photos.Album

	// If there is something to add, add the photos to albums
	if (len(files) > 0 || len(uploadTokens) > 0) && !settings.DryRun {
		// Get / create album by albumName
		if a, err := getAlbum(albumName); err != nil {
			log.Fatalf("Failed to get album: %v", err)
		} else {
			album = a
		}

		log.Debugf("Adding photos to album %+v", album)

		// We must split the tokens into groups of max MaxAddPhotosPerCall items
		chunks := util.Chunked(uploadTokens, photos.MaxAddPhotosPerCall)
		for _, c := range chunks {
			// Create n media items at a time in the album
			if err := photos.MustGetClient().AddToAlbum(album, c); err != nil {
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
	if settings.Recurse {
		for _, d := range dirs {
			subDir := filepath.Join(dir, d.Name())
			MustProcessDir(subDir)
		}
	}

	log.Debugf("Directory '%v' processed.", dirName)
}
