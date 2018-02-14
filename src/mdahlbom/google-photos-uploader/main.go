package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	photos "mdahlbom/google-photos-uploader/google_photos"
	"mdahlbom/google-photos-uploader/google_photos/util"

	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
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

	// Whether to skip (assume Yes) all confirmations)
	skipConfirmation = false

	// Whether doing a 'dry run', ie not actually sending anything.
	dryRun = false

	// Google Photos API client
	photosClient *photos.Client
)

// Simulates the upload of a file.
func simulateUpload(progressBar *uiprogress.Bar, dir, dirName string,
	file os.FileInfo) error {

	const steps = 5

	remaining := file.Size()
	sent := int64(0)
	perStep := remaining / steps

	for i := 0; i < steps; i++ {
		time.Sleep(time.Millisecond * 250)

		if remaining < perStep {
			sent += remaining
		} else {
			sent += perStep
		}

		progressBar.Set(int(sent))

		remaining -= perStep
	}

	return nil
}

func upload(dir, dirName string, file os.FileInfo,
	padLength int) error {

	log.Debugf("Uploading '%v' for dirName '%v'", file.Name(), dirName)

	paddedName := strutil.PadRight(file.Name(), padLength, ' ')

	progress := uiprogress.New()
	bar := progress.AddBar(int(file.Size())).PrependElapsed().AppendCompleted()
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return paddedName
	})
	bar.Fill = '#'
	bar.Head = '#'
	bar.Empty = ' '

	progress.Start()
	defer progress.Stop()

	return simulateUpload(bar, dir, dirName, file)
}

func defaultAction(c *cli.Context) error {
	log.Debugf("Running Default action..")

	authorize := GlobalBoolT(c, "authorize")

	// Make sure we have an auth token, ie. the user has performed the
	// authorization flow.
	if appConfig.AuthToken == nil || appConfig.UserInfo == nil && !authorize {
		fmt.Printf("Not authorized; you must perform the authorization " +
			"flow. Run again and specify the --authorize flag.")
		return fmt.Errorf("Missing authorization.")
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

	log.Debugf("baseDir: %v", baseDir)

	// Resolve the base dir
	baseDir, err := filepath.Abs(baseDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path for '%v': %v", err)
	}
	log.Debugf("Cleaned baseDir: %v", baseDir)

	// Check if need to authenticate the user
	if authorize {
		log.Debugf("Running authorization flow..")
		a := util.NewAuthenticator(appConfig.ClientID,
			appConfig.ClientSecret)
		token, userInfo, err := a.Authorize()
		if err != nil {
			log.Fatalf("Failed to get authorization token")
		} else {
			log.Debugf("Got oauth2 token: %v", token)
			fmt.Println("Authorization OK!")
			appConfig.AuthToken = token
			appConfig.UserInfo = userInfo
			mustWriteAppConfig(appConfig)
		}
	}

	fmt.Printf("Authorized as '%v' -- specify --authorize to authorize "+
		"on a different account.\n", appConfig.UserInfo.Name)

	// Create the API client
	photosClient = photos.NewClient(appConfig.ClientID, appConfig.ClientSecret,
		appConfig.AuthToken)

	disregardJournal := GlobalBoolT(c, "disregard-journal")
	if disregardJournal {
		log.Debugf("Disregarding reading journal files..")
	}

	recursive := GlobalBoolT(c, "recursive")
	log.Debugf("Recurse into subdirectories: %v", recursive)

	skipConfirmation = GlobalBoolT(c, "yes")
	dryRun = GlobalBoolT(c, "dry-run")

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

	capitalize = GlobalBoolT(c, "capitalize")
	log.Debugf("Capitalizing folder name words: %v", capitalize)

	mustProcessDir(baseDir, recursive, disregardJournal)

	return nil
}

func main() {
	// Set up logging
	setupLogging()

	appname := os.Args[0]
	log.Debugf("main(): running %v..", appname)

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
			Name:  "dry-run",
			Usage: "Specify to just scan, not actually upload anything",
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
			Name: "capitalize, c",
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

	//app.Commands = commands
	app.Run(os.Args)
}
