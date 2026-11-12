#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path

from rules import CheckResult, check_issue


def _result_to_dict(result: CheckResult) -> dict:
    return {
        "ok": result.ok,
        "skipped": result.skipped,
        "deficiencies": [
            {"id": d.id, "en": d.en, "zh": d.zh} for d in result.deficiencies
        ],
    }


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Check GitHub issue body quality.")
    parser.add_argument(
        "--body-file",
        type=Path,
        help="Path to a file containing the issue body.",
    )
    parser.add_argument(
        "--labels",
        default="",
        help="Comma-separated issue labels.",
    )
    args = parser.parse_args(argv)

    if args.body_file:
        body = args.body_file.read_text(encoding="utf-8")
    else:
        body = sys.stdin.read()

    labels = [part.strip() for part in args.labels.split(",") if part.strip()]
    result = check_issue(body, labels)
    print(json.dumps(_result_to_dict(result), ensure_ascii=False))
    return 0 if result.ok or result.skipped else 1


if __name__ == "__main__":
    raise SystemExit(main())
