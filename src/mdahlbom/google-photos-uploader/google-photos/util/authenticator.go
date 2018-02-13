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
	logging "github.com/op/go-logging"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Our authenticator type. Construct using NewAuthenticator().
// This class implements the Google OAuth2 authorization code flow; for
// more info, see:
// https://developers.google.com/actions/identity/oauth2-code-flow
type Authenticator struct {
	oauth2Config *oauth2.Config

	stateToken string
	ch         chan *oauth2.Token
	errCh      chan error
}

// Our local logger
var log = logging.MustGetLogger("google-photos-utils")

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

func NewAuthenticator(clientID, clientSecret,
	username string) *Authenticator {

	a := &Authenticator{oauth2Config: &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"https://picasaweb.google.com/data/"},
		Endpoint:     google.Endpoint,
	},
		ch:    make(chan *oauth2.Token),
		errCh: make(chan error),
	}

	return a
}

// HTTP handler for the Google's auth callbacks
func (a *Authenticator) auth(w http.ResponseWriter, r *http.Request) {
	log.Debugf("auth(): request: %+v", r)

	// Get 'state' which is our self-generated nonce token
	state := r.FormValue("state")

	// Get the authorization code
	code := r.FormValue("code")

	// Both values must be supplied
	if state == "" || code == "" {
		log.Errorf("Missing value(s) for state/code: %v/%v", state, code)
		a.errCh <- fmt.Errorf("Missing request parameters")
		http.Error(w, "missing values", http.StatusBadRequest)
		return
	}

	// Make sure the state token matches
	if state != a.stateToken {
		log.Errorf("Invalid OAuth2 state received; expected '%s' "+
			"but got '%s'", a.stateToken, state)
		a.errCh <- fmt.Errorf("invalid OAuth2 state")
		http.Error(w, "invalid state", http.StatusUnauthorized)
		return
	}

	// Exchange the authorization code for an access token
	token, err := a.oauth2Config.Exchange(oauth2.NoContext, code)
	if err != nil {
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
func (a *Authenticator) listenToHttp(pathNonce string) (string, error) {
	l, err := net.Listen("tcp", "localhost:")
	if err != nil {
		log.Errorf("Failed to Listen(): %v", err)
		return "", err
	}
	log.Debugf("Listening at %v", l.Addr().String())

	r := mux.NewRouter()
	path := fmt.Sprintf("/auth/%v", pathNonce)
	r.HandleFunc(path, a.auth).Methods("GET")

	go func() {
		log.Debugf("Starting HTTP serving..")
		http.Handle("/", r)
		http.Serve(l, r)
		log.Debugf("HTTP started.")
	}()
	log.Debugf("HTTP serving..")

	return l.Addr().String(), nil
}

// Synchronously waits for an access token.
func (a *Authenticator) GetToken() (*oauth2.Token, error) {
	appname := os.Args[0]
	fmt.Printf("%v needs to authorize to access Google Photos. "+
		"Opening a browser to perform this step..", appname)

	nonce := uuid.New().String()
	log.Debugf("Allocated a path nonce: %v", nonce)

	a.stateToken = uuid.New().String()
	log.Debugf("Allocated state token: %v", a.stateToken)

	addr, err := a.listenToHttp(nonce)
	if err != nil {
		return nil, err
	}

	a.oauth2Config.RedirectURL = fmt.Sprintf("http://%v/auth/%v",
		addr, nonce)
	log.Debugf("RedirectURL: %v", a.oauth2Config.RedirectURL)

	// Retrieve an URL where the user can authorize this app and open
	// that URL in a browser
	url := a.oauth2Config.AuthCodeURL(a.stateToken, oauth2.AccessTypeOffline)

	fmt.Printf("If the browser window fails to open, open the following URL "+
		"manually in your favourite browser:\n\n%v\n\n", url)
	if err := openBrowser(url); err != nil {
		log.Errorf("Failed to open web browser: %v", err)
		return nil, err
	}

	// Wait for the code on the channel; the HTTP handler will send it
	select {
	case token := <-a.ch:
		log.Debugf("Got token: %v", token)
		fmt.Println("Authorization OK.")
		return token, nil
	case err := <-a.errCh:
		return nil, err
	}
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
