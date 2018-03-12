# Google Photos CLI Uploader

A command line utility for uploading a local directory structure into Google Photos as new Albums. Each directory is uploaded into an Album whose name is derived from the name of the directory.

Halfway into the work it turned out that Google in their infinite wisdom have, however, deprecated the Album creation API. Automatic Album creation being very crucial for this tool to be very useful, this project was completed mostly as an academic exercise.

## Installing dependencies

```sh
go get -u github.com/op/go-logging
go get -u github.com/urfave/cli
go get -u github.com/golang/protobuf
go get -u github.com/gosuri/uiprogress
go get -u golang.org/x/oauth2
go get -u github.com/google/uuid
go get -u github.com/gorilla/mux
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

## License

Released under the MIT license. See [LICENSE.md](LICENSE.md).

## Acknowledgments

Parts of this work are based on [Photobak](https://github.com/mholt/photobak) by Matt Holt.

## Contact

For any questions, concact Matti Dahlbom <matti@777-team.org>.
