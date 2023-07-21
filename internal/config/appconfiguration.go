package config

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
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
	ClientID     string              `json:"clientId"`
	ClientSecret string              `json:"clientSecret"`
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
	log.Debugf("Reading application configuration file %v", appCfgFilePath)

	file, err := os.Open(appCfgFilePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// This is OK, it wont exist on first run
			log.Debugf("Application configuration file not found.")
			return &cfg
		}

		log.Errorf("Failed to open app cfg file: %v", err)
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
	log.Debugf("Writing application configuration file %v", appCfgFilePath)

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
		"get these, go to https://console.developers.google.com/ "+
		"for your GCP project, navigate to Credentials and select "+
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
