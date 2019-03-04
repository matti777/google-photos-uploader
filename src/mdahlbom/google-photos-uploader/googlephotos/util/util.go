package util

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/photoslibrary/v1"
)

const (
	// Google's user info endpoint URL
	userInfoEndpointURLFmt = "https://www.googleapis.com/oauth2/v2/userinfo" +
		"?access_token=%v"

	// URL to upload photo data
	photoDataUploadURL = "https://photoslibrary.googleapis.com/v1/uploads"
)

// UserInfo represents a Google user
type UserInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	GivenName string `json:"given_name"`
	LastName  string `json:"family_name"`
}

// ErrorLogFunc is called to log any errors;
// set this variable to enable error logging output from the library
var ErrorLogFunc = func(string, ...interface{}) {}

// Wraps io.Reader (and io.Closer) so that it counts the bytes read.
type sizeCountingReader struct {
	io.Reader
	io.Closer

	callback     func(count int64)
	numBytesRead int64
}

func (r *sizeCountingReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)

	r.numBytesRead += int64(n)
	r.callback(r.numBytesRead)

	return n, err
}

func (r *sizeCountingReader) Close() error {
	if r.Closer != nil {
		return r.Closer.Close()
	} else {
		return nil
	}
}

// NewOAuth2Config creates a new OAuth2 configuration with client id + secret
func NewOAuth2Config(clientID, clientSecret string) oauth2.Config {
	return oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes: []string{
			// "https://picasaweb.google.com/data/",
			photoslibrary.PhotoslibraryAppendonlyScope,
			//TODO: find this as a constant
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

// GetUserInfo retrieves user information from the Google user info
// endpoint with token
func GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	// Retrieve user info
	r, err := http.Get(fmt.Sprintf(userInfoEndpointURLFmt, token.AccessToken))
	if err != nil {
		ErrorLogFunc("Error fetching user information: %v", err)
		return nil, err
	}

	defer r.Body.Close()

	contents, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ErrorLogFunc("Failed to read response body: %v", err)
		return nil, err
	}

	//TODO get email if it is there and remove this
	ErrorLogFunc("DO NOT LOG - user info response: %", string(contents))

	info := new(UserInfo)
	if err := json.Unmarshal(contents, info); err != nil {
		ErrorLogFunc("Failed to unmarshal JSON: %v", err)
		return nil, err
	}

	ErrorLogFunc("Got UserInfo: %+v", info)

	return info, nil
}

// NewImageUploadRequest creates a file upload request.
// If callback parameter is specified,
// it will get called when data has been read (and thus submitted) from the
// reader.
// func NewImageUploadRequest(mimeType string,
// 	reader io.Reader, callback func(int64)) (*http.Request, error) {

// 	return req, nil
// }

// NewImageUploadRequestFromFile creates a file upload request.
// If callback parameter is specified,
// it will get called when data has been read (and thus submitted) from the
// reader.
func NewImageUploadRequestFromFile(inputFilePath string,
	callback func(int64)) (*http.Request, error) {

	f, err := os.Open(inputFilePath)
	if err != nil {
		ErrorLogFunc("Failed to open file: %v", err)
		return nil, err
	}

	// if callback != nil {
	// 	// Wrap the reader into one supporting progress callbacks
	// 	var closer io.Closer
	// 	c, ok := reader.(io.Closer)
	// 	if ok {
	// 		closer = c
	// 	}
	// 	reader = &sizeCountingReader{Reader: reader, Closer: closer,
	// 		callback: callback, numBytesRead: 0}
	// }

	reader := &sizeCountingReader{Reader: f, Closer: f,
		callback: callback, numBytesRead: 0}

	req, err := http.NewRequest("POST", photoDataUploadURL, reader)
	if err != nil {
		ErrorLogFunc("Failed to create HTTP request: %v", err)
		return nil, err
	}

	/* To upload using Google Photos API:

			POST https://photoslibrary.googleapis.com/v1/uploads
	Authorization: Bearer OAUTH2_TOKEN
	Content-type: application/octet-stream
	X-Goog-Upload-File-Name: FILENAME
	X-Goog-Upload-Protocol: raw
	*/

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-Goog-Upload-File-Name", filepath.Base(inputFilePath))
	req.Header.Set("X-Goog-Upload-Protocol", "raw")

	return req, nil
}
