# csuitebluelight CLI

Command-line tool to check CSuite deployment status.

## Installation

Download the latest binary for your platform from [Releases](../../releases).

Or build from source:
```
cd cli
go build -o deploy-status
```

## Usage

```
deploy-status              # Check once
deploy-status --watch      # Continuous refresh (clears screen)
deploy-status --interval 30  # Refresh every 30 seconds (default: 60)
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

## Building for Multiple Platforms

```
GOOS=linux GOARCH=amd64 go build -o dist/deploy-status-linux-amd64
GOOS=linux GOARCH=arm64 go build -o dist/deploy-status-linux-arm64
GOOS=darwin GOARCH=amd64 go build -o dist/deploy-status-darwin-amd64
GOOS=darwin GOARCH=arm64 go build -o dist/deploy-status-darwin-arm64
GOOS=windows GOARCH=amd64 go build -o dist/deploy-status-windows-amd64.exe
```
