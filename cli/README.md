# csuitebluelight CLI

Command-line tool to check CSuite deployment status.

## Installation

Download the latest binary for your platform from [Releases](../../../releases).

Or build from source:
```
cd cli
go build -o deploy-status
```

## Usage

```
deploy-status              # Check once
deploy-status --watch      # Continuous monitoring
```

## Example Output

Single check:
```
CSuite Deploy Status

Status     complete
AU         complete
CA         complete
OR         complete
US         complete
```

Watch mode (with timestamps):
```
CSuite Deploy Status

Status     testing
AU         complete
CA         complete
OR         complete
US         complete

Cache read:    3:04:35 PM
Cache written: 3:04:05 PM
Press Ctrl+C to exit
```

## Watch Mode

In watch mode (`--watch`), the CLI uses independent read and write loops:

- **Display refresh**: Every 30 seconds, reloads from disk cache and updates the screen
- **Network fetch**: Variable interval based on status:
  - 30 seconds when deployment is active (status != "complete")
  - 85 seconds when deployment is complete (reduces unnecessary polling)
- **Cache writes**: Only writes to disk when status values actually change

The "Cache read" timestamp shows when the display last refreshed from disk.
The "Cache written" timestamp shows when new data was fetched from the network.

## Caching

Status data is cached to disk at:
- **Linux**: `~/.cache/csuitebluelight/statuses.json`
- **macOS**: `~/Library/Caches/csuitebluelight/statuses.json`
- **Windows**: `%LocalAppData%\csuitebluelight\statuses.json`

## Creating a Release

Releases are created via GitHub Actions:

1. Go to **Actions** → **Release** in the GitHub repo
2. Click **Run workflow**
3. Enter a version number (e.g., `1.0.0`)
4. Click **Run workflow**

The workflow runs tests, builds binaries for all platforms, smoke tests each binary on native runners (except linux-arm64 which is cross-compiled), creates a `v1.0.0` tag, and publishes a GitHub Release with the binaries attached.

## Building Locally

For local development, build for your platform:
```
go build -o deploy-status
```

Cross-compile for all platforms:
```
GOOS=linux GOARCH=amd64 go build -o dist/deploy-status-linux-amd64
GOOS=linux GOARCH=arm64 go build -o dist/deploy-status-linux-arm64
GOOS=darwin GOARCH=amd64 go build -o dist/deploy-status-darwin-amd64
GOOS=darwin GOARCH=arm64 go build -o dist/deploy-status-darwin-arm64
GOOS=windows GOARCH=amd64 go build -o dist/deploy-status-windows-amd64.exe
```
