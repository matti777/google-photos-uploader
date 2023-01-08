package util

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	photoslibrary "google.golang.org/api/photoslibrary/v1"
)

const (
	// Google's user info endpoint URL
	userInfoEndpointURLFmt = "https://www.googleapis.com/oauth2/v3/userinfo" +
		"?access_token=%v"

	// URL to upload photo data
	photoDataUploadURL = "https://photoslibrary.googleapis.com/v1/uploads"
)

// UserInfo represents a Google user. The JSON field names are dictated by
// the Google userinfo API.
type UserInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	GivenName string `json:"given_name"`
	LastName  string `json:"family_name"`
	Email     string `json:"email"`
}

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

	scopes := []string{
		"https://www.googleapis.com/auth/userinfo.profile",
		photoslibrary.PhotoslibraryScope,
	}

	return oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}
}

// GetUserInfo retrieves user information from the Google user info
// endpoint with token
func GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	// fmt.Sprintf("access token: %+v\n", token)

	// Retrieve user info
	url := fmt.Sprintf(userInfoEndpointURLFmt, token.AccessToken)
	// fmt.Printf("Getting UserInfo from: %v\n", url)

	res, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching user information")
	}

	defer res.Body.Close()

	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}
	fmt.Printf("Read userinfo contents: %v\n", string(contents))

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, errors.Errorf("failed to call GetUserInfo: %v", string(contents))
	}

	info := new(UserInfo)
	if err := json.Unmarshal(contents, info); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal JSON")
	}
	// fmt.Printf("Received UserInfo: %+v\n", info)

	return info, nil
}

// NewImageUploadRequestFromFile creates a file upload request.
// If callback parameter is specified,
// it will get called when data has been read (and thus submitted) from the
// reader.
func NewImageUploadRequestFromFile(inputFilePath string,
	callback func(int64)) (*http.Request, error) {

	f, err := os.Open(inputFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}

	reader := &sizeCountingReader{Reader: f, Closer: f,
		callback: callback, numBytesRead: 0}

	req, err := http.NewRequest("POST", photoDataUploadURL, reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create upload request")
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-Goog-Upload-File-Name", filepath.Base(inputFilePath))
	req.Header.Set("X-Goog-Upload-Protocol", "raw")

	return req, nil
}
