# Google Photos CLI Uploader

A command line utility for uploading a local directory structure of (JPEG) images into Google Photos as new Albums. Each directory is uploaded into an Album whose name is derived from the name of the directory.

## Prerequisites

You must have `protoc` and `protoc-gen-go` installed, eg.:

```sh
brew install protobuf
brew install protoc-gen-go
```

## Building the application

To build the proto files, run:

```sh
make protoc
```

To build the binary, run:

```sh
make compile
```

## Omitting Debug logging

To build a binary without debug logs, specify `-tags nodebug` at the `go build` command you are running to build the binary.

## Development Guidelines

For any Go code you write to this module, follow these guidelines:

1. Do not write long lines. Keep your code line length to < 100 chars.
2. Use `gofmt` for code formatting; hook it to your text editor's save step so it gets run automatically.
3. Follow any and all best practices for writing clean Go code.

## License

Released under the MIT license. See [LICENSE.md](LICENSE.md).

## Contact

For any questions, concact Matti Dahlbom <mdahlbom666@gmail.com>.
