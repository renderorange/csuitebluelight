"""Tests for the Lambda function."""

import json
from unittest.mock import patch, MagicMock
import pytest

from lambda_function import (
    get_status_emoji,
    fetch_status,
    fetch_all_statuses,
    format_slack_response,
    lambda_handler,
)


class TestGetStatusEmoji:
    """Tests for emoji mapping."""

    def test_error_states_return_red(self):
        assert get_status_emoji('testfail') == '🟥'
        assert get_status_emoji('error') == '🟥'
        assert get_status_emoji('TESTFAIL') == '🟥'  # case insensitive

    def test_progress_states_return_green(self):
        assert get_status_emoji('testok') == '🟩'
        assert get_status_emoji('testing') == '🟩'
        assert get_status_emoji('merging') == '🟩'
        assert get_status_emoji('building') == '🟩'
        assert get_status_emoji('deploy') == '🟩'

    def test_pr_returns_blue(self):
        assert get_status_emoji('pr') == '🟦'

    def test_complete_returns_white(self):
        assert get_status_emoji('complete') == '⬜'

    def test_unknown_status_returns_white(self):
        assert get_status_emoji('unknown') == '⬜'
        assert get_status_emoji('random') == '⬜'

    def test_fetch_error_returns_red(self):
        assert get_status_emoji(None) == '🟥'


class TestFetchStatus:
    """Tests for fetching individual status."""

    @patch('lambda_function.urllib.request.urlopen')
    def test_successful_fetch(self, mock_urlopen):
        mock_response = MagicMock()
        mock_response.read.return_value = b'complete'
        mock_response.__enter__ = MagicMock(return_value=mock_response)
        mock_response.__exit__ = MagicMock(return_value=False)
        mock_urlopen.return_value = mock_response

        region, status, error = fetch_status('overall')

        assert region == 'overall'
        assert status == 'complete'
        assert error is None

    @patch('lambda_function.urllib.request.urlopen')
    def test_fetch_error(self, mock_urlopen):
        mock_urlopen.side_effect = Exception('Connection timeout')

        region, status, error = fetch_status('au')

        assert region == 'au'
        assert status is None
        assert error == 'Connection timeout'


class TestFormatSlackResponse:
    """Tests for Slack message formatting."""

    def test_all_complete(self):
        statuses = {
            'overall': ('complete', None),
            'au': ('complete', None),
            'ca': ('complete', None),
            'or': ('complete', None),
            'us': ('complete', None),
        }

        response = format_slack_response(statuses)

        assert response['response_type'] == 'ephemeral'
        text = response['blocks'][0]['text']['text']
        assert '⬜ *Status:* complete' in text
        assert '⬜ *AU:* complete' in text
        assert '⬜ *CA:* complete' in text
        assert '⬜ *OR:* complete' in text
        assert '⬜ *US:* complete' in text

    def test_mixed_statuses(self):
        statuses = {
            'overall': ('testing', None),
            'au': ('complete', None),
            'ca': ('testfail', None),
            'or': ('pr', None),
            'us': ('building', None),
        }

        response = format_slack_response(statuses)
        text = response['blocks'][0]['text']['text']

        assert '🟩 *Status:* testing' in text
        assert '⬜ *AU:* complete' in text
        assert '🟥 *CA:* testfail' in text
        assert '🟦 *OR:* pr' in text
        assert '🟩 *US:* building' in text

    def test_fetch_error_shows_red(self):
        statuses = {
            'overall': (None, 'HTTP 500'),
            'au': ('complete', None),
            'ca': ('complete', None),
            'or': ('complete', None),
            'us': ('complete', None),
        }

        response = format_slack_response(statuses)
        text = response['blocks'][0]['text']['text']

        assert '🟥 *Status:* HTTP 500' in text


class TestLambdaHandler:
    """Tests for the Lambda handler."""

    @patch('lambda_function.fetch_all_statuses')
    def test_handler_returns_200(self, mock_fetch):
        mock_fetch.return_value = {
            'overall': ('complete', None),
            'au': ('complete', None),
            'ca': ('complete', None),
            'or': ('complete', None),
            'us': ('complete', None),
        }

        event = {
            'body': 'command=/deploy-status',
            'isBase64Encoded': False,
        }

        response = lambda_handler(event, None)

        assert response['statusCode'] == 200
        assert response['headers']['Content-Type'] == 'application/json'

        body = json.loads(response['body'])
        assert body['response_type'] == 'ephemeral'
        assert 'blocks' in body


class TestIntegration:
    """Integration tests that hit real endpoints."""

    @pytest.mark.integration
    def test_fetch_real_statuses(self):
        """Test fetching from real status URLs."""
        statuses = fetch_all_statuses()

        assert 'overall' in statuses
        assert 'au' in statuses
        assert 'ca' in statuses
        assert 'or' in statuses
        assert 'us' in statuses

        # All should succeed (no errors)
        for region, (status, error) in statuses.items():
            assert error is None, f"{region} had error: {error}"
            assert status is not None, f"{region} had no status"
