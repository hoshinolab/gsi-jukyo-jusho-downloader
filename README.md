# gsi-jukyo-jusho-downloader
国土地理院 電子国土基本図（地名情報）「住居表示住所」を一括でダウンロードするCLIツールです。

## Usage

1. Download binary or build this. And deploy the binary to your working directory
2. Make output directory. (NOTE: this program doesn't make the directory automatically.)
3. `$ ./gsi-downloader --outdir ./your_output_directory  --del`
4. wait

## CLI Options

- `--nodownload`: Skip downloading zip files from GSI (default: FALSE)
- `--nounzip`: Skip extracting zip files from GSI (default: FALSE)
- `--del`: Delete temporary files (csv/zip) (default: FALSE)

## Build

1. Clone this repository in your `$GOPATH/src`
2. `$ go build ./cmd/gsi-downloader/main.go -o gsi-downloader`

