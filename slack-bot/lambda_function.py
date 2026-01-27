import json
import urllib.request
import urllib.parse
from concurrent.futures import ThreadPoolExecutor

STATUS_URLS = {
    'overall': 'https://content.fcsuite.com/deploy/deploy',
    'au': 'https://content.fcsuite.com/deploy/deploy-au',
    'ca': 'https://content.fcsuite.com/deploy/deploy-ca',
    'or': 'https://content.fcsuite.com/deploy/deploy-or',
    'us': 'https://content.fcsuite.com/deploy/deploy-us',
}

REGION_LABELS = {
    'overall': 'Overall',
    'au': 'AU',
    'ca': 'CA',
    'or': 'OR',
    'us': 'US',
}

# Status emoji mapping
# Red: error states
# Green: in-progress states
# Blue: PR state
# White: complete/default
STATUS_EMOJI = {
    'testfail': '🟥',
    'error': '🟥',
    'testok': '🟩',
    'testing': '🟩',
    'merging': '🟩',
    'building': '🟩',
    'deploy': '🟩',
    'pr': '🟦',
}

DEFAULT_EMOJI = '⬜'
ERROR_EMOJI = '🟥'


def get_status_emoji(status):
    """Get the emoji for a status value."""
    if status is None:
        return ERROR_EMOJI  # red for fetch errors
    return STATUS_EMOJI.get(status.lower(), DEFAULT_EMOJI)


def fetch_status(region):
    """Fetch status for a single region."""
    url = STATUS_URLS[region]
    try:
        with urllib.request.urlopen(url, timeout=10) as response:
            status = response.read().decode('utf-8').strip()
            return region, status, None
    except Exception as e:
        return region, None, str(e)


def fetch_all_statuses():
    """Fetch all statuses concurrently."""
    regions = ['overall', 'au', 'ca', 'or', 'us']
    with ThreadPoolExecutor(max_workers=5) as executor:
        results = list(executor.map(fetch_status, regions))
    return {region: (status, error) for region, status, error in results}


def format_slack_response(statuses):
    """Format statuses as a Slack message with emoji indicators."""
    lines = []

    # Overall status
    overall_status, overall_error = statuses['overall']
    if overall_error:
        emoji = ERROR_EMOJI
        lines.append(f"{emoji} *Status:* {overall_error}")
    else:
        emoji = get_status_emoji(overall_status)
        lines.append(f"{emoji} *Status:* {overall_status}")

    # Regional statuses
    for region in ['au', 'ca', 'or', 'us']:
        status, error = statuses[region]
        label = REGION_LABELS[region]
        if error:
            emoji = ERROR_EMOJI
            lines.append(f"{emoji} *{label}:* {error}")
        else:
            emoji = get_status_emoji(status)
            lines.append(f"{emoji} *{label}:* {status}")

    return {
        "response_type": "ephemeral",
        "blocks": [
            {
                "type": "section",
                "text": {
                    "type": "mrkdwn",
                    "text": "\n".join(lines)
                }
            }
        ]
    }


def lambda_handler(event, context):
    """AWS Lambda handler for Slack slash command."""
    # Parse the Slack request body
    if 'body' in event:
        body = event['body']
        if event.get('isBase64Encoded'):
            import base64
            body = base64.b64decode(body).decode('utf-8')
        params = urllib.parse.parse_qs(body)
    else:
        params = {}

    # Fetch all statuses
    statuses = fetch_all_statuses()

    # Format and return response
    slack_response = format_slack_response(statuses)

    return {
        'statusCode': 200,
        'headers': {
            'Content-Type': 'application/json'
        },
        'body': json.dumps(slack_response)
    }
