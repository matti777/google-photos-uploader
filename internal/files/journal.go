package files

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/matti777/google-photos-uploader/internal/pb"

	"google.golang.org/protobuf/proto"
)

const (
	journalFileName = ".photos-uploader.journal"
)

// Our read-write locks for accessing the journals
var journalLocks = map[*pb.Journal]*sync.RWMutex{}

// Returns the read-write lock for the journal, or panics if not found
func mustGetJournalLock(journal *pb.Journal) *sync.RWMutex {
	l, ok := journalLocks[journal]
	if !ok {
		log.Fatalf("Failed to get lock for journal")
		return nil
	}

	return l
}

// Writes a directory's journal. Panics on failure. Access must be synchronized
// from the outside; this method does not touch the locks.
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

	journalLocks[journal] = new(sync.RWMutex)

	return journal, nil
}

// Creates an empty journal (and a lock for it)
func newEmptyJournal() *pb.Journal {
	journal := &pb.Journal{}
	journalLocks[journal] = new(sync.RWMutex)

	return journal
}

// Adds a journal entry and persists the directory's journal entry to disk.
// Panics on failure.
func mustAddJournalEntry(dir, name, uploadToken string,
	journal *pb.Journal, journalMap *map[string]*pb.JournalEntry) {

	lock := mustGetJournalLock(journal)
	lock.Lock()
	defer lock.Unlock()

	log.Debugf("Adding journal entry for '%v'", name)

	// Make sure there isnt already such an entry (sanity check)
	for _, e := range journal.Entries {
		if e.Name == name {
			log.Fatalf("Already found journal entry '%v' in journal", name)
		}
	}
	if (*journalMap)[name] != nil {
		log.Fatalf("Already found journal map entry '%v' in journal map", name)
	}

	entry := &pb.JournalEntry{Name: name, UploadToken: uploadToken}

	journal.Entries = append(journal.Entries, entry)
	(*journalMap)[name] = entry

	// Save the journal
	mustWriteJournalFile(dir, journal)
}

// Creates a map of the journal's file name entries for faster access
func newJournalMap(journal *pb.Journal) map[string]*pb.JournalEntry {
	lock := mustGetJournalLock(journal)
	lock.RLock()
	defer lock.RUnlock()

	m := map[string]*pb.JournalEntry{}

	if journal.Entries == nil {
		return m
	}

	for _, e := range journal.Entries {
		m[e.Name] = e
	}

	return m
}
