package files

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"

	"github.com/matti777/google-photos-uploader/internal/config"
	"github.com/matti777/google-photos-uploader/internal/exif"
	photos "github.com/matti777/google-photos-uploader/internal/googlephotos"
	"github.com/matti777/google-photos-uploader/internal/logging"
	"github.com/matti777/google-photos-uploader/internal/util"
)

var (
	log            = logging.MustGetLogger()
	settings       = config.MustGetSettings()
	utcLocation, _ = time.LoadLocation("UTC")
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

// Returns files and subdirectories. Panics on errors.
func mustScanDirectory(dir string) ([]os.FileInfo, []os.FileInfo) {
	d, err := os.Open(dir)
	if err != nil {
		log.Fatalf("Failed to open directory '%v': %v", dir, err)
	}
	defer d.Close()

	infos, err := d.Readdir(0)
	if err != nil {
		log.Fatalf("Failed to read directory '%v': %v", dir, err)
	}

	dirs := []os.FileInfo{}
	files := []os.FileInfo{}

	for _, info := range infos {
		if info.IsDir() {
			dirs = append(dirs, info)
		} else {
			files = append(files, info)
		}
	}

	return files, dirs
}

// formAlbumName forms the album name from the directory name
func formAlbumName(dirName string, capitalize bool, substitutionTokens string) string {
	albumName, err := util.ReplaceInString(dirName, substitutionTokens)
	if err != nil {
		log.Fatalf("Failed to replace in string: %v", err)
	}

	if capitalize {
		// TODO fix deprecation
		albumName = strings.Title(albumName)
		log.Debugf("Capitalized album name: %v", albumName)
	}
	albumName = strings.Trim(albumName, " \n\r")

	return albumName
}

func createAlbum(name string) (*photos.Album, error) {
	log.Debugf("Creating new Photos album: '%v'", name)

	album, err := photos.MustGetClient().CreateAlbum(name)
	if err != nil {
		log.Errorf("Failed to create Photos Album: %v", err)
		return nil, err
	}

	log.Debugf("Album created.")

	return album, nil
}

// Simulates the upload of a file.
func simulateUploadPhoto(path string, size int64,
	callback func(int64)) (string, error) {

	const steps = 10
	const duration = 1.0
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
	padLength, albumYear int) (string, error) {

	paddedName := strutil.PadRight(file.Name(), padLength, ' ')

	filePath := filepath.Join(dir, file.Name())
	fileSize := file.Size()

	// Write creation date to EXIF data so Google Photos album will get a proper year
	if !settings.DryRun {
		tempFile, err := ioutil.TempFile("", "*.jpeg")
		if err != nil {
			log.Fatalf("failed to create temp file: %v", err)
		}
		fileDate := getDateForFile(albumYear, file)
		log.Debugf("Writing file date %v to image", fileDate)
		fileSize, err = exif.WriteImageDate(filePath, fileDate, tempFile.Name())
		if err != nil {
			log.Fatalf("failed to write EXIF info: %v", err)
		}
		filePath = tempFile.Name()
	}

	bar := progress.AddBar(int(fileSize)).PrependElapsed().AppendCompleted()
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return paddedName
	})
	bar.Fill = '#'
	bar.Head = '#'
	bar.Empty = ' '

	progressCallback := func(count int64) {
		bar.Set(int(count))
	}

	if !settings.DryRun {
		return photos.MustGetClient().UploadPhoto(filePath, progressCallback)
	} else {
		return simulateUploadPhoto(filePath, file.Size(), progressCallback)
	}
}

func getDateForFile(albumYear int, file fs.FileInfo) time.Time {
	fileDate := file.ModTime()

	// If the file date's year matches the parsed album year
	// (or album year has not been parsed successfully) use the file date
	if albumYear <= 0 || albumYear == fileDate.Year() {
		return fileDate
	}

	// Otherwise form an arbitrary date using the parsed album year
	return time.Date(albumYear, 1, 10, 10, 10, 10, 0, utcLocation)
}

// Uploads the given files and returns the upload tokens for the uploaded photos.
func uploadAll(absoluteDirPath string, albumYear int, files []os.FileInfo) []string {
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

	for _, f := range files {
		file := f

		q.Add(func() {
			uploadToken, err := upload(progress, absoluteDirPath, file, padLength, albumYear)
			if err != nil {
				log.Fatalf("File upload failed: %v", err)
			}

			if uploadToken != "" {
				uploadTokens = append(uploadTokens, uploadToken)
			} else {
				log.Debugf("Uploaded photo didn't receive upload token " +
					"-- it has already been uploaded with another token.")
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

// Out of a list of files as input, filters out our non-supported files and
// attempts to upload the valid ones. Successfully uploaded files are then added to the
// specified album.
func handleFileUpload(absoluteDirPath string, files []fs.FileInfo,
	album *photos.Album, albumYear int) {

	// Filter out all non-supported files by extension
	imageFiles := make([]fs.FileInfo, 0, len(files))
	for _, f := range files {
		if mime.TypeByExtension(filepath.Ext(f.Name())) != "image/jpeg" {
			continue
		}
		imageFiles = append(imageFiles, f)
	}

	if len(imageFiles) == 0 {
		log.Debugf("No image files to uplkoad.")
		return
	}

	// Ask the user whether to continue uploading to this album
	util.MustConfirm("About to upload directory %v (%v image files) as album '%v'",
		absoluteDirPath, len(imageFiles), album.Title)

	// Upload all the files in this directory
	uploadTokens := uploadAll(absoluteDirPath, albumYear, imageFiles)

	// If there is something to add, add the photos to albums
	if len(uploadTokens) > 0 {
		log.Debugf("Adding %v photos to album %v", len(uploadTokens), album.Title)

		if !settings.DryRun {
			// We must split the tokens into groups of max MaxAddPhotosPerCall items
			chunks := util.Chunked(uploadTokens, photos.MaxAddPhotosPerCall)
			for _, c := range chunks {
				// Create n media items at a time in the album
				if err := photos.MustGetClient().AddToAlbum(album, c); err != nil {
					log.Fatalf("failed to add photos to album: %v", err)
				}
			}
		}
	}
}

func mustProcessPhotoAlbumSubDirectory(absoluteDirPath string, album *photos.Album, albumYear int) {
	// Find all the files & subdirectories
	files, dirs := mustScanDirectory(absoluteDirPath)

	handleFileUpload(absoluteDirPath, files, album, albumYear)

	if settings.Recurse {
		for _, d := range dirs {
			absoluteSubDirPath := filepath.Join(absoluteDirPath, d.Name())
			mustProcessPhotoAlbumSubDirectory(absoluteSubDirPath, album, albumYear)
		}
	}

	log.Debugf("Photo Album '%v' subdirectory %v processed.", album.Title, absoluteDirPath)
}

// Processes a Photo Album directory. Handles all the files in the directory and
// optionally all the subdirectories as well. Aborts as soon as an upload fails.
func mustProcessPhotoAlbumDirectory(absoluteDirPath string) {
	// Check that the diretory exists
	if exists, _ := directoryExists(absoluteDirPath); !exists {
		log.Fatalf("directory '%v' does not exist!", absoluteDirPath)
	}

	dirName := filepath.Base(absoluteDirPath)
	albumName := formAlbumName(dirName, settings.Capitalize, settings.NameSubstitutionTokens)

	log.Debugf("Processing directory %v with name %v, album name: %v..",
		absoluteDirPath, dirName, albumName)

	albumYear := time.Now().Year()
	var err error

	if !settings.NoParseYear {
		// Attempt to parse the album year from the directory name
		albumYear, err = util.ParseAlbumYear(dirName)
		if err != nil {
			fmt.Printf("failed to parse album year from directory name '%v' -- "+
				"skipping this directory. You can disable album year parsing by supplying "+
				"command line parameter --no-parse-year.", dirName)
			return
		}
	}

	log.Debugf("Using album year: %v", albumYear)

	// First check that there isn't already and album with such name
	album := settings.FindAlbum(albumName)
	if album != nil {
		fmt.Printf("Album '%v' already exists\n", albumName)
		return
	}

	// Create album by albumName
	fmt.Printf("Creating new Google Photos album: %v\n", albumName)
	if settings.DryRun {
		album = &photos.Album{Title: albumName, ID: "123"}
	} else {
		album, err = createAlbum(albumName)
		if err != nil {
			log.Fatalf("failed to create album: %v", err)
		}
	}

	// Find all the files & subdirectories
	files, dirs := mustScanDirectory(absoluteDirPath)

	handleFileUpload(absoluteDirPath, files, album, albumYear)

	if settings.Recurse {
		for _, d := range dirs {
			absoluteSubDirPath := filepath.Join(absoluteDirPath, d.Name())
			mustProcessPhotoAlbumSubDirectory(absoluteSubDirPath, album, albumYear)
		}
	}

	log.Debugf("Photo Album directory %v processed.", absoluteDirPath)
}

// Scans the "base" directory (one containing all the subdirectories of photos to
// be uploaded as albums)
func ProcessBaseDir(absoluteDirPath string) {
	// Check that the diretory exists
	if exists, _ := directoryExists(absoluteDirPath); !exists {
		log.Fatalf("Directory '%v' does not exist!", absoluteDirPath)
	}

	_, subdirs := mustScanDirectory(absoluteDirPath)

	for _, d := range subdirs {
		subDir := filepath.Join(absoluteDirPath, d.Name())
		mustProcessPhotoAlbumDirectory(subDir)
	}

	fmt.Printf("%v album(s) created.\n", len(subdirs))
}
