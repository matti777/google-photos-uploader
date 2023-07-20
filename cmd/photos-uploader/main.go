package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/matti777/google-photos-uploader/internal/config"
	"github.com/matti777/google-photos-uploader/internal/files"
	photos "github.com/matti777/google-photos-uploader/internal/googlephotos"
	photosutil "github.com/matti777/google-photos-uploader/internal/googlephotos/util"
	"github.com/matti777/google-photos-uploader/internal/logging"

	"github.com/urfave/cli/v2"
)

var (
	log      = logging.MustGetLogger()
	settings = config.MustGetSettings()

	// Application configuration
	appConfig *config.AppConfiguration
)

func readFlags(c *cli.Context) {
	settings.Recurse = c.Bool("recursive")
	log.Debugf("Recurse into subdirectories: %v", settings.Recurse)

	settings.SkipConfirmation = c.Bool("yes")

	settings.DryRun = c.Bool("dry-run")
	if settings.DryRun {
		log.Debugf("--dry-run enabled, not changes will be made")
	}

	settings.NameSubstitutionTokens = c.String("folder-name-substitutions")
	if settings.NameSubstitutionTokens != "" {
		log.Debugf("Using folder name substitution tokens: %v",
			settings.NameSubstitutionTokens)
	}

	settings.NoParseYear = c.Bool("no-parse-year")
	log.Debugf("Skipping parsing folder year?: %v", settings.NoParseYear)

	settings.Capitalize = c.Bool("capitalize")
	log.Debugf("Capitalizing folder name words: %v", settings.Capitalize)

	settings.MaxConcurrency = c.Int("concurrency")
	log.Debugf("maxConcurrency = %v", settings.MaxConcurrency)
}

func defaultAction(c *cli.Context) error {
	log.Debugf("Running Default action..")

	readFlags(c)

	baseDir := c.Args().Get(0)
	if baseDir == "" {
		cli.ShowAppHelp(c)
		return nil
	}

	// Resolve the base dir
	baseDir, err := filepath.Abs(baseDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path for '%v': %v", err)
	}
	log.Debugf("Finding photo album directories under base directory %v", baseDir)

	authorize := c.Bool("authorize")

	appConfig = config.ReadAppConfig()

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
	if appConfig.AuthToken == nil || appConfig.UserInfo.ID == "" {
		fmt.Printf("Authenticating you..\n")
		a := photosutil.NewAuthenticator(appConfig.ClientID, appConfig.ClientSecret)
		token, userInfo, err := a.Authorize()
		if err != nil {
			log.Fatalf("Failed to get authorization token")
		} else {
			fmt.Println("Authorization OK!")
			appConfig.AuthToken = token
			appConfig.UserInfo = *userInfo
			// TODO REMOVE
			log.Debugf("REMOVE ME: UserInfo: %+v", appConfig.UserInfo)
			config.MustWriteAppConfig(appConfig)
		}
		fmt.Printf("Authorized as '%v' (%v) -- specify --authorize to authorize "+
			"on a different account.\n", appConfig.UserInfo.Name,
			appConfig.UserInfo.Email)
	} else if !settings.SkipConfirmation {
		fmt.Printf("You have authenticated as %v (%v). Continue? [y/N]\n",
			appConfig.UserInfo.Name, appConfig.UserInfo.Email)
		reader := bufio.NewReader(os.Stdin)
		res, _ := reader.ReadString('\n')
		if strings.ToLower(strings.Trim(res, " \n\t\r")) != "y" {
			fmt.Print("Re-run with --authorize to re-authorize as a different user.")
			return nil
		}
	}

	photosClient := photos.MustCreateClient(appConfig.ClientID, appConfig.ClientSecret,
		appConfig.AuthToken)

	// Retrieve the list of albums
	fmt.Printf("Fetching the list of Photos albums..\n")
	if l, err := photosClient.ListAlbums(); err != nil {
		log.Fatalf("Failed to list Google Photos albums: %v", err)
	} else {
		settings.Albums = l
	}

	for _, a := range settings.Albums {
		log.Debugf("Found existing Photos Album: '%v'", a.Title)
	}

	fmt.Printf("Will look for images with file extensions: %v\n", settings.ImageExtensions)

	files.ProcessBaseDir(baseDir)

	return nil
}

func main() {
	appname := os.Args[0]
	log.Debugf("main(): running binary %v..", appname)

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
	app.Copyright = "(c) 2018-2023 Matti Dahlbom"
	app.Version = "1.0.0"
	app.Action = defaultAction
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:  "authorize",
			Value: true,
			Usage: "Trigger Google authorization flow. " +
				"You only have to run this one time; " +
				"after you have authenticated, the authentication token " +
				"will be stored. If you want to authenticate with another " +
				"account, simply define this flag again. Specifying this flag also " +
				"causes the client ID / secret to be reset and they can be re-entered.",
		},
		&cli.BoolFlag{
			Name:  "recursive, r",
			Usage: "Process subdirectories of the photo directories recursively",
		},
		&cli.BoolFlag{
			Name:  "yes, y",
			Value: true,
			Usage: "Answer Yes to all confirmations",
		},
		&cli.BoolFlag{
			Name:  "dry-run, n",
			Value: true,
			Usage: "Specify to just scan, not actually upload anything",
		},
		&cli.IntFlag{
			Name:  "concurrency, c",
			Usage: "Maximum number of simultaneous uploads",
			Value: 1,
		},
		&cli.StringFlag{
			Name: "folder-name-substitutions, s",
			Usage: "Directory name -> Photos Folder substition " +
				"tokens, default is no substitution. " +
				"The format is CSV like so: old1,new1,old2,new2 where token " +
				"new1 would replace token old1 etc. For example to replace " +
				"all underscores with spaces and add spaces around " +
				"all dashes, specify -s \"_, ,-, - \"",
		},
		&cli.BoolFlag{
			Name:  "no-parse-year",
			Value: true,
			Usage: "Do not attempt to parse the year from the directory " +
				"name; by default, an attempt is made to extract Photos Folder " +
				"creation date using regex '.*[\\- _][0-9]{4}'. Eg. " +
				"'Pictures_from_Thailand-2009' would generate folder creation " +
				"year 2009.",
		},
		&cli.BoolFlag{
			Name:  "capitalize, a",
			Value: true,
			Usage: "When forming the Photos Folder names, capitalize the " +
				"first letter of each word, ie 'trip to tonga, 2018' " +
				"would become 'Trip To Tonga, 2018'. Combine with " +
				"folder-name-substitutions to clean up the directory names",
		},
	}

	app.Run(os.Args)
}
