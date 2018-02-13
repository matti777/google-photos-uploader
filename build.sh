#!/bin/bash

BUILD_FLAGS=""

if [ "X$1" == "Xhelp" ];
then
    echo "Usage: $0 [nodebug]"
    echo "Specify 'nodebug' argument to disable debug logging."
fi

if [ "X$1" == "Xnodebug" ];
then
    echo "Disabling debug logging."
    BUILD_FLAGS="-tags nodebug"
fi

MY_GOPATH="`pwd`:$GOPATH"
GOPATH=$MY_GOPATH go build $BUILD_FLAGS -o photos-uploader \
      mdahlbom/google-photos-uploader
