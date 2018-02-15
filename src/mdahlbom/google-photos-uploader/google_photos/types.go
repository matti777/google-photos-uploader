// Contains all the types used with communication with Picasa Web API
package google_photos

// Feed type (eg. list of Albums)
type Feed struct {
	Entries []FeedEntry `xml:"entry"`
}

// FeedEntry type (eg. an Album or a Photo)
type FeedEntry struct {
	AlbumID string `xml:"albumid"`
	Title   string `xml:"title"`
}

// Returns the number of entries in the feed
func (f *Feed) Count() int {
	return len(f.Entries)
}

// Retrieves a feed entry by it Title. Returns nil if not found
func (f *Feed) EntryByTitle(title string) *FeedEntry {
	for _, e := range f.Entries {
		if e.Title == title {
			return &e
		}
	}

	return nil
}
