package exif

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"time"

	goexif "github.com/dsoprea/go-exif/v3"
	goexifundefined "github.com/dsoprea/go-exif/v3/undefined"
	jpeg "github.com/dsoprea/go-jpeg-image-structure/v2"
	"github.com/matti777/google-photos-uploader/internal/logging"
	"github.com/pkg/errors"
)

const (
	tagDateTime          = "DateTime"
	tagModifyDate        = "ModifyDate"
	tagDateTimeOriginal  = "DateTimeOriginal"
	tagCreateDate        = "CreateDate"
	tagDateTimeDigitized = "DateTimeDigitized"
)

var (
	log = logging.MustGetLogger()
)

func setTagTimeValue(ifdIb *goexif.IfdBuilder, tagName string, newValue time.Time) error {
	if err := ifdIb.SetStandardWithName(tagName, newValue); err != nil {
		if errors.Is(err, goexif.ErrTagNotFound) {
			// This is fine
			return nil
		}
		return errors.Wrap(err, fmt.Sprintf("failed to set tag value for %v", tagName))
	}
	log.Debugf("tag %v value updated.", tagName)

	return nil
}

// Sets the EXIF DateTime to the given Time unless it has
// already been defined.
// Returns the number of bytes written to outputPath or error.
func WriteImageDate(filepath string, t time.Time, outputPath string) (int64, error) {
	log.Debugf("writing creation date to image %v", filepath)

	parser := jpeg.NewJpegMediaParser()
	mediaCtx, err := parser.ParseFile(filepath)
	if err != nil {
		return 0, errors.Errorf("failed to parse JPEG file: %v", err)
	}

	sl := mediaCtx.(*jpeg.SegmentList)

	rootIb, err := sl.ConstructExifBuilder()
	if err != nil {
		if errors.Is(err, goexifundefined.ErrUnparseableValue) {
			log.Debugf("Encountered unparseable tag value in input file %v", filepath)
			dropped, err := sl.DropExif()
			if err != nil {
				return 0, errors.Wrap(err, "failed to drop EXIF data")
			}
			if !dropped {
				log.Debugf("EXIF data not dropped") // wheh does this even happen..
			}
			rootIb, err = sl.ConstructExifBuilder()
			if err != nil {
				return 0, errors.Wrap(err, "failed to construct EXIF builder from scratch")
			}
		} else {
			return 0, errors.Wrap(err, "Failed to construct EXIF builder")
		}
	}

	// For these values, see https://exiftool.org/TagNames/EXIF.html

	ifdIb, err := goexif.GetOrCreateIbFromRootIb(rootIb, "IFD0")
	if err != nil {
		return 0, errors.Wrap(err, "failed to get or create ib IFD0")
	}

	if err := setTagTimeValue(ifdIb, tagDateTime, t); err != nil {
		return 0, err
	}

	if err := setTagTimeValue(ifdIb, tagModifyDate, t); err != nil {
		return 0, err
	}

	if err := setTagTimeValue(ifdIb, tagDateTimeOriginal, t); err != nil {
		return 0, err
	}
	if err := setTagTimeValue(ifdIb, tagDateTimeDigitized, t); err != nil {
		return 0, err
	}
	if err := setTagTimeValue(ifdIb, tagCreateDate, t); err != nil {
		return 0, err
	}

	// Update the exif segment.
	if err := sl.SetExif(rootIb); err != nil {
		return 0, errors.Wrap(err, "failed to set EXIF to jpeg")
	}

	// Write the modified file
	b := new(bytes.Buffer)
	if err := sl.Write(b); err != nil {
		return 0, errors.Wrap(err, "failed to write JPEG data")
	}

	// Save the file
	bytes := b.Bytes()
	if err := ioutil.WriteFile(outputPath, bytes, 0644); err != nil {
		return 0, errors.Wrap(err, "failed to write JPEG file")
	}

	return int64(len(bytes)), nil
}
