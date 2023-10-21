package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/matti777/google-photos-uploader/internal/config"
	"github.com/matti777/google-photos-uploader/internal/exiftool"
	"github.com/matti777/google-photos-uploader/internal/files"
	photos "github.com/matti777/google-photos-uploader/internal/googlephotos"
	photosutil "github.com/matti777/google-photos-uploader/internal/googlephotos/util"
	"github.com/matti777/google-photos-uploader/internal/logging"
	"github.com/matti777/google-photos-uploader/internal/util"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var (
	log      *logrus.Logger
	settings = config.MustGetSettings()

	// Application configuration
	appConfig *config.AppConfiguration
)

func readFlags(c *cli.Context) {
	settings.Recurse = c.IsSet("recursive")
	log.Debugf("Recurse into subdirectories: %v", settings.Recurse)

	settings.SkipConfirmation = c.IsSet("yes")
	if settings.SkipConfirmation {
		log.Debugf("--yes defined, will skip all confirmations")
	}

	settings.DryRun = c.IsSet("dry-run")
	if settings.DryRun {
		log.Debugf("--dry-run enabled, not changes will be made")
	}

	settings.NameSubstitutionTokens = c.String("folder-name-substitutions")
	if settings.NameSubstitutionTokens != "" {
		log.Debugf("Using folder name substitution tokens: %v",
			settings.NameSubstitutionTokens)
	}

	settings.NoParseYear = c.IsSet("no-parse-year")
	log.Debugf("Skipping parsing folder year?: %v", settings.NoParseYear)

	settings.Capitalize = c.Bool("capitalize")
	log.Debugf("Capitalizing folder name words: %v", settings.Capitalize)

	settings.MaxConcurrency = c.Int("concurrency")
	log.Debugf("maxConcurrency = %v", settings.MaxConcurrency)
}

func handleAuthorize(c *cli.Context) error {
	appConfig = config.ReadAppConfig()

	authorize := c.IsSet("authorize")

	if authorize {
		appConfig.ClientID = ""
		appConfig.ClientSecret = ""
		appConfig.AuthToken = nil
		appConfig.UserInfo = photosutil.UserInfo{}
		config.MustWriteAppConfig(appConfig)
		log.Debugf("Re-authentication requested; all authentication data has been reset.")
	}

	if appConfig.ClientID == "" || appConfig.ClientSecret == "" {
		appConfig.ClientID, appConfig.ClientSecret = config.MustReadAppCredentials()
		config.MustWriteAppConfig(appConfig)
	}

	// Check if need to authenticate the user
	if appConfig.AuthToken == nil {
		fmt.Printf("Authenticating you..\n")
		a := photosutil.NewAuthenticator(appConfig.ClientID, appConfig.ClientSecret)
		token, userInfo, err := a.Authorize()
		if err != nil {
			log.Fatalf("Failed to get authorization token")
		} else {
			fmt.Println("Authorization OK!")
			appConfig.AuthToken = token
			appConfig.UserInfo = *userInfo

			// TODO fetch further user info to get email address etc

			config.MustWriteAppConfig(appConfig)
		}
		fmt.Printf("Authorized as '%v' (%v) -- specify --authorize to authorize "+
			"on a different account.\n", appConfig.UserInfo.Name,
			appConfig.UserInfo.Email)
	} else {
		util.MustConfirm(fmt.Sprintf("You have authenticated as %v (%v).",
			appConfig.UserInfo.Name, appConfig.UserInfo.Email),
			"Re-run with --authorize to re-authorize as a different user.")
	}

	return nil
}

func mustInitGooglePhotos() {
	if appConfig.ClientID == "" || appConfig.ClientSecret == "" || appConfig.AuthToken == nil {
		log.Fatalf("appConfig missing credentials to create Photos client")
	}

	photosClient := photos.MustCreateClient(appConfig.ClientID, appConfig.ClientSecret,
		appConfig.AuthToken)

	// Retrieve the list of albums and store into settings
	fmt.Printf("Fetching the list of existing Google Photos albums..\n")
	if l, err := photosClient.ListAlbums(); err != nil {
		log.Fatalf("Failed to list Google Photos albums: %v", err)
	} else {
		settings.Albums = l
	}

	for _, a := range settings.Albums {
		log.Debugf("Found existing Photos Album: '%v'", a.Title)
	}
}

func defaultAction(c *cli.Context) error {
	logLevel := logrus.ErrorLevel
	if c.IsSet("verbose") {
		logLevel = logrus.DebugLevel
	}
	log = logging.MustGetLogger()
	log.SetLevel(logLevel)

	readFlags(c)

	exiftool.MustCheckExiftoolInstalled()

	baseDir := c.Args().Get(0)
	if baseDir == "" {
		cli.ShowAppHelp(c)
		return cli.Exit("Must define a base directory!", -1)
	}

	// Resolve the base dir
	baseDir, err := filepath.Abs(baseDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path for '%v': %v", baseDir, err)
	}
	log.Debugf("Base directory is: %v", baseDir)

	if !settings.DryRun {
		if err := handleAuthorize(c); err != nil {
			return nil
		}

		mustInitGooglePhotos()
	}

	files.ProcessBaseDir(baseDir)

	return nil
}

func main() {
	appname := os.Args[0]

	// Setup CLI app framework
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Name = appname
	app.ArgsUsage = "[directory]"
	app.Usage = "A command line Google Photos upload utility"
	app.UsageText = fmt.Sprintf("%v [options] directory", appname)
	app.Description = fmt.Sprintf("Command-line utility for uploading "+
		"photos to Google Photos from a local disk directory.\n\n"+
		"You must supply a directory as argument; the contents of the subdirectories of that"+
		"directory will be uploaded as albums to Google Photos.\n\n"+
		"Currently only JPEG images are supported.\n\n"+
		"For help, run '%v help'", appname)
	app.Copyright = "(c) 2018-2023 Matti Dahlbom"
	app.Version = "1.0.0"
	app.Action = defaultAction
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:    "authorize",
			Aliases: []string{"a"},
			Value:   false,
			Usage: "Trigger Google authorization flow. " +
				"You only have to run this one time; " +
				"after you have authenticated, the authentication token " +
				"will be stored. If you want to authenticate with another " +
				"account, simply define this flag again. Specifying this flag also " +
				"causes the client ID / secret to be reset and they can be re-entered.",
		},
		&cli.BoolFlag{
			Name:    "recursive",
			Aliases: []string{"r"},
			Value:   false,
			Usage:   "Process subdirectories of the photo directories recursively",
		},
		&cli.BoolFlag{
			Name:    "yes",
			Aliases: []string{"y"},
			Value:   false,
			Usage:   "Answer Yes to all confirmations",
		},
		&cli.BoolFlag{
			Name:    "dry-run",
			Aliases: []string{"n"},
			Value:   false,
			Usage:   "Specify to just scan, not actually upload anything",
		},
		&cli.IntFlag{
			Name:    "concurrency",
			Aliases: []string{"c"},
			Usage:   "Maximum number of simultaneous uploads",
			Value:   1,
		},
		&cli.StringFlag{
			Name:    "folder-name-substitutions",
			Aliases: []string{"s"},
			Usage: "Directory name -> Photos Folder substition " +
				"tokens, default is no substitution. " +
				"The format is CSV like so: old1,new1,old2,new2 where token " +
				"new1 would replace token old1 etc. For example to replace " +
				"all underscores with spaces and add spaces around " +
				"all dashes, specify -s \"_, ,-, - \"",
		},
		&cli.BoolFlag{
			Name:  "no-parse-year",
			Value: false,
			Usage: "Do not attempt to parse the year from the directory " +
				"name; by default, an attempt is made to extract Photos Folder " +
				"creation date using regex '.*[\\- _][0-9]{4}'. Eg. " +
				"'Pictures_from_Thailand-2009' would generate folder creation " +
				"year 2009.",
		},
		&cli.BoolFlag{
			Name:  "capitalize",
			Value: true,
			Usage: "When forming the Photos Folder names, capitalize the " +
				"first letter of each word, ie 'trip to tonga, 2018' " +
				"would become 'Trip To Tonga, 2018'. Combine with " +
				"folder-name-substitutions to clean up the directory names",
		},
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"vv"},
			Value:   false,
			Usage:   "Specify to enable debug logging",
		},
	}

	app.Run(os.Args)
}
