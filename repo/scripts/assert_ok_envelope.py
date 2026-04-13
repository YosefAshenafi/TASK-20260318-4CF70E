#!/usr/bin/env python3
"""Validate PharmaOps standard API JSON envelope on stdin (integration tests)."""
from __future__ import annotations

import json
import sys


def main() -> int:
    try:
        raw = sys.stdin.read()
        data = json.loads(raw)
    except json.JSONDecodeError as e:
        print(f"[envelope] Invalid JSON: {e}", file=sys.stderr)
        return 1

    code = data.get("code")
    if code != "OK":
        print(f"[envelope] Expected code 'OK', got {code!r}: {raw[:500]!r}", file=sys.stderr)
        return 1

    request_id = data.get("requestId")
    if not request_id or not isinstance(request_id, str):
        print(f"[envelope] Missing or invalid requestId: {data!r}", file=sys.stderr)
        return 1

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
