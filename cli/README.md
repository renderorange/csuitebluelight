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
deploy-status --watch      # Continuous refresh (30s active, 85s idle)
```

## Example Output

```
CSuite Deploy Status

Status     complete
AU         complete
CA         complete
OR         complete
US         complete
```

## Creating a Release

Releases are created via GitHub Actions:

1. Go to **Actions** → **Release** in the GitHub repo
2. Click **Run workflow**
3. Enter a version number (e.g., `1.0.0`)
4. Click **Run workflow**

The workflow runs tests, builds binaries for all platforms, creates a `v1.0.0` tag, and publishes a GitHub Release with the binaries attached.

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
