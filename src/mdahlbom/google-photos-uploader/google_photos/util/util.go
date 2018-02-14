// Misc utility functions
package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	logging "github.com/op/go-logging"
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

// Our local logger
var log = logging.MustGetLogger("google-photos-utils")

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
		log.Errorf("Error fetching user information: %v", err)
		return nil, err
	}

	defer r.Body.Close()

	contents, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("Failed to read response body: %v", err)
		return nil, err
	}

	info := new(UserInfo)
	if err := json.Unmarshal(contents, info); err != nil {
		log.Errorf("Failed to unmarshal JSON: %v", err)
		return nil, err
	}

	log.Debugf("Got UserInfo: %+v", info)

	return info, nil
}

// Configures the local logger
func setupLogging() {
	var format = logging.MustStringFormatter("%{color}%{time:15:04:05.000} " +
		"%{shortfunc} ▶ %{level} " +
		"%{color:reset} %{message}")
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	formatter := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(formatter)
	if enableDebugLogging {
		logging.SetLevel(logging.DEBUG, "uploader")
	} else {
		logging.SetLevel(logging.INFO, "uploader")
	}
}

func init() {
	setupLogging()
}
