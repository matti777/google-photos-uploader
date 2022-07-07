package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/matti777/google-photos-uploader/internal/config"
	"github.com/matti777/google-photos-uploader/internal/files"
	photos "github.com/matti777/google-photos-uploader/internal/googlephotos"
	photosutil "github.com/matti777/google-photos-uploader/internal/googlephotos/util"
	"github.com/matti777/google-photos-uploader/internal/logging"
	"github.com/matti777/google-photos-uploader/internal/util"

	"github.com/urfave/cli"
)

var (
	log      = logging.MustGetLogger()
	settings = config.MustGetSettings()

	// Application configuration
	appConfig *config.AppConfiguration
)

func readFlags(c *cli.Context) {
	settings.DisregardJournal = util.GlobalBoolT(c, "disregard-journal")
	if settings.DisregardJournal {
		log.Debugf("Disregarding reading journal files..")
	}

	settings.Recurse = util.GlobalBoolT(c, "recursive")
	log.Debugf("Recurse into subdirectories: %v", settings.Recurse)

	settings.SkipConfirmation = util.GlobalBoolT(c, "yes")

	settings.DryRun = util.GlobalBoolT(c, "dry-run")
	if settings.DryRun {
		log.Debugf("--dry-run enabled, not uploading anything")
	}

	exts := c.String("extensions")
	if exts != "" {
		s := strings.Split(exts, ",")
		settings.ImageExtensions = make([]string, len(s))
		for i, item := range s {
			settings.ImageExtensions[i] = strings.ToLower(strings.Trim(item, " "))
		}
	}

	log.Debugf("Using image extensions: %v", settings.ImageExtensions)

	settings.NameSubstitutionTokens = c.String("folder-name-substitutions")
	log.Debugf("Using folder name substitution tokens: %v",
		settings.NameSubstitutionTokens)

	useDefaultSubs := util.GlobalBoolT(c, "default-substitutions")
	if useDefaultSubs {
		settings.NameSubstitutionTokens = "_, "
	}

	settings.NoParseYear = util.GlobalBoolT(c, "no-parse-year")
	log.Debugf("Skipping parsing folder year?: %v", settings.NoParseYear)

	settings.Capitalize = util.GlobalBoolT(c, "capitalize")
	log.Debugf("Capitalizing folder name words: %v", settings.Capitalize)

	settings.MaxConcurrency = c.Int("concurrency")
	log.Debugf("maxConcurrency = %v", settings.MaxConcurrency)
}

func defaultAction(c *cli.Context) error {
	log.Debugf("Running Default action..")

	readFlags(c)

	authorize := util.GlobalBoolT(c, "authorize")

	// Make sure we have an auth token, ie. the user has performed the
	// authorization flow.
	if (appConfig.AuthToken == nil || appConfig.UserInfo.ID == "") &&
		!authorize {
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

	// Resolve the base dir
	baseDir, err := filepath.Abs(baseDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path for '%v': %v", err)
	}
	log.Debugf("Using baseDir: %v", baseDir)

	// Check if need to authenticate the user
	if authorize {
		log.Debugf("Running authorization flow..")
		a := photosutil.NewAuthenticator(appConfig.ClientID,
			appConfig.ClientSecret)
		token, userInfo, err := a.Authorize()
		if err != nil {
			log.Fatalf("Failed to get authorization token")
		} else {
			fmt.Println("Authorization OK!")
			appConfig.AuthToken = token
			appConfig.UserInfo = *userInfo
			log.Debugf("UserInfo: %+v", appConfig.UserInfo)
			config.MustWriteAppConfig(appConfig)
		}
	}

	fmt.Printf("Authorized as '%v' (%v) -- specify --authorize to authorize "+
		"on a different account.\n", appConfig.UserInfo.Name,
		appConfig.UserInfo.ID)

	photosClient := photos.MustCreateClient(appConfig.ClientID, appConfig.ClientSecret,
		appConfig.AuthToken)

	// Retrieve the list of albums
	fmt.Printf("Fetching the list of Photos Albums..\n")
	if l, err := photosClient.ListAlbums(); err != nil {
		log.Fatalf("Failed to list Google Photos albums: %v", err)
	} else {
		settings.Albums = l
	}

	for _, a := range settings.Albums {
		log.Debugf("Got Photos Album: '%v'", a.Title)
	}

	files.MustProcessDir(baseDir)

	return nil
}

func main() {
	appname := os.Args[0]
	log.Debugf("main(): running binary %v..", appname)

	appConfig = config.ReadAppConfig()
	log.Debugf("Read user info: %+v", appConfig.UserInfo)
	if appConfig.ClientID == "" || appConfig.ClientSecret == "" {
		appConfig.ClientID, appConfig.ClientSecret = config.MustReadAppCredentials()
		log.Debugf("Got appConfig from stdin: %+v", appConfig)
		config.MustWriteAppConfig(appConfig)
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
			Name:  "default-substitutions, u",
			Usage: "Same as defining -s \"_, \"",
		},
		cli.BoolTFlag{
			Name: "no-parse-year",
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
