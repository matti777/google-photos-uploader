package exif

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"time"

	goexif "github.com/dsoprea/go-exif/v3"
	jpeg "github.com/dsoprea/go-jpeg-image-structure/v2"
	"github.com/pkg/errors"
)

const (
	// tagProcessingSoftware = "ProcessingSoftware"
	tagDateTimeOriginal = "DateTimeOriginal"
	tagDateTime         = "DateTime"
)

func tagHasValue(ifdIb *goexif.IfdBuilder, tagName string) bool {
	tag, err := ifdIb.FindTagWithName(tagName)
	if err != nil {
		return false
	}

	value := tag.Value()
	if value == nil {
		return false
	}

	return value.String() != ""
}

// Sets the EXIF DateTime to the given Time unless it has
// already been defined.
// Returns the number of bytes written to outputPath or error.
func WriteImageDate(filepath string, t time.Time, outputPath string) (int64, error) {
	parser := jpeg.NewJpegMediaParser()
	mediaCtx, err := parser.ParseFile(filepath)
	if err != nil {
		return 0, errors.Errorf("failed to parse JPEG file: %v", err)
	}

	sl := mediaCtx.(*jpeg.SegmentList)

	rootIb, err := sl.ConstructExifBuilder()
	if err != nil {
		// log.Printf("No EXIF data, creating it from scratch: %v", err)
		return 0, errors.Wrap(err, "Failed to construct EXIF builder")
	}

	ifdPath := "IFD0"

	ifdIb, err := goexif.GetOrCreateIbFromRootIb(rootIb, ifdPath)
	if err != nil {
		return 0, errors.Wrap(err, "Failed to get or create ib")
	}

	if !tagHasValue(ifdIb, tagDateTime) {
		if err := ifdIb.SetStandardWithName(tagDateTime, t); err != nil {
			return 0, errors.Wrap(err, "failed to set DateTime tag value")
		}
	}

	if !tagHasValue(ifdIb, tagDateTimeOriginal) {
		if err := ifdIb.SetStandardWithName(tagDateTimeOriginal, t); err != nil {
			return 0, errors.Wrap(err, "failed to set DateTimeOriginal tag value")
		}
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

	fmt.Printf("Number of image bytes: %v\n", len(b.Bytes()))

	// Save the file
	bytes := b.Bytes()
	if err := ioutil.WriteFile(outputPath, bytes, 0644); err != nil {
		return 0, errors.Wrap(err, "failed to write JPEG file")
	}

	fmt.Printf("Wrote %v\n", outputPath)

	return int64(len(bytes)), nil
}
