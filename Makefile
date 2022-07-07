MAIN_BINARY=bin/photos-uploader
MAIN_SRC=./cmd/photos-uploader/

EXIFTEST_BINARY=bin/exiftest
EXIFTEST_SRC=./cmd/exiftest/

protoc:
	protoc -I=proto --go_out=internal proto/journal.proto

compile:
	go build -o $(MAIN_BINARY) $(MAIN_SRC) 

foo:
#  go build -o $(MAIN_BINARY) $(MAIN_SRC) 
#	echo "Building application binaries"
#	go build -o $(EXIFTEST_BINARY) $(EXIFTEST_SRC)

# test:
#   go test ./...

# vet:
#  go vet

# clean:
#    go clean
#    rm $(EXIFTEST_BINARY)
#    rm $(MAIN_BINARY)
