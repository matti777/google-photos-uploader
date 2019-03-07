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

const (
	// MaxAddPhotosPerCall is the maximum number of photos to add to
	// an album in one single call.
	MaxAddPhotosPerCall = 50
)

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
		return nil, err
	}

	return &Client{photosClient: photosClient, httpClient: httpClient}, nil
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

// CreateAlbum creates a new album
func (c *Client) CreateAlbum(name string) (*Album, error) {
	req := &photoslibrary.CreateAlbumRequest{
		Album: &photoslibrary.Album{
			Title: name,
		},
	}

	album, err := c.photosClient.Albums.Create(req).Do()
	if err != nil {
		return nil, err
	}

	return &Album{ID: album.Id, Title: album.Title}, nil
}

// AddToAlbum adds the photos identified by their upload tokens to the album
func (c *Client) AddToAlbum(album *Album, uploadTokens []string) error {
	if len(uploadTokens) > MaxAddPhotosPerCall {
		return fmt.Errorf("Maximum number of photos to add per call is %v",
			MaxAddPhotosPerCall)
	}

	mediaItems := make([]*photoslibrary.NewMediaItem, len(uploadTokens))
	for i, tok := range uploadTokens {
		mediaItems[i] = &photoslibrary.NewMediaItem{
			SimpleMediaItem: &photoslibrary.SimpleMediaItem{
				UploadToken: tok,
			},
		}
	}

	req := &photoslibrary.BatchCreateMediaItemsRequest{
		AlbumId:       album.ID,
		NewMediaItems: mediaItems,
	}

	res, err := c.photosClient.MediaItems.BatchCreate(req).Do()
	if err != nil {
		return err
	}

	numFailed := 0

	for _, r := range res.NewMediaItemResults {
		if r.MediaItem == nil {
			fmt.Printf("Failed to add a photo to the album with token: %v: %v\n",
				r.UploadToken, r.Status.Message)
			numFailed++
		}
	}

	if numFailed == len(uploadTokens) {
		// All failed to add
		return fmt.Errorf("Failed to add all of the photos to album")
	}

	// At least some photos added successfully
	return nil
}

// UploadPhoto uploads a photo to an album synchronously.
// If callback parameter is specified,
// it will get called when data has been submitted.
// Returns either an upload token or an error.
func (c *Client) UploadPhoto(path string,
	callback func(int64)) (string, error) {

	req, err := util.NewImageUploadRequestFromFile(path, callback)
	if err != nil {
		return "", err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Failed to POST new image: %v", err)
	}

	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", fmt.Errorf("Photo upload failed: %v", res.Status)
	}

	// In a success response, the response body should hold a single line of
	// text which is the upload token for the image.
	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read Photos API response: %v", err)
	}

	uploadToken := strings.Trim(string(contents), "\n ")

	return uploadToken, nil
}
