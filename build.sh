#!/bin/bash

# Stop script execution at first error
set -e

if [ "X$1" == "Xhelp" ];
then
    echo "Usage: $0 [nodebug]"
    echo "Specify 'nodebug' argument to disable debug logging."
fi

if [ "X$1" == "Xnodebug" ];
then
    echo "Disabling debug logging."
    build_flags="-tags nodebug"
fi

my_gopath="`pwd`:$GOPATH"
GOPATH=$my_gopath go build $build_flags -o photos-uploader \
      mdahlbom/google-photos-uploader
