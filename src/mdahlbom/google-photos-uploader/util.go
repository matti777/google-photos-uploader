package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	photos "mdahlbom/google-photos-uploader/googlephotos"
	"mdahlbom/google-photos-uploader/googlephotos/util"

	logging "github.com/op/go-logging"
	"github.com/urfave/cli"
	"golang.org/x/oauth2"
)

// Application configuration file structure
type appConfiguration struct {
	ClientID     string
	ClientSecret string
	AuthToken    *oauth2.Token
	UserInfo     *util.UserInfo
}

const (
	// App configuration file name
	appConfigFilename = ".photos-uploader.config"
)

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

	// Hook up the library's error logger functions to the error log
	photos.ErrorLogFunc = log.Errorf
	util.ErrorLogFunc = log.Errorf
}

// Returns the path to the app config file. Panics on failure.
func mustGetAppConfigPath() string {
	u, err := user.Current()
	if err != nil {
		log.Fatalf("Failed to get current user: %v", err)
	}

	return filepath.Join(u.HomeDir, appConfigFilename)
}

// Reads the app configuration file. If the file is not found, returns
// an empty config
func readAppConfig() *appConfiguration {
	appCfgFilePath := mustGetAppConfigPath()
	log.Debugf("Got app config file path: %v", appCfgFilePath)

	file, err := os.Open(appCfgFilePath)
	if err != nil {
		log.Debugf("Failed to open app cfg file; perhaps it doesnt "+
			"exist? error: %v", err)
		return &appConfiguration{}
	}

	decoder := json.NewDecoder(file)
	cfg := new(appConfiguration)
	if err := decoder.Decode(cfg); err != nil {
		log.Errorf("Failed to read app config file: %v", err)
		return &appConfiguration{}
	}

	return cfg
}

// Write the app configuration file. Panics on failure.
func mustWriteAppConfig(c *appConfiguration) {
	appCfgFilePath := mustGetAppConfigPath()
	log.Debugf("Using app config file path: %v", appCfgFilePath)

	file, err := os.Create(appCfgFilePath)
	if err != nil {
		log.Fatalf("Failed to open app cfg file for writing: %v", err)
	}

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(c); err != nil {
		log.Fatalf("Failed to write app cfg file: %v", err)
	}
}

// Reads the app credentials (ClientID and ClientSecret) from stdin
func mustReadAppCredentials() (string, string) {
	appCfgFilePath := mustGetAppConfigPath()
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

// Asks the user interactively a confirmation question; if the user declines
// (answers anything but Y or defalut - empty string - stop execution.
func mustConfirm(format string, args ...interface{}) {
	if skipConfirmation {
		return
	}

	text := fmt.Sprintf(format, args...)
	fmt.Print(fmt.Sprintf("%v\nContinue? [Y/n] ", text))
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')

	if input != "Y\n" && input != "\n" {
		os.Exit(1)
	}
}

// Finds the longest file name
func findLongestName(infos []os.FileInfo) int {
	longest := 0

	for _, info := range infos {
		if len(info.Name()) > longest {
			longest = len(info.Name())
		}
	}

	return longest
}

// There seems to be a bug in the GolbalBoolT API, or I am using
// it incorrectly - however, this method checks for the presence of a BoolT
// flag and returns false if it is not specified.
func GlobalBoolT(c *cli.Context, name string) bool {
	if !c.IsSet(name) {
		return false
	} else {
		return c.GlobalBoolT(name)
	}
}

// Replaces substrings in the string with other strings, using strings.Replacer.
// The tokens parameter should be formatted as a valid CSV.
func replaceInString(s, tokens string) (string, error) {
	if tokens == "" {
		return s, nil
	}

	r := csv.NewReader(strings.NewReader(tokens))
	records, err := r.ReadAll()
	if err != nil {
		log.Errorf("Failed to read CSV: %v", err)
		return "", err
	}

	if len(records) != 1 {
		log.Errorf("Invalid number of CSV records: %v", len(records))
		return "", errors.New("Invalid number of CSV records. " +
			"Only single line CSV supported.")
	}

	tokenArray := records[0]
	replacer := strings.NewReplacer(tokenArray...)

	return replacer.Replace(s), nil
}
