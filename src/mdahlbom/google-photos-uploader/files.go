// File operations

package main

import (
	//"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"mdahlbom/google-photos-uploader/pb"

	"github.com/golang/protobuf/proto"
)

const (
	journalFileName = ".photos-uploader.journal"
)

func directoryExists(dir string) (bool, error) {
	log.Debugf("Checking if %v exists..", dir)

	if info, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			log.Fatalf("Failed to stat '%v': %v", dir, err)
			return false, err
		}
	} else {
		return info.IsDir(), nil
	}
}

func writeJournalFile(dir string, journal *pb.Journal) error {
	fileName := filepath.Join(dir, journalFileName)
	log.Debugf("Writing Journal file: %v", fileName)

	out, err := proto.Marshal(journal)
	if err != nil {
		log.Fatalf("Failed to encode journal file: %v", err)
	}

	if err := ioutil.WriteFile(fileName, out, 0644); err != nil {
		log.Fatalf("Failed to write journal file: %v", err)
	}

	return nil
}

func readJournalFile(dir string) (*pb.Journal, error) {
	fileName := filepath.Join(dir, journalFileName)
	log.Debugf("Reading Journal file: %v", fileName)

	in, err := ioutil.ReadFile(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			log.Debugf("Journal file '%v' does not exist", fileName)
			return nil, nil
		}

		log.Errorf("Error reading journal file: %v", err)
		return nil, err
	}

	journal := &pb.Journal{}

	if err := proto.Unmarshal(in, journal); err != nil {
		log.Errorf("Failed to parse journal proto file: %v", err)
		return nil, err
	}

	return journal, nil
}
