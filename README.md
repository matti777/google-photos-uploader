# Google Photos CLI Uploader

A command line utility for uploading a local directory structure of (JPEG) images into Google Photos as new Albums. The contents of each subdirectory (and, optionally, the contents of its children recursively) are uploaded as an Album whose name is derived from the name of the directory.

## Prerequisites

Using the Google Photos API requires the use of the Google OAuth2 authorization flow. Unfortunately
this means you have to set up a GCP project for this purpose.

See the instructions here: [https://developers.google.com/photos/library/guides/authorization](https://developers.google.com/photos/library/guides/authorization).

Make a note of the Client ID / Client Secret values from the GCP console.

## Building the application

To build the binary with debug logs, run:

```sh
make uploader-debug
```

To build the binary without debug logs, run:

```sh
make uploader
```

## Development Guidelines

For any Go code you write to this module, follow these guidelines:

1. Do not write long lines. Keep your code line length to < 100 chars.
2. Use `gofmt` for code formatting; hook it to your text editor's save step so it gets run automatically.
3. Follow any and all best practices for writing clean Go code.

## License

Released under the MIT license. See [LICENSE.md](LICENSE.md).

## Contact

For any questions, concact Matti Dahlbom <mdahlbom666@gmail.com>.
