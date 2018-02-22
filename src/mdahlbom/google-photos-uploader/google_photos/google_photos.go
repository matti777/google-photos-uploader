// Manages authenticating to Google and providing an OAuth2 token
package google_photos

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"mdahlbom/google-photos-uploader/google_photos/util"

	"golang.org/x/oauth2"
)

// Set this variable to enable error logging output from the library
var ErrorLogFunc func(string, ...interface{}) = nil

// Our API client type. Create with NewClient().
type Client struct {
	httpClient *http.Client
}

// Creates a new API client using an OAuth2 token. To acquire the token,
// run the authorization flow with util.Authenticator.
func NewClient(clientID, clientSecret string, token *oauth2.Token) *Client {
	config := util.NewOAuth2Config(clientID, clientSecret)
	httpClient := config.Client(oauth2.NoContext, token)

	return &Client{httpClient: httpClient}
}

// Lists all the Albums
func (c *Client) ListAlbums() (*Feed, error) {
	url := "https://picasaweb.google.com/data/feed/api/user/default"
	return c.fetchFeed(url)
}

// Upload a photo to an album synchronously.
func (c *Client) UploadPhoto(name string, data []byte, album *FeedEntry) error {
	url := fmt.Sprintf("https://picasaweb.google.com/data/feed/api/user/"+
		"default/albumid/%v", album.AlbumID)

	ErrorLogFunc("Using upload url: %v", url)

	//TODO

	return nil
}

// Returns a Feed for the given resource (endpoint URL)
func (c *Client) fetchFeed(endpoint string) (*Feed, error) {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		ErrorLogFunc("Failed to create HTTP request: %v", err)
		return nil, err
	}

	req.Header.Set("GData-Version", "3")

	res, err := c.httpClient.Do(req)
	if err != nil {
		ErrorLogFunc("Failed to fetch the feed: %v", err)
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		ErrorLogFunc("Got non-OK response code: %v", res.StatusCode)
		return nil, errors.New(res.Status)
	}

	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		ErrorLogFunc("Failed to read response body: %v", err)
		return nil, err
	}

	feed := new(Feed)
	if err := xml.Unmarshal(contents, feed); err != nil {
		ErrorLogFunc("Failed to unmarshal XML: %v", err)
		return nil, err
	}

	return feed, nil
}
