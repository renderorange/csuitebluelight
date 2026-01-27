# csuitebluelight

Monitor CSuite deployment status across regions.

## About

`csuitebluelight` is an experiment to build and deploy software using only Claude Code.

Designed, built, verified, and tested only using instructions through the prompt.

## Status Color Coding

| Color | Statuses |
|-------|----------|
| Red | testfail, error, fetch errors |
| Green | testok, testing, merging, building, deploy |
| Blue | pr |
| White | complete, unknown |

## Web Dashboard

Browser-based dashboard with auto-refresh. See [web](web/).

## CLI

Command-line tool with optional watch mode. See [cli](cli/).

## Slack Bot

Slash command for checking status from Slack. See [slack-bot](slack-bot/).

## License and Copyright

This software is available under the MIT license.

Copyright (c) 2026 Foundant
