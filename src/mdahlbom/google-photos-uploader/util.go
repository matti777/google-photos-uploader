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
	"regexp"
	"strings"

	"mdahlbom/google-photos-uploader/googlephotos/util"

	logging "github.com/op/go-logging"
	"github.com/urfave/cli"
	"golang.org/x/oauth2"
)

// Application configuration file structure
type appConfiguration struct {
	ClientID     string        `json:"clientId"`
	ClientSecret string        `json:"clientSecret"`
	AuthToken    *oauth2.Token `json:"authToken"`
	UserInfo     util.UserInfo `json:"userInfo"`
}

const (
	// App configuration file name
	appConfigFilename = ".photos-uploader.config"
)

var (
	albumYearRegex *regexp.Regexp
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

// GlobalBoolT checks for the presence of a BoolT
// flag and returns false if it is not specified.
func GlobalBoolT(c *cli.Context, name string) bool {
	if !c.IsSet(name) {
		return false
	}

	return c.GlobalBoolT(name)
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

// chunked returns an array of arrays so that the original array is divided
// into chunks of equal size (except for the remainder chunk).
func chunked(arr []string, chunkSize int) [][]string {
	arrayLen := len(arr)
	numChunks := arrayLen / chunkSize
	if arrayLen%chunkSize > 0 {
		numChunks++
	}

	chunks := make([][]string, 0, numChunks)

	for i := 0; i < arrayLen; i += chunkSize {
		chunkEnd := i + chunkSize
		if chunkEnd > arrayLen {
			chunkEnd = arrayLen
		}
		chunks = append(chunks, arr[i:chunkEnd])
	}

	return chunks
}

// parseAlbumYear tries to parse the year of the album from a string
// (directory name) using a certain regex pattern (eg. 'Trip to X - 2009')
// Returns empty string if not found.
func parseAlbumYear(name string) string {
	res := albumYearRegex.FindStringSubmatch(name)

	if len(res) == 2 {
		return res[1]
	}

	return ""
}

func init() {
	albumYearRegex = regexp.MustCompile("^.+[-_ ]([12]\\d{3})$")
}
