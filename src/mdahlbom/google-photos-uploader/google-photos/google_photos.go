// Manages authenticating to Google and providing an OAuth2 token
package google_photos

import (
	"net/http"

	"mdahlbom/google-photos-uploader/google-photos/util"

	"golang.org/x/oauth2"
)

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
