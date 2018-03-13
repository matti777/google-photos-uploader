// Misc utility functions
package util

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	// Google's user info endpoint URL
	userInfoEndpointUrlFmt = "https://www.googleapis.com/oauth2/v2/userinfo" +
		"?access_token=%v"
)

type UserInfo struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	GivenName string `json:"given_name"`
	LastName  string `json:"family_name"`
}

// Set this variable to enable error logging output from the library
var ErrorLogFunc func(string, ...interface{}) = func(string, ...interface{}) {}

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

func NewOAuth2Config(clientID, clientSecret string) oauth2.Config {
	return oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes: []string{
			"https://picasaweb.google.com/data/",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

// Retrieves user information from the Google user info endpoint with token
func GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	// Retrieve user info
	r, err := http.Get(fmt.Sprintf(userInfoEndpointUrlFmt, token.AccessToken))
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

	info := new(UserInfo)
	if err := json.Unmarshal(contents, info); err != nil {
		ErrorLogFunc("Failed to unmarshal JSON: %v", err)
		return nil, err
	}

	ErrorLogFunc("Got UserInfo: %+v", info)

	return info, nil
}

// Creates a file upload request. If callback parameter is specified,
// it will get called when data has been read (and thus submitted) from the
// reader.
func NewImageUploadRequest(uri, mimeType string,
	reader io.Reader, callback func(int64)) (*http.Request, error) {

	if callback != nil {
		// Wrap the reader into one supporting progress callbacks
		var closer io.Closer = nil
		c, ok := reader.(io.Closer)
		if ok {
			closer = c
		}
		reader = &sizeCountingReader{Reader: reader, Closer: closer,
			callback: callback, numBytesRead: 0}
	}

	req, err := http.NewRequest("POST", uri, reader)
	if err != nil {
		ErrorLogFunc("Failed to create HTTP request: %v", err)
		return nil, err
	}

	if mimeType != "" {
		req.Header.Set("Content-Type", mimeType)
	}

	return req, nil
}

// Creates a file upload request. If callback parameter is specified,
// it will get called when data has been read (and thus submitted) from the
// reader.
func NewImageUploadRequestFromFile(uri string,
	path string, callback func(int64)) (*http.Request, error) {

	f, err := os.Open(path)
	if err != nil {
		ErrorLogFunc("Failed to open file: %v", err)
		return nil, err
	}

	// Decide the MIME type by the file extension
	ext := filepath.Ext(path)
	mimeType := mime.TypeByExtension(ext)

	return NewImageUploadRequest(uri, mimeType, f, callback)
}
