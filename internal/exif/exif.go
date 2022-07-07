package exif

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	goexif "github.com/dsoprea/go-exif/v3"
	jpeg "github.com/dsoprea/go-jpeg-image-structure/v2"
	"github.com/pkg/errors"
)

const (
	tagProcessingSoftware = "ProcessingSoftware"
	tagDateTimeOriginal   = "DateTimeOriginal"
	tagDateTime           = "DateTime"
)

func setTag(rootIB *goexif.IfdBuilder, ifdPath, tagName, tagValue string) error {
	fmt.Printf("setTag(): ifdPath: %v, tagName: %v, tagValue: %v\n",
		ifdPath, tagName, tagValue)

	ifdIb, err := goexif.GetOrCreateIbFromRootIb(rootIB, ifdPath)
	if err != nil {
		return fmt.Errorf("failed to get or create IB: %v", err)
	}

	// See if the tag is already is
	tag, err := ifdIb.FindTagWithName(tagName)
	if err != nil {
		//return fmt.Errorf("failed to find tag %v: %v", tagName, err)
		if err == goexif.ErrTagEntryNotFound {
			log.Printf("No tag: %v", err)
		}
	} else {
		log.Printf("tag %v value: %v", tagName, tag.String())
	}

	if err := ifdIb.SetStandardWithName(tagName, tagValue); err != nil {
		return fmt.Errorf("failed to set DateTime tag: %v", err)
	}

	return nil
}

// SetDateIfNone sets the EXIF DateTime to the given Time unless it has
// already been defined.
func SetDateIfNone(filepath string, t time.Time) error {
	parser := jpeg.NewJpegMediaParser()
	mediaCtx, err := parser.ParseFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to parse JPEG file: %v", err)
	}

	sl := mediaCtx.(*jpeg.SegmentList)

	rootIb, err := sl.ConstructExifBuilder()
	if err != nil {
		// log.Printf("No EXIF data, creating it from scratch: %v", err)
		return errors.Wrap(err, "Failed to construct EXIF builder")
	}

	ifdPath := "IFD0"

	ifdIb, err := goexif.GetOrCreateIbFromRootIb(rootIb, ifdPath)
	if err != nil {
		return errors.Wrap(err, "Failed to get or create ib")
	}

	// rootIb, err := sl.ConstructExifBuilder()
	// if err != nil {
	// 	log.Printf("No EXIF data, creating it from scratch: %v", err)

	// 	im := goexif.NewIfdMappingWithStandard()
	// 	ti := goexif.NewTagIndex()
	// 	if err := goexif.LoadStandardTags(ti); err != nil {
	// 		return fmt.Errorf("failed to load standard tags: %v", err)
	// 	}

	// 	rootIb = goexif.NewIfdBuilder(im, ti, goexif.IfdPathStandard,
	// 		goexif.EncodeDefaultByteOrder)
	// }

	// Form our timestamp string
	ts := goexif.ExifFullTimestampString(t)

	//TODO remove
	_ = ts

	// Set tags in IFD0
	ifdPath := "IFD0"
	if err := setTag(rootIb, ifdPath, tagDateTime, ts); err != nil {
		return fmt.Errorf("failed to set tag %v: %v", tagDateTime, err)
	}
	// if err := setTag(rootIb, ifdPath, tagProcessingSoftware,
	// 	"photos-uploader"); err != nil {
	// 	return fmt.Errorf("failed to set tag %v: %v", tagDateTime, err)
	// }

	// Set tags in IFD/Exif
	// ifdPath := "IFD/Exif"
	// if err := setTag(rootIb, ifdPath, tagDateTimeOriginal, ts); err != nil {
	// 	return fmt.Errorf("failed to set tag %v: %v", tagDateTime, err)
	// }

	// Update the exif segment.
	if err := sl.SetExif(rootIb); err != nil {
		return fmt.Errorf("failed to set EXIF to jpeg: %v", err)
	}

	// Write the modified file
	b := new(bytes.Buffer)
	if err := sl.Write(b); err != nil {
		return fmt.Errorf("failed to create JPEG data: %v", err)
	}

	fmt.Printf("Number of image bytes: %v\n", len(b.Bytes()))

	//TODO overwrite the original file; check that JpegMediaParser closes
	// the file properly
	filepath = "/tmp/test.jpg"

	// Save the file
	if err := ioutil.WriteFile(filepath, b.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write JPEG file: %v", err)
	}

	fmt.Printf("Wrote %v\n", filepath)

	return nil
}
