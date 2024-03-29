package util

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/matti777/google-photos-uploader/internal/config"
	"github.com/matti777/google-photos-uploader/internal/logging"
)

var (
	log      = logging.MustGetLogger()
	settings = config.MustGetSettings()

	albumYearRegex *regexp.Regexp
)

var (
	ErrCouldNotParseYear = errors.Errorf("failed to parse the album year")
)

// Asks the user interactively a confirmation question; if the user declines
// (answers anything but Y or default - empty string - stop execution.
func MustConfirm(prompt, declineText string) {
	if settings.SkipConfirmation {
		return
	}

	fmt.Printf("%v\nContinue? [Y/n] ", prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')

	if input != "Y\n" && input != "\n" {
		if declineText != "" {
			fmt.Print(declineText)
		}
		os.Exit(1)
	}
}

// Finds the longest file name
func FindLongestName(infos []os.FileInfo) int {
	longest := 0

	for _, info := range infos {
		if len(info.Name()) > longest {
			longest = len(info.Name())
		}
	}

	return longest
}

// Replaces substrings in the string with other strings, using strings.Replacer.
// The tokens parameter should be formatted as a valid CSV.
func ReplaceInString(s, tokens string) (string, error) {
	if tokens == "" {
		return s, nil
	}

	r := csv.NewReader(strings.NewReader(tokens))
	records, err := r.ReadAll()
	if err != nil {
		return "", errors.Wrap(err, "failed to read CSV")
	}

	if len(records) != 1 {
		return "", errors.Errorf("Invalid number of CSV records. " +
			"Only single line CSV supported.")
	}

	tokenArray := records[0]
	replacer := strings.NewReplacer(tokenArray...)

	return replacer.Replace(s), nil
}

// Chunked returns an array of arrays so that the original array is divided
// into chunks of equal size (except for the remainder chunk).
func Chunked(arr []string, chunkSize int) [][]string {
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

// ParseAlbumYear tries to parse the year of the album from a string
// (directory name) using a certain regex pattern (eg. 'Trip to X - 2009')
// Returns ErrCouldNotParseYear if not found.
func ParseAlbumYear(dirName string) (int, error) {
	res := albumYearRegex.FindStringSubmatch(dirName)

	if len(res) != 2 {
		return 0, ErrCouldNotParseYear
	}

	if res[1] == "" {
		return 0, ErrCouldNotParseYear
	}

	year, err := strconv.Atoi(res[1])
	if err != nil {
		log.Fatalf("failed to parse year string %v to int: %v", res[1], err)
		return 0, ErrCouldNotParseYear
	}

	return year, nil
}

func init() {
	albumYearRegex = regexp.MustCompile(`^.+[-_ ]([12]\d{3})$`)
}
