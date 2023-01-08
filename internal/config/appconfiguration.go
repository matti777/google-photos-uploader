package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2"

	photosutil "github.com/matti777/google-photos-uploader/internal/googlephotos/util"
	"github.com/matti777/google-photos-uploader/internal/logging"
)

// Application configuration file structure
type AppConfiguration struct {
	ClientID     string              `json:"-"`
	ClientSecret string              `json:"-"`
	AuthToken    *oauth2.Token       `json:"authToken"`
	UserInfo     photosutil.UserInfo `json:"userInfo"`
}

const (
	// App configuration file name
	appConfigFilename = ".photos-uploader.config"
)

var (
	log = logging.MustGetLogger()
)

// Returns the path to the app config file. Panics on failure.
func MustGetAppConfigPath() string {
	u, err := user.Current()
	if err != nil {
		log.Fatalf("Failed to get current user: %v", err)
	}

	return filepath.Join(u.HomeDir, appConfigFilename)
}

// Reads the app configuration file. If the file is not found, returns
// an empty config
func ReadAppConfig() *AppConfiguration {
	var cfg AppConfiguration

	appCfgFilePath := MustGetAppConfigPath()
	log.Debugf("Got app config file path: %v", appCfgFilePath)

	file, err := os.Open(appCfgFilePath)
	if err != nil {
		log.Debugf("Failed to open app cfg file; perhaps it doesnt "+
			"exist? error: %v", err)
		return &cfg
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		log.Errorf("Failed to read app config file: %v", err)
		return &AppConfiguration{}
	}

	return &cfg
}

// Write the app configuration file. Panics on failure.
func MustWriteAppConfig(c *AppConfiguration) {
	appCfgFilePath := MustGetAppConfigPath()
	log.Debugf("Using app config file path: %v", appCfgFilePath)

	file, err := os.Create(appCfgFilePath)
	if err != nil {
		log.Fatalf("Failed to open app cfg file for writing: %v", err)
	}

	// Make sure the file is not readable by others
	if err := os.Chmod(appCfgFilePath, 0600); err != nil {
		log.Fatalf("Failed to chmod config file: %v", err)
	}

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(c); err != nil {
		log.Fatalf("Failed to write app cfg file: %v", err)
	}
}

// Reads the app credentials (ClientID and ClientSecret) from stdin
func MustReadAppCredentials() (string, string) {
	appCfgFilePath := MustGetAppConfigPath()
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("\nYou must enter the application credentials.\n\n"+
		"Enter the ClientID and Client Secret for the app. To "+
		"get these, go to https://console.developers.google.com/ and "+
		"create a new project, navigate to Credentials and select "+
		"Create credentials > OAuth client ID.\n\n"+
		"The credentials will be stored in the app configuration file %v.\n\n",
		appCfgFilePath)

	clientID := ""
	for clientID == "" {
		fmt.Print("Enter the ClientID: ")
		clientID, _ = reader.ReadString('\n')
		clientID = strings.Trim(clientID, " \n")
	}

	clientSecret := ""
	for clientSecret == "" {
		fmt.Print("Enter the Client Secret: ")
		clientSecret, _ = reader.ReadString('\n')
		clientSecret = strings.Trim(clientSecret, " \n")
	}

	log.Debug("Read app credentials: %v, %v", clientID, clientSecret)

	return clientID, clientSecret
}
