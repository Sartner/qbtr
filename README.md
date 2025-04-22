# qbtr

qbtr is a CLI tool that transfers completed torrents from qBittorrent to Transmission.

## Features

- Connects to qBittorrent and Transmission via their APIs
- Transfers completed torrents only
- Preserves the downloaded files (doesn't delete them from disk)
- Automatically uses qBittorrent's save path for Transmission
- Configurable via command-line arguments

## Installation


## Cross-Platform Building

This project includes scripts for cross-platform compilation using Docker:

```bash
# Basic build script (builds for multiple platforms)
./build/build.sh
```

The build scripts will create binaries for the following platforms:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64, arm64)

All binaries will be placed in the `target` directory along with checksums.

## Usage

```bash
# usage
qbtr \
  --qb-url=http://localhost:8080 \
  --qb-username=admin \
  --qb-password=adminadmin \
  --qb-auto-delete=true \
  --tr-url=http://localhost:9091/transmission/rpc \
  --tr-username=admin \
  --tr-password=adminadmin \
  --tr-auto-start=true \
  --qb-torrents-dir=/path/to/torrent/files \
  --dry-run=false
```

### Options

| Flag | Description | Default                                |
|------|-------------|----------------------------------------|
| `--qb-url` | qBittorrent WebUI URL |               |
| `--qb-username` | qBittorrent username |                                |
| `--qb-password` | qBittorrent password |                                 |
| `--qb-torrents-dir` | Directory containing torrent files |                         |
| `--qb-auto-delete` | Automatically delete torrents from qBittorrent after transfer | false                                  |
| `--tr-url` | Transmission RPC URL | |
| `--tr-username` | Transmission username |                                |
| `--tr-password` | Transmission password |                                |
| `--tr-auto-start` | Automatically start torrents in Transmission | false                                  |
| `--dry-run` | Test mode that won't delete torrents from qBittorrent, but will add and then delete from Transmission | false                                  |

## Requirements

- qBittorrent with WebUI enabled
- Transmission with RPC enabled
- Access to the .torrent files for completed downloads

## License

MIT 