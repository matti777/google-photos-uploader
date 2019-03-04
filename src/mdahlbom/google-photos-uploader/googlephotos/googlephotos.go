// Package google_photos contains all the types used with communication
// with Google Photos API
package googlephotos

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"mdahlbom/google-photos-uploader/googlephotos/util"

	"golang.org/x/oauth2"
	"google.golang.org/api/photoslibrary/v1"
)

// ErrorLogFunc is an optional logging function.
// Set this variable to enable error logging output from the library
var ErrorLogFunc func(string, ...interface{})

// Client is Our API client type. Create with NewClient().
type Client struct {
	httpClient   *http.Client
	photosClient *photoslibrary.Service
}

// NewClient creates a new API client using an OAuth2 token. To acquire the
// token, run the authorization flow with util.Authenticator.
func NewClient(clientID, clientSecret string,
	token *oauth2.Token) (*Client, error) {

	config := util.NewOAuth2Config(clientID, clientSecret)
	httpClient := config.Client(oauth2.NoContext, token)

	photosClient, err := photoslibrary.New(httpClient)
	if err != nil {
		ErrorLogFunc("Failed to create Photos API client: %v", err)
		return nil, err
	}

	return &Client{photosClient: photosClient}, nil
}

// ListAlbums Lists all the Albums
func (c *Client) ListAlbums() ([]*Album, error) {
	var res *photoslibrary.ListAlbumsResponse
	done := false
	albums := make([]*Album, 0)

	for !done {
		req := c.photosClient.Albums.List().PageSize(50)
		if res != nil && res.NextPageToken != "" {
			req = req.PageToken(res.NextPageToken)
		}

		res, err := req.Do()
		if err != nil {
			ErrorLogFunc("Failed to list albums: %v", err)
			return nil, err
		}

		for _, a := range res.Albums {
			albums = append(albums, &Album{ID: a.Id, Title: a.Title})
		}

		if res.NextPageToken == "" {
			done = true
		}
	}

	return albums, nil
}

// UploadPhoto uploads a photo to an album synchronously.
// If callback parameter is specified,
// it will get called when data has been submitted.
// Returns either an upload token or an error.
func (c *Client) UploadPhoto(path string, album *Album,
	callback func(int64)) (string, error) {

	/* To upload using Google Photos API:

			POST https://photoslibrary.googleapis.com/v1/uploads
	Authorization: Bearer OAUTH2_TOKEN
	Content-type: application/octet-stream
	X-Goog-Upload-File-Name: FILENAME
	X-Goog-Upload-Protocol: raw

	Request body:
	BINARY FILE

	Response body:
	UPLOAD TOKEN

	Then, see batchCreate method of google.golang.org/api/photoslibrary/v1
	*/

	// url := fmt.Sprintf("https://picasaweb.google.com/data/feed/api/user/"+
	// 	"default/albumid/%v", album.ID)

	req, err := util.NewImageUploadRequestFromFile(path, callback)
	if err != nil {
		return "", err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		ErrorLogFunc("Failed to POST new image: %v", err)
		return "", err
	}

	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		ErrorLogFunc("Got non-OK status code: %v", res.StatusCode)
		return "", fmt.Errorf("Photo upload failed: %v", res.Status)
	}

	// In a success response, the response body should hold a single line of
	// text which is the upload token for the image.
	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		ErrorLogFunc("Failed to read response body: %v", err)
		return "", err
	}

	uploadToken := strings.Trim(string(contents), "\n ")

	ErrorLogFunc("TODO: remove: upload token: %v", uploadToken)

	return uploadToken, nil
}

// Returns a Feed for the given resource (endpoint URL)
// func (c *Client) fetchFeed(endpoint string) (*Feed, error) {
// 	req, err := http.NewRequest("GET", endpoint, nil)
// 	if err != nil {
// 		ErrorLogFunc("Failed to create HTTP request: %v", err)
// 		return nil, err
// 	}

// 	req.Header.Set("GData-Version", "3")

// 	res, err := c.httpClient.Do(req)
// 	if err != nil {
// 		ErrorLogFunc("Failed to fetch the feed: %v", err)
// 		return nil, err
// 	}

// 	defer res.Body.Close()

// 	if res.StatusCode != http.StatusOK {
// 		ErrorLogFunc("Got non-OK response code: %v", res.StatusCode)
// 		return nil, errors.New(res.Status)
// 	}

// 	contents, err := ioutil.ReadAll(res.Body)
// 	if err != nil {
// 		ErrorLogFunc("Failed to read response body: %v", err)
// 		return nil, err
// 	}

// 	feed := new(Feed)
// 	if err := xml.Unmarshal(contents, feed); err != nil {
// 		ErrorLogFunc("Failed to unmarshal XML: %v", err)
// 		return nil, err
// 	}

// 	return feed, nil
// }
