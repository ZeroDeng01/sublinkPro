#!/usr/bin/env python3
"""
SublinkPro API Helper
Handles authentication, content-type routing, and response envelope checking.
"""
import json
import os
import sys
import urllib.parse
import urllib.request
from typing import Dict, List, Tuple

def parse_args(args: List[str]) -> Tuple[str, str, Dict, Dict, Dict, bool]:
    """Parse CLI arguments into method, path, and parameter dicts."""
    if len(args) < 2:
        print("Usage: scripts/sublink.py <METHOD> <PATH> [--form k=v ...] [--json k=v ...] [--query k=v ...] [--raw]", file=sys.stderr)
        sys.exit(1)

    method = args[0].upper()
    path = args[1]
    form_params = {}
    json_params = {}
    query_params = {}
    raw_mode = False

    i = 2
    while i < len(args):
        arg = args[i]
        if arg == '--raw':
            raw_mode = True
            i += 1
        elif arg in ('--form', '--json', '--query'):
            if i + 1 >= len(args):
                print(f"Error: {arg} requires a value", file=sys.stderr)
                sys.exit(1)

            param_str = args[i + 1]
            if '=' not in param_str:
                print(f"Error: {arg} value must be in format key=value", file=sys.stderr)
                sys.exit(1)

            key, value = param_str.split('=', 1)

            if arg == '--form':
                form_params[key] = value
            elif arg == '--json':
                # Try to parse as JSON literal, fall back to string
                try:
                    json_params[key] = json.loads(value)
                except json.JSONDecodeError:
                    json_params[key] = value
            else:  # --query
                query_params[key] = value

            i += 2
        else:
            print(f"Error: unknown argument {arg}", file=sys.stderr)
            sys.exit(1)

    if form_params and json_params:
        print("Error: cannot mix --form and --json in the same request", file=sys.stderr)
        sys.exit(1)

    return method, path, form_params, json_params, query_params, raw_mode

def main():
    base_url = os.getenv('SUBLINK_BASE_URL')
    api_key = os.getenv('SUBLINK_API_KEY')

    if not base_url:
        print(
            "SUBLINK_BASE_URL is not set.\n"
            "\n"
            "This is the address of your running SublinkPro instance. Set it like:\n"
            "  export SUBLINK_BASE_URL=http://localhost:8000      # local instance\n"
            "  export SUBLINK_BASE_URL=http://your-server-ip:8000 # remote instance\n"
            "\n"
            "Don't have SublinkPro running yet? The skill can deploy one for you —\n"
            "just ask: \"help me deploy SublinkPro\".",
            file=sys.stderr,
        )
        sys.exit(1)

    if not api_key:
        print(
            "SUBLINK_API_KEY is not set.\n"
            "\n"
            "Get one from the SublinkPro web UI, then set it:\n"
            "  1. Open the web UI in a browser (your SUBLINK_BASE_URL address)\n"
            "  2. Sign in. On a fresh install the default is admin / 123456\n"
            "     (change this password right away).\n"
            "  3. Go to Settings -> Access Keys -> Create.\n"
            "  4. Copy the key (it is shown only once). It looks like prefix_xxx_yyy.\n"
            "  5. export SUBLINK_API_KEY=prefix_xxx_yyy",
            file=sys.stderr,
        )
        sys.exit(1)

    method, path, form_params, json_params, query_params, raw_mode = parse_args(sys.argv[1:])

    # Build URL with query params
    url = base_url.rstrip('/') + path
    if query_params:
        query_string = urllib.parse.urlencode(query_params)
        url = url + ('&' if '?' in url else '?') + query_string

    # Prepare request
    headers = {
        'X-API-Key': api_key,
        'User-Agent': 'SublinkPro-CLI/1.0'
    }

    data = None
    if form_params:
        headers['Content-Type'] = 'application/x-www-form-urlencoded'
        data = urllib.parse.urlencode(form_params).encode('utf-8')
    elif json_params:
        headers['Content-Type'] = 'application/json'
        data = json.dumps(json_params).encode('utf-8')

    # Make request
    try:
        req = urllib.request.Request(url, data=data, headers=headers, method=method)
        with urllib.request.urlopen(req) as response:
            body = response.read().decode('utf-8')

            # Raw mode: print body as-is (for /c/ subscription output, SSE streams)
            if raw_mode:
                print(body)
                sys.exit(0)

            # Parse JSON envelope
            try:
                result = json.loads(body)
            except json.JSONDecodeError:
                print(f"Error: response is not valid JSON", file=sys.stderr)
                print(body, file=sys.stderr)
                sys.exit(1)

            # Check envelope code (200=success, 500=error, etc.)
            code = result.get('code', 0)
            msg = result.get('msg', '')
            data = result.get('data')

            if code in (400, 403, 404, 500):
                print(f"API Error ({code}): {msg}", file=sys.stderr)
                if code == 403:
                    print(
                        "\nThis is an authorization failure. Common causes:\n"
                        "  - SUBLINK_API_KEY is missing, wrong, or expired -> re-create it\n"
                        "    in the web UI (Settings -> Access Keys) and re-export it.\n"
                        "  - SUBLINK_BASE_URL points at a different instance than the key.\n"
                        "  - Note: /settings/ai-assistant* endpoints reject API-key auth by\n"
                        "    design (web-login session only) and will always 403 here.",
                        file=sys.stderr,
                    )
                sys.exit(1)

            # Success: print data field (or full result if no data field)
            if data is not None:
                print(json.dumps(data, indent=2, ensure_ascii=False))
            else:
                print(json.dumps(result, indent=2, ensure_ascii=False))
            sys.exit(0)

    except urllib.error.HTTPError as e:
        print(f"HTTP Error {e.code}: {e.reason}", file=sys.stderr)
        try:
            body = e.read().decode('utf-8')
            result = json.loads(body)
            msg = result.get('msg', '')
            if msg:
                print(f"API Error: {msg}", file=sys.stderr)
        except Exception:
            pass
        if e.code in (401, 403):
            print(
                "\nAuthorization failed. Check that SUBLINK_API_KEY is set, correct, and\n"
                "not expired (re-create it in the web UI under Settings -> Access Keys),\n"
                "and that SUBLINK_BASE_URL points at the same instance that issued it.",
                file=sys.stderr,
            )
        sys.exit(1)
    except urllib.error.URLError as e:
        print(f"Could not reach SublinkPro at {base_url}", file=sys.stderr)
        print(f"  reason: {e.reason}", file=sys.stderr)
        print(
            "\nChecklist:\n"
            "  - Is a SublinkPro instance actually running at that address?\n"
            "  - Is SUBLINK_BASE_URL using the right host and port (default 8000)?\n"
            "  - For a remote host: is the port exposed and not blocked by a firewall?\n"
            "  - Confirm it's up:  curl " + base_url.rstrip('/') + "/api/v1/version\n"
            "\nDon't have an instance yet? Ask me to deploy SublinkPro for you\n"
            "(docker, docker-compose, or the install script; local or remote).",
            file=sys.stderr,
        )
        sys.exit(1)
    except Exception as e:
        print(f"Unexpected Error: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == '__main__':
    main()
