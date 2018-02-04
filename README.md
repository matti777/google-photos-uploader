# Google Photos CLI Uploader

A command line utility for uploading a local directory structure into Google Photos as new Albums. Each directory is uploaded as a new Album whose name is derived from the name of the directory.

## WORK IN PROGRESS.

This is project just being started.

## Install dependencies:

```sh
go get -u github.com/op/go-logging
go get -u github.com/urfave/cli
go get -u github.com/golang/protobuf
```

## Building the application

To build the binary, run:

```sh
./build.sh
```

To rebuild the proto files, run:

```sh
go get -u github.com/golang/protobuf/protoc-gen-go
./run-protoc.sh
```

## Omitting Debug logging

To build a binary without debug logs, specify `-tags nodebug` at the `go build` command you are running to build the binary.

To pass it to the build script, just run:

```sh
./build.sh nodebug
```

## Development Guidelines

For any Go code you write to this module, follow these guidelines:

1. Do not write long lines. Keep your code line length to < 80 chars.
2. Use `gofmt` for code formatting; hook it to your text editor's save step so it gets run automatically.
3. Follow any and all best practices for writing clean Go code.

## Contact

For any questions, concact Matti Dahlbom <matti@777-team.org.fi>.