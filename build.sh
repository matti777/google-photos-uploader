#!/bin/bash

MY_GOPATH="`pwd`:$GOPATH"
GOPATH=$MY_GOPATH go build -o photos-uploader mdahlbom/google-photos-uploader
