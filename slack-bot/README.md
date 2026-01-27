# csuitebluelight Slack Bot

A Slack slash command that displays deployment status for all regions.

## Usage

In Slack, type:
```
/deploy-status
```

Returns the overall status plus regional statuses.

## Setup

### 1. Create the Lambda Function

1. Go to AWS Lambda console
2. Create function → Author from scratch
3. Function name: `csuite-deploy-status`
4. Runtime: Python 3.11 (or later)
5. Create function
6. Copy contents of `lambda_function.py` into the code editor
7. Deploy

### 2. Create a Function URL

1. In the Lambda function, go to Configuration → Function URL
2. Create function URL
3. Auth type: NONE (Slack handles its own verification)
4. Save and copy the Function URL

### 3. Create the Slack App

1. Go to https://api.slack.com/apps
2. Create New App → From scratch
3. Name: `Deploy Status` (or your preference)
4. Select your workspace

### 4. Configure the Slash Command

1. In the Slack app settings, go to Slash Commands
2. Create New Command:
   - Command: `/deploy-status`
   - Request URL: (paste your Lambda Function URL)
   - Short Description: `Check CSuite deployment status`
   - Usage Hint: (leave blank)
3. Save

### 5. Install the App

1. Go to Install App in the sidebar
2. Install to Workspace
3. Authorize

## Testing

Type `/deploy-status` in any channel or DM. You should see:

```
Overall Status: complete
AU: complete | CA: complete | OR: complete | US: complete
```

## Configuration

To change the status URLs, edit `STATUS_URLS` in `lambda_function.py`.

## Response Visibility

The bot uses `"response_type": "ephemeral"` so only the user who triggered the command sees the response. Change to `"in_channel"` in `format_slack_response()` if you want everyone in the channel to see it.
