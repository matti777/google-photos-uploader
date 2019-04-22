package exif

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	exif "github.com/dsoprea/go-exif"
	jpeg "github.com/dsoprea/go-jpeg-image-structure"
)

const (
	tagProcessingSoftware = "ProcessingSoftware"
	tagDateTimeOriginal   = "DateTimeOriginal"
	tagDateTime           = "DateTime"
)

func setTag(rootIB *exif.IfdBuilder, ifdPath, tagName, tagValue string) error {
	fmt.Printf("setTag(): ifdPath: %v, tagName: %v, tagValue: %v\n",
		ifdPath, tagName, tagValue)

	ifdIb, err := exif.GetOrCreateIbFromRootIb(rootIB, ifdPath)
	if err != nil {
		return fmt.Errorf("failed to get or create IB: %v", err)
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
	sl, err := parser.ParseFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to parse JPEG file: %v", err)
	}

	rootIb, err := sl.ConstructExifBuilder()
	if err != nil {
		log.Println("No EXIF; creating it from scratch")

		//TODO what is wrong with this - no EXIF data end up in the JPEG
		im := exif.NewIfdMappingWithStandard()
		ti := exif.NewTagIndex()
		if err := exif.LoadStandardTags(ti); err != nil {
			return fmt.Errorf("failed to load standard tags: %v", err)
		}

		rootIb = exif.NewIfdBuilder(im, ti, exif.IfdPathStandard,
			exif.EncodeDefaultByteOrder)
		// if err := rootIb.AddStandardWithName(tagProcessingSoftware,
		// 	"photos-uploader"); err != nil {
		// 	return fmt.Errorf("failed to add ProcessingSoftware tag: %v", err)
		// }
	}

	// Form our timestamp string
	ts := exif.ExifFullTimestampString(t)

	// Set tags in IFD0
	ifdPath := "IFD0"
	if err := setTag(rootIb, ifdPath, tagDateTime, ts); err != nil {
		return fmt.Errorf("failed to set tag %v: %v", tagDateTime, err)
	}
	if err := setTag(rootIb, ifdPath, tagProcessingSoftware,
		"photos-uploader"); err != nil {
		return fmt.Errorf("failed to set tag %v: %v", tagDateTime, err)
	}

	// Set tags in IFD/Exif
	ifdPath = "IFD/Exif"
	if err := setTag(rootIb, ifdPath, tagDateTimeOriginal, ts); err != nil {
		return fmt.Errorf("failed to set tag %v: %v", tagDateTime, err)
	}

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
