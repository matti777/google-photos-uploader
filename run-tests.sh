#! /bin/bash

# Append current dir to GOPATH so the src dir will be found by the compiler
MY_GOPATH="$GOPATH:`pwd`"

# Run all the tests
GOPATH=$MY_GOPATH go test -v mdahlbom/google-photos-uploader
#GOPATH=$MY_GOPATH go test -v \
#      mdahlbom/google-photos-uploader/googlephotos/util/
