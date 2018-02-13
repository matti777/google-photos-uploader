// Manages authenticating to Google and providing an OAuth2 token
package google-photos

import (
	"net/http"
)

// Our API client type. Create with NewClient().
type Client struct {
	httpClient *http.Client
}

// Creates a new API client using an OAuth2 token. To acquire the token,
// run the authorization flow with util.Authenticator.
func NewClient(token *oauth2.Token) (*Client, error) {
	return &Client{httpClient: 
}



