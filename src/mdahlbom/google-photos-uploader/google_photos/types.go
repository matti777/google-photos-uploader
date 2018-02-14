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
