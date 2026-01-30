# Foundant CSuite Deployment Dashboard

## Project Goal
Create a dashboard to display software deployment status from text files in AWS regions.

## Requirements
- Data source: Plain text files accessible via HTTPS URLs, each containing a single status string
- Regions: AU, CA, OR, US (4 regional files + 1 overall status file)
- Refresh rate: Every 60 seconds
- No alerting/notifications needed - on-demand viewing only

## Status URLs
```
Overall: https://content.fcsuite.com/deploy/deploy
AU:      https://content.fcsuite.com/deploy/deploy-au
CA:      https://content.fcsuite.com/deploy/deploy-ca
OR:      https://content.fcsuite.com/deploy/deploy-or
US:      https://content.fcsuite.com/deploy/deploy-us
```

## Status Color Coding
- **Red**: testfail, error, fetch errors
- **Green**: testok, testing, merging, building, deploy
- **Blue**: pr
- **White/default**: complete, unknown

## Completed

### 1. Standalone webpage (`web/`)
- Displays overall status prominently at top
- Shows 4 regional status cards (AU, CA, OR, US)
- Auto-refreshes every 60 seconds via JavaScript
- Dark theme with clean card-based layout
- Shows last-updated timestamp
- Error handling for failed fetches
- Configured with real status URLs (HTTPS)
- CORS note: serve from `content.fcsuite.com` to avoid cross-origin issues

### 2. Slack bot (`slack-bot/`)
- `lambda_function.py` - AWS Lambda handler for slash command
- Emoji color coding (🟥 red, 🟩 green, 🟦 blue, ⬜ white)
- `test_lambda.py` - pytest tests with mocked HTTP
- `README.md` - Setup instructions for Lambda + Slack app
- Usage: `/deploy-status` slash command
- Returns status visible only to the requesting user (ephemeral)

### 3. CLI tool (`cli/`)
- `main.go` - Cross-platform Go CLI
- Color-coded terminal output using `github.com/fatih/color`
- `--watch` flag for continuous refresh mode with independent read/write loops:
  - Display refreshes every 30 seconds (reads from disk cache)
  - Network fetch uses variable interval: 30s when active, 85s when complete
  - Cache only writes to disk when status values change
  - Shows "last cache read" and "last cache write" timestamps in watch mode (format: `Mon Jan 2 15:04:05 2006`)
- Disk cache stored in OS-appropriate location (`~/.cache/`, `~/Library/Caches/`, `%LocalAppData%`)
- `main_test.go` - Tests using mock RoundTripper (no network calls)
- Binaries distributed via GitHub Releases (not committed to repo)
- `cli/dist/` directory preserved via `.gitkeep` for local builds

## Testing

### Expectations
- All code should have thorough unit tests
- Use mocked HTTP calls for unit tests (no network dependencies)
- Test both success and error cases
- Test edge cases (empty strings, unknown values, etc.)

### Running Tests
```bash
# Go CLI tests
cd cli && go test -v ./...

# Python Slack bot tests
cd slack-bot && source venv/bin/activate && pytest test_lambda.py -v
```

### Verifying Code Coverage
After making changes, verify test coverage:
```bash
# Go coverage summary
cd cli && go test -cover ./...

# Go coverage by function
cd cli && go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

# Python coverage
cd slack-bot && source venv/bin/activate && pytest --cov=lambda_function test_lambda.py
```

## Code Style
- **Go**: Always run `gofmt -w` on all Go files after making changes
- **Git**: Do not add "Co-Authored-By" lines to commit messages

## After Every Change
- **Tests**: Always update or add tests to cover your changes. Run tests to verify they pass.
- **Documentation**: Update relevant documentation (README files, CLAUDE.md) to reflect any user-facing or behavioral changes.

## Releasing

CLI releases are automated via GitHub Actions (`.github/workflows/release.yml`):

1. Go to **Actions** → **Release** in GitHub
2. Click **Run workflow**
3. Enter version number (e.g., `1.0.0`)
4. Workflow will:
   - Run tests (fails release if tests fail)
   - Build binaries for Linux, macOS, Windows (amd64 + arm64)
   - Smoke test binaries on native runners (linux-arm64 excluded, cross-compiled)
   - Create git tag `v1.0.0`
   - Create GitHub Release with binaries attached

## Notes
- CORS: Status endpoints don't have CORS headers; dashboard must be served from same domain
- Go is installed at `/usr/local/go/bin/go`
- Python venv for slack-bot tests: `slack-bot/venv/`
