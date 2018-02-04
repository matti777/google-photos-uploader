#!/bin/bash

echo "Regenerating the protobuffer source file(s).."
protoc -I=proto --go_out=src proto/journal.proto
