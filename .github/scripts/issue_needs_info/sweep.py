from __future__ import annotations

import argparse
import json
import os
import re
import urllib.error
import urllib.request
from datetime import datetime, timedelta, timezone

LABEL = "needs info"
REMINDER_MARKER = "<!-- needs-info-reminder -->"
REMIND_AFTER = timedelta(days=4)
CLOSE_AFTER = timedelta(days=7)

CLOSE_MESSAGE = (
    "This issue has been automatically closed because the required information "
    "was not provided within 7 days. Please reopen this issue after you add the "
    "missing details, or open a new issue with a complete report.\n\n"
    "---\n\n"
    "此 Issue 已被自动关闭：在 7 天内未补充必要信息。补充详情后欢迎重新打开此 "
    "Issue，或提交包含完整信息的新 Issue。"
)

REMINDER_MESSAGE = (
    "Reminder: this issue will be closed in about 3 days if the requested "
    "information is not provided. You can reply here and the timer will reset.\n\n"
    "---\n\n"
    "提醒：若仍未补充所需信息，此 Issue 将在约 3 天后自动关闭。在此回复后计时将会重置。"
)

_NEXT_LINK = re.compile(r'<([^<>]+)>;\s*rel="next"')


def decide(
    now: datetime,
    labeled_at: datetime,
    author_reply_dates: list[datetime],
    reminder_dates: list[datetime],
) -> str:
    if any(date > labeled_at for date in author_reply_dates):
        return "none"
    if now - labeled_at >= CLOSE_AFTER:
        return "close"
    if now - labeled_at >= REMIND_AFTER and not any(
        date > labeled_at for date in reminder_dates
    ):
        return "remind"
    return "none"


def parse_iso(value: str) -> datetime:
    return datetime.fromisoformat(value.replace("Z", "+00:00"))


def _headers() -> dict[str, str]:
    headers = {
        "Accept": "application/vnd.github+json",
        "X-GitHub-Api-Version": "2022-11-28",
        "User-Agent": "certimate-needs-info-sweep",
    }
    token = os.environ.get("GITHUB_TOKEN")
    if token:
        headers["Authorization"] = f"Bearer {token}"
    return headers


def request(
    method: str, url: str, payload: dict | None = None
) -> tuple[int, str, str]:
    data = json.dumps(payload).encode("utf-8") if payload is not None else None
    req = urllib.request.Request(url, data=data, method=method, headers=_headers())
    if data is not None:
        req.add_header("Content-Type", "application/json")
    try:
        with urllib.request.urlopen(req) as resp:
            return resp.status, resp.read().decode("utf-8"), resp.headers.get("Link", "")
    except urllib.error.HTTPError as err:
        body = err.read().decode("utf-8", "replace")
        raise RuntimeError(
            f"HTTP {err.code} from {method} {url}: {body}"
        ) from err


def paginate(url: str) -> list[dict]:
    items: list[dict] = []
    while url:
        status, body, link = request("GET", url)
        if status != 200:
            raise RuntimeError(f"HTTP {status} from GET {url}: {body}")
        items.extend(json.loads(body))
        match = _NEXT_LINK.search(link)
        url = match.group(1) if match else ""
    return items


def _last_labeled_at(events: list[dict]) -> datetime | None:
    labeled_at = None
    for event in events:
        if (
            event.get("event") == "labeled"
            and event.get("label", {}).get("name") == LABEL
        ):
            labeled_at = parse_iso(event["created_at"])
    return labeled_at


def _is_author_comment(comment: dict, author: str) -> bool:
    user = comment.get("user") or {}
    login = user.get("login") or ""
    if login != author:
        return False
    if user.get("type") == "Bot":
        return False
    if login.endswith("[bot]"):
        return False
    if (comment.get("body") or "").startswith(REMINDER_MARKER):
        return False
    return True


def create_comment(base: str, repo: str, number: int, body: str) -> None:
    request("POST", f"{base}/repos/{repo}/issues/{number}/comments", {"body": body})


def set_state(base: str, repo: str, number: int, state: str) -> None:
    request("PATCH", f"{base}/repos/{repo}/issues/{number}", {"state": state})


def remove_needs_info_label(base: str, repo: str, number: int) -> None:
    request("DELETE", f"{base}/repos/{repo}/issues/{number}/labels/needs%20info")


def sweep(dry_run: bool) -> None:
    repo = os.environ["GITHUB_REPOSITORY"]
    base = os.environ.get("GITHUB_API_URL", "https://api.github.com")
    now = datetime.now(timezone.utc)
    issues = paginate(
        f"{base}/repos/{repo}/issues?state=open&labels=needs+info&per_page=100"
    )
    for issue in issues:
        if "pull_request" in issue:
            continue
        number = issue["number"]
        comments = paginate(
            f"{base}/repos/{repo}/issues/{number}/comments?per_page=100"
        )
        events = paginate(
            f"{base}/repos/{repo}/issues/{number}/events?per_page=100"
        )
        labeled_at = _last_labeled_at(events)
        if labeled_at is None:
            print(f"#{number} none (no labeled event found)")
            continue
        author = issue["user"]["login"]
        author_reply_dates = [
            parse_iso(comment["created_at"])
            for comment in comments
            if _is_author_comment(comment, author)
        ]
        reminder_dates = [
            parse_iso(comment["created_at"])
            for comment in comments
            if (comment.get("body") or "").startswith(REMINDER_MARKER)
        ]
        action = decide(now, labeled_at, author_reply_dates, reminder_dates)
        print(f"#{number} {action}")
        if dry_run or action == "none":
            continue
        if action == "remind":
            create_comment(
                base, repo, number, REMINDER_MARKER + "\n\n" + REMINDER_MESSAGE
            )
        elif action == "close":
            create_comment(base, repo, number, CLOSE_MESSAGE)
            set_state(base, repo, number, "closed")
            remove_needs_info_label(base, repo, number)


def main() -> int:
    parser = argparse.ArgumentParser(description="Sweep stale needs-info issues.")
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Fetch and print planned actions without mutating anything.",
    )
    args = parser.parse_args()
    sweep(dry_run=args.dry_run)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
