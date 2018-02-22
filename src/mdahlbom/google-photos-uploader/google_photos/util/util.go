// Misc utility functions
package util

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// Creates a file upload request
func NewImageUploadRequest(uri, mimeType string,
	data []byte) (*http.Request, error) {

	// Allocate a buffer for storing the multipart body
	bodyReader := bytes.NewReader(data)

	req, err := http.NewRequest("POST", uri, bodyReader)
	if err != nil {
		ErrorLogFunc("Failed to create HTTP request: %v", err)
		return nil, err
	}

	if mimeType != "" {
		req.Header.Set("Content-Type", mimeType)
	}

	return req, nil
}

// Creates a file upload request
func NewImageUploadRequestFromFile(uri string,
	path string) (*http.Request, error) {

	f, err := os.Open(path)
	if err != nil {
		ErrorLogFunc("Failed to open file: %v", err)
		return nil, err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		ErrorLogFunc("Failed to read file: %v", err)
		return nil, err
	}

	// Decide the MIME type by the file extension
	ext := filepath.Ext(path)
	mimeType := mime.TypeByExtension(ext)

	return NewImageUploadRequest(uri, mimeType, data)
}
