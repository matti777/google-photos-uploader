package config

import (
	"sync"

	photos "github.com/matti777/google-photos-uploader/internal/googlephotos"
)

type Settings struct {
	// Filename extensions to consider as images
	ImageExtensions []string

	// Directory name -> photos folder substitution CSV string; should
	// be formatted as old1,new1,old2,new2, ... where new1 replaces old1 etc
	NameSubstitutionTokens string

	// Whether to capitalize words in directory name when forming
	// folder names
	Capitalize bool

	// Whether to skip parsing folder year from the folder name; in this case,
	// file date will be used as the EXIF date.
	NoParseYear bool

	// Whether to skip (assume Yes) all confirmations)
	SkipConfirmation bool

	// Whether doing a 'dry run', ie not actually sending anything.
	DryRun bool

	// Whether to recurse into subdirectories
	Recurse bool

	// Maximum concurrency (number of simultaneous uploads)
	MaxConcurrency int

	// List of albums
	// TODO this may need to change
	Albums []*photos.Album
}

var (
	settings     *Settings
	settingsOnce sync.Once
)

func MustGetSettings() *Settings {
	settingsOnce.Do(func() {
		settings = &Settings{
			ImageExtensions:        []string{"jpg", "jpeg"},
			NameSubstitutionTokens: "",
			Capitalize:             false,
			NoParseYear:            false,
			SkipConfirmation:       false,
			DryRun:                 false,
			Recurse:                false,
			MaxConcurrency:         1,
		}
	})

	return settings
}

func (s *Settings) FindAlbum(name string) *photos.Album {
	log.Debugf("Trying to find existing album with name '%v'", name)

	for _, a := range s.Albums {
		if a.Title == name {
			log.Debugf("Found album '%v'", name)
			return a
		}
	}

	return nil
}
