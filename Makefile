MAIN_BINARY=bin/photos-uploader
MAIN_SRC=./cmd/photos-uploader/

EXIFTEST_BINARY=bin/exiftest
EXIFTEST_SRC=./cmd/exiftest/

all: protoc compile test

protoc:
	protoc -I=proto --go_out=internal proto/journal.proto

exiftest:
	go build -o $(EXIFTEST_BINARY) $(EXIFTEST_SRC)

uploader:
	go build -o $(MAIN_BINARY) $(MAIN_SRC) 

compile: exiftest uploader

test:
	go test ./...

vet:
	go vet

clean:
	go clean
	rm -f $(EXIFTEST_BINARY)
	rm -f $(MAIN_BINARY)

list:
	@grep '^[^#[:space:]].*:' Makefile
