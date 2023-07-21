package files

import (
	"testing"
)

func TestFormAlbumName(t *testing.T) {
	if formAlbumName("Foo_Bar-2010", false, "_, ") != "Foo Bar-2010" {
		t.Errorf("failed to form album name")
	}

	if formAlbumName("foo bar_baz_-_2010", true, "_, ") != "Foo Bar Baz - 2010" {
		t.Errorf("failed to form album name")
	}
}
