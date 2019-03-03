// Package googlephotos contains all the types used with communication
// with Google Photos API
package googlephotos

// Feed type (eg. list of Albums)
// type AlbumList struct {
// 	Entries []Album
// }

// Album type
type Album struct {
	ID    string
	Title string
}

// Count Returns the number of entries in the feed
// func (f *Feed) Count() int {
// 	return len(f.Entries)
// }

// // Retrieves a feed entry by it Title. Returns nil if not found
// func (f *Feed) EntryByTitle(title string) *FeedEntry {
// 	for _, e := range f.Entries {
// 		if e.Title == title {
// 			return &e
// 		}
// 	}

// 	return nil
// }
