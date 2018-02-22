// Contains methods for dealing with directory journals
package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"mdahlbom/google-photos-uploader/pb"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
)

const (
	journalFileName = ".photos-uploader.journal"
)

// Writes a directory's journal. Panics on failure.
func mustWriteJournalFile(dir string, journal *pb.Journal) {
	fileName := filepath.Join(dir, journalFileName)
	log.Debugf("Writing Journal file: %v", fileName)

	out, err := proto.Marshal(journal)
	if err != nil {
		log.Fatalf("Failed to encode journal file: %v", err)
	}

	if err := ioutil.WriteFile(fileName, out, 0644); err != nil {
		log.Fatalf("Failed to write journal file: %v", err)
	}
}

// Reads a directory's journal
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

// Adds a journal entry and persists the directory's journal entry to disk.
// Panics on failure.
func mustAddJournalEntry(dir string, name string, /*isDir bool,*/
	journal *pb.Journal, journalMap *map[string]*pb.JournalEntry) {

	// Make sure there isnt already such an entry (sanity check)
	for _, e := range journal.Entries {
		if e.Name == name {
			log.Fatalf("Already found journal entry '%v' in journal", name)
		}
	}
	if (*journalMap)[name] != nil {
		log.Fatalf("Already found journal map entry '%v' in journal", name)
	}

	entry := &pb.JournalEntry{Name: name, /*IsDirectory: isDir,*/
		Completed: ptypes.TimestampNow()}

	journal.Entries = append(journal.Entries, entry)
	(*journalMap)[name] = entry

	// Save the journal
	mustWriteJournalFile(dir, journal)
	log.Debugf("Added journal entry: %+v", entry)
}

// Creates a map of the journal's file name entries for faster access
func newJournalMap(journal *pb.Journal) map[string]*pb.JournalEntry {
	m := map[string]*pb.JournalEntry{}

	if journal.Entries == nil {
		return m
	}

	for _, e := range journal.Entries {
		m[e.Name] = e
	}

	return m
}
