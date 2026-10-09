from __future__ import annotations

import re
from dataclasses import dataclass
from typing import Literal

IssueKind = Literal["bug", "feature", "question", "unknown"]


@dataclass(frozen=True)
class Deficiency:
    id: str
    en: str
    zh: str


@dataclass(frozen=True)
class CheckResult:
    ok: bool
    skipped: bool
    deficiencies: tuple[Deficiency, ...]


_SECTION_HEADER = re.compile(r"^###\s+(.+)$", re.MULTILINE)
_PLACEHOLDER_STEP = re.compile(r"^\d+\.\s*\.{3}\s*$")


def parse_sections(body: str) -> dict[str, str]:
    sections: dict[str, str] = {}
    matches = list(_SECTION_HEADER.finditer(body))
    if not matches:
        return sections

    for i, match in enumerate(matches):
        header = match.group(1).strip()
        start = match.end()
        end = matches[i + 1].start() if i + 1 < len(matches) else len(body)
        content = body[start:end].strip()
        key = _header_key(header)
        if key:
            sections[key] = content
    return sections


def _header_key(header: str) -> str | None:
    lower = header.lower()
    if lower.startswith("release version"):
        return "version"
    if lower.startswith("description"):
        return "description"
    if lower.startswith("steps to reproduce"):
        return "steps"
    if lower.startswith("motivation"):
        return "motivation"
    return None


def detect_kind(sections: dict[str, str], labels: list[str]) -> IssueKind:
    normalized = {label.lower() for label in labels}
    if "bug" in normalized:
        return "bug"
    if "enhancement" in normalized:
        return "feature"
    if "steps" in sections:
        return "bug"
    if "motivation" in sections:
        return "feature"
    if "version" in sections and "description" in sections:
        return "question"
    return "unknown"


def is_template_shaped(sections: dict[str, str]) -> bool:
    return bool(sections)


def _invalid_version(value: str) -> bool:
    trimmed = value.strip()
    if len(trimmed) < 3:
        return True
    if trimmed.lower() == "latest":
        return True
    return False


def _is_placeholder_steps(steps: str) -> bool:
    lines = [line.strip() for line in steps.splitlines() if line.strip()]
    if not lines:
        return True
    return all(_PLACEHOLDER_STEP.match(line) for line in lines)


def check_issue(body: str, labels: list[str] | None = None) -> CheckResult:
    labels = labels or []
    sections = parse_sections(body)
    if not is_template_shaped(sections):
        return CheckResult(ok=True, skipped=True, deficiencies=())

    kind = detect_kind(sections, labels)
    deficiencies: list[Deficiency] = []

    version = sections.get("version", "")
    if _invalid_version(version):
        deficiencies.append(
            Deficiency(
                id="invalid_version",
                en="Provide a concrete release version (not `latest` or empty).",
                zh="请填写具体的软件版本号（不要使用 `latest` 或留空）。",
            )
        )

    description = sections.get("description", "")
    if len(description.strip()) < 40:
        deficiencies.append(
            Deficiency(
                id="short_description",
                en="Expand the description to at least 40 characters with clear context.",
                zh="请把描述写清楚（至少 40 个字符），便于维护者理解问题。",
            )
        )

    if kind == "bug":
        steps = sections.get("steps", "")
        if len(steps.strip()) < 20 or _is_placeholder_steps(steps):
            deficiencies.append(
                Deficiency(
                    id="incomplete_steps",
                    en="Add detailed, reproducible steps (not placeholder lines like `1. ...`).",
                    zh="请提供可复现的完整步骤（不要只保留 `1. ...` 这类占位内容）。",
                )
            )

    if kind == "feature":
        motivation = sections.get("motivation", "")
        if len(motivation.strip()) < 20:
            deficiencies.append(
                Deficiency(
                    id="short_motivation",
                    en="Explain why this feature helps the project (at least 20 characters).",
                    zh="请说明该功能对项目的价值（至少 20 个字符）。",
                )
            )

    ok = len(deficiencies) == 0
    return CheckResult(ok=ok, skipped=False, deficiencies=tuple(deficiencies))
