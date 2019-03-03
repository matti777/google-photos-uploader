// Manages authenticating to Google and providing an OAuth2 token
package util

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
)

// Authenticator is our authenticator type. Construct using NewAuthenticator().
// This class implements the Google OAuth2 authorization code flow; for
// more info, see:
// https://developers.google.com/actions/identity/oauth2-code-flow
type Authenticator struct {
	oauth2Config oauth2.Config

	stateToken string
	ch         chan *oauth2.Token
	errCh      chan error
}

// Opens a browser window to the specified URL
func openBrowser(url string) error {
	var err error

	switch runtime.GOOS {
	case "linux":
	case "freebsd":
	case "netbsd":
	case "openbsd":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32",
			"url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	return err
}

// NewAuthenticator creates a new Authenticator object with given
// client id + secret.
func NewAuthenticator(clientID, clientSecret string) *Authenticator {
	return &Authenticator{oauth2Config: NewOAuth2Config(clientID, clientSecret),
		ch:    make(chan *oauth2.Token),
		errCh: make(chan error),
	}
}

// HTTP handler for the Google's auth callbacks
func (a *Authenticator) auth(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		ErrorLogFunc("Failed to parse form: %v", err)
		a.errCh <- fmt.Errorf("Invalid request")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Get 'state' which is our self-generated nonce token
	state := r.FormValue("state")

	// Get the authorization code
	code := r.FormValue("code")

	// Both values must be supplied
	if state == "" || code == "" {
		ErrorLogFunc("Missing value(s) for state/code: %v/%v", state, code)
		a.errCh <- fmt.Errorf("Missing request parameters")
		http.Error(w, "missing values", http.StatusBadRequest)
		return
	}

	// Make sure the state token matches
	if state != a.stateToken {
		ErrorLogFunc("Invalid OAuth2 state received; expected '%s' "+
			"but got '%s'", a.stateToken, state)
		a.errCh <- fmt.Errorf("invalid OAuth2 state")
		http.Error(w, "invalid state", http.StatusUnauthorized)
		return
	}

	// Exchange the authorization code for an access token
	token, err := a.oauth2Config.Exchange(oauth2.NoContext, code)
	if err != nil {
		ErrorLogFunc("Code exchange failed: %v", code)
		a.errCh <- fmt.Errorf("Code exchange failed: %v", code)
		http.Error(w, "Code exchange failed", http.StatusInternalServerError)
		return
	}

	// Done; send the token for the waiting routine
	a.ch <- token

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w,
		"<h1>Authentication successful, you may close this window</h1>")
}

// Starts listening to HTTP requests. Returns the listening address:port,
// eg. localhost:12345
func (a *Authenticator) listenToHTTP(pathNonce string) (string, error) {
	l, err := net.Listen("tcp", "localhost:")
	if err != nil {
		ErrorLogFunc("Failed to Listen(): %v", err)
		return "", err
	}
	ErrorLogFunc("Listening at %v", l.Addr().String())

	r := mux.NewRouter()
	path := fmt.Sprintf("/auth/%v", pathNonce)
	r.HandleFunc(path, a.auth).Methods("GET")

	go func() {
		http.Handle("/", r)
		http.Serve(l, r)
	}()

	return l.Addr().String(), nil
}

// Authorize synchronously waits for an access token.
func (a *Authenticator) Authorize() (*oauth2.Token, *UserInfo, error) {
	appname := os.Args[0]
	fmt.Printf("%v needs to authorize to access Google Photos. "+
		"Opening a browser to perform this step..\n\n", appname)

	nonce := uuid.New().String()
	ErrorLogFunc("Allocated a path nonce: %v", nonce)

	a.stateToken = uuid.New().String()
	ErrorLogFunc("Allocated state token: %v", a.stateToken)

	addr, err := a.listenToHTTP(nonce)
	if err != nil {
		return nil, nil, err
	}

	a.oauth2Config.RedirectURL = fmt.Sprintf("http://%v/auth/%v",
		addr, nonce)

	// Retrieve an URL where the user can authorize this app and open
	// that URL in a browser
	url := a.oauth2Config.AuthCodeURL(a.stateToken, oauth2.AccessTypeOffline)

	fmt.Printf("If the browser window fails to open, open the following URL "+
		"manually in your favourite browser:\n\n%v\n\n", url)
	if err := openBrowser(url); err != nil {
		ErrorLogFunc("Failed to open web browser: %v", err)
		return nil, nil, err
	}

	// Wait for the code on the channel; the HTTP handler will send it
	select {
	case token := <-a.ch:
		fmt.Println("Authorization OK.")

		info, err := GetUserInfo(token)

		return token, info, err
	case err := <-a.errCh:
		return nil, nil, err
	}
}
