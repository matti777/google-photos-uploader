package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	photos "mdahlbom/google-photos-uploader/googlephotos"
	"mdahlbom/google-photos-uploader/googlephotos/util"

	logging "github.com/op/go-logging"
	"github.com/urfave/cli"
)

// Our local logger
var log = logging.MustGetLogger("uploader")

var (
	// Application configuration
	appConfig *appConfiguration

	// Filename extensions to consider as images
	imageExtensions = []string{"jpg", "jpeg"}

	// Directory name -> photos folder substitution CSV string; should
	// be formatted as old1,new1,old2,new2, ... where new1 replaces old1 etc
	nameSubstitutionTokens = ""

	// Whether to capitalize words in directory name when forming
	// folder names
	capitalize = false

	// Whether to skip parsing folder year from the folder name
	noParseYear = false

	// Whether to skip (assume Yes) all confirmations)
	skipConfirmation = false

	// Whether doing a 'dry run', ie not actually sending anything.
	dryRun = false

	// Whether to skip reading journal files
	disregardJournal = false

	// Whether to recurse into subdirectories
	recurse = false

	// Maximum concurrency (number of simultaneous uploads)
	maxConcurrency = 1

	// Google Photos API client
	photosClient *photos.Client

	// // Feed representing the list of Albums
	// albumFeed *photos.Feed

	// List of albums
	albums []*photos.Album
)

func readFlags(c *cli.Context) {
	disregardJournal = GlobalBoolT(c, "disregard-journal")
	if disregardJournal {
		log.Debugf("Disregarding reading journal files..")
	}

	recurse = GlobalBoolT(c, "recursive")
	log.Debugf("Recurse into subdirectories: %v", recurse)

	skipConfirmation = GlobalBoolT(c, "yes")

	dryRun = GlobalBoolT(c, "dry-run")
	if dryRun {
		fmt.Printf("--dry-run enabled, not uploading anything\n")
	}

	exts := c.String("extensions")
	if exts != "" {
		s := strings.Split(exts, ",")
		imageExtensions = make([]string, len(s))
		for i, item := range s {
			imageExtensions[i] = strings.ToLower(strings.Trim(item, " "))
		}
	}

	log.Debugf("Using image extensions: %v", imageExtensions)

	nameSubstitutionTokens = c.String("folder-name-substitutions")
	log.Debugf("Using folder name substitution tokens: %v",
		nameSubstitutionTokens)

	useDefaultSubs := GlobalBoolT(c, "default-substitutions")
	if useDefaultSubs {
		nameSubstitutionTokens = "_, ,-, - "
	}

	noParseYear = GlobalBoolT(c, "no-parse-year")
	log.Debugf("Skipping parsing folder year?: %v", noParseYear)

	capitalize = GlobalBoolT(c, "capitalize")
	log.Debugf("Capitalizing folder name words: %v", capitalize)

	maxConcurrency = c.Int("concurrency")
	log.Debugf("maxConcurrency = %v", maxConcurrency)
}

func defaultAction(c *cli.Context) error {
	log.Debugf("Running Default action..")

	readFlags(c)

	authorize := GlobalBoolT(c, "authorize")

	// Make sure we have an auth token, ie. the user has performed the
	// authorization flow.
	if appConfig.AuthToken == nil || appConfig.UserInfo == nil && !authorize {
		fmt.Printf("Not authorized; you must perform the authorization " +
			"flow. Run again and specify the --authorize flag.")
		return fmt.Errorf("Missing authorization")
	}

	baseDir := c.Args().Get(0)
	if baseDir == "" {
		log.Debugf("No base dir defined, using the CWD..")
		dir, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get CWD: %v", err)
		}
		baseDir = dir
	}

	log.Debugf("baseDir input: %v", baseDir)

	// Resolve the base dir
	baseDir, err := filepath.Abs(baseDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path for '%v': %v", err)
	}
	log.Debugf("Using baseDir: %v", baseDir)

	// Check if need to authenticate the user
	if authorize {
		log.Debugf("Running authorization flow..")
		a := util.NewAuthenticator(appConfig.ClientID,
			appConfig.ClientSecret)
		token, userInfo, err := a.Authorize()
		if err != nil {
			log.Fatalf("Failed to get authorization token")
		} else {
			fmt.Println("Authorization OK!")
			appConfig.AuthToken = token
			appConfig.UserInfo = userInfo
			mustWriteAppConfig(appConfig)
		}
	}

	fmt.Printf("Authorized as '%v' (%v) -- specify --authorize to authorize "+
		"on a different account.\n", appConfig.UserInfo.Name,
		appConfig.UserInfo.ID)

	// Create the API client
	photosClient, err = photos.NewClient(appConfig.ClientID,
		appConfig.ClientSecret, appConfig.AuthToken)
	if err != nil {
		return err
	}

	// Retrieve the list of albums
	fmt.Printf("Fetching the list of Photos Albums..\n")
	if l, err := photosClient.ListAlbums(); err != nil {
		log.Fatalf("Failed to list Google Photos albums: %v", err)
	} else {
		albums = l
	}

	for _, a := range albums {
		log.Debugf("Got Photos Album: '%v'", a.Title)
	}

	mustProcessDir(baseDir)

	return nil
}

func main() {
	// Set up logging
	setupLogging()

	appname := os.Args[0]
	log.Debugf("main(): running binary %v..", appname)

	appConfig = readAppConfig()
	log.Debugf("Read appConfig: %+v", appConfig)
	if appConfig.ClientID == "" || appConfig.ClientSecret == "" {
		appConfig.ClientID, appConfig.ClientSecret = mustReadAppCredentials()
		log.Debugf("Got appConfig from stdin: %+v", appConfig)
		mustWriteAppConfig(appConfig)
	}

	// Setup CLI app framework
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Name = appname
	app.ArgsUsage = "[directory]"
	app.Usage = "A command line Google Photos upload utility"
	app.UsageText = fmt.Sprintf("%v [options] directory", appname)
	app.Description = fmt.Sprintf("Command-line utility for uploading "+
		"photos to Google Photos from a local disk directory. "+
		"For help, run '%v help'", appname)
	app.Copyright = "(c) 2018 Matti Dahlbom"
	app.Version = "0.0.1-alpha"
	app.Action = defaultAction
	app.Flags = []cli.Flag{
		cli.BoolTFlag{
			Name: "authorize",
			Usage: "Trigger Google authorization flow. " +
				"You only have to run this one time; " +
				"after you have authenticated, the authentication token " +
				"will be stored. If you want to authenticate with another " +
				"account, simply define this flag again.",
		},
		cli.BoolTFlag{
			Name:  "disregard-journal, d",
			Usage: "Disregard reading journal files; re-upload everything",
		},
		cli.BoolTFlag{
			Name:  "recursive, r",
			Usage: "Process subdirectories recursively",
		},
		cli.BoolTFlag{
			Name:  "yes, y",
			Usage: "Answer Yes to all confirmations",
		},
		cli.BoolTFlag{
			Name:  "dry-run, n",
			Usage: "Specify to just scan, not actually upload anything",
		},
		cli.IntFlag{
			Name:  "concurrency, c",
			Usage: "Maximum number of simultaneous uploads",
			Value: 1,
		},
		cli.StringFlag{
			Name: "folder-name-substitutions, s",
			Usage: "Directory name -> Photos Folder substition " +
				"tokens, default is no substitution. " +
				"The format is CSV like so: old1,new1,old2,new2 where token " +
				"new1 would replace token old1 etc. For example to replace " +
				"all underscores with spaces and add spaces around " +
				"all dashes, specify -s \"_, ,-, - \"",
		},
		cli.BoolTFlag{
			Name:  "default-substitutions, -ds",
			Usage: "Same as defining -s \"_, ,-, - \"",
		},
		cli.BoolTFlag{
			Name: "no-parse-year, -ny",
			Usage: "Do not attempt to parse the year from the directory " +
				"name; by default, an attempt is made to extract Photos Folder " +
				"creation date using regex '.*[\\- _][0-9]{4}'. Eg. " +
				"'Pictures_from_Thailand-2009' would generate folder creation " +
				"year 2009.",
		},
		cli.BoolTFlag{
			Name: "capitalize, a",
			Usage: "When forming the Photos Folder names, capitalize the " +
				"first letter of each word, ie 'trip to tonga, 2018' " +
				"would become 'Trip To Tonga, 2018'. Combine with " +
				"folder-name-substitutions to clean up the directory names",
		},
		cli.StringFlag{
			Name: "extensions, e",
			Usage: "File extensions to consider as uploadable images;  " +
				"eg. \"jpg, jpeg, png\". Default is \"jpg, jpeg\"",
		},
	}

	app.Run(os.Args)
}
