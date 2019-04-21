package exif

import (
	"fmt"
	"time"

	exif "github.com/dsoprea/go-exif"
	jpeg "github.com/dsoprea/go-jpeg-image-structure"
)

// SetDateIfNone sets the EXIF DateTime to the given Time unless it has
// already been defined.
func SetDateIfNone(filepath string, t time.Time) error {
	parser := jpeg.NewJpegMediaParser()
	sl, err := parser.ParseFile(filepath)
	if err != nil {
		return fmt.Errorf("Failed to parse JPEG file: %v", err)
	}

	fmt.Printf("sl = %v\n", sl)

	rootIb, err := sl.ConstructExifBuilder()
	if err != nil {
		return fmt.Errorf("Failed to construct EXIF builder: %v", err)
	}

	ifdPath := "IFD0"
	ifdIb, err := exif.GetOrCreateIbFromRootIb(rootIb, ifdPath)
	if err != nil {
		return fmt.Errorf("Failed to get or create IB: %v", err)
	}

	fmt.Printf("ifdIb: %v\n", ifdIb.DumpToStrings())

	// now := time.Now().UTC()
	// updatedTimestampPhrase := exif.ExifFullTimestampString(now)

	// err = ifdIb.SetStandardWithName("DateTime", updatedTimestampPhrase)
	// log.PanicIf(err)

	//TODO

	return nil
}
