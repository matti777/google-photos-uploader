# Google Photos CLI Uploader

A command line utility for uploading a local directory structure of (JPEG) images into Google Photos as new Albums. The contents of each subdirectory (and, optionally, the contents of its children recursively) are uploaded as an Album whose name is derived from the name of the directory.

## Prerequisites

Using the Google Photos API requires the use of the Google OAuth2 authorization flow. Unfortunately
this means you have to set up a GCP project for this purpose.

See the instructions here: [https://developers.google.com/photos/library/guides/get-started](https://developers.google.com/photos/library/guides/get-started).

Make a note of the Client ID / Client Secret values from the GCP console.

### Exiftool

This project uses [https://exiftool.org/](Exiftool) to write EXIF data into the JPEG files.

To install the tool:

- MacOS: `brew install exiftool`
- Debian: `sudo apt-get install exiftool`
- Windows: See https://exiftool.org/install.html

## Building the application

To build the binary (into bin/), run:

```sh
make uploader
```

## License

Released under the MIT license. See [LICENSE.md](LICENSE.md).

## Contact

For any questions, concact Matti Dahlbom <mdahlbom666@gmail.com>.
