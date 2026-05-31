#!/usr/bin/env python3
"""Validate EchoNext Agent Skills for cross-harness portability.

Checks every skills/*/SKILL.md for:
  - parseable YAML frontmatter delimited by `---`
  - required `name` and `description` fields
  - `name` is kebab-case and equals the skill's directory name
  - `description` is non-empty and within sensible length bounds
  - any referenced `references/<file>` paths actually exist

Uses only the standard library so it runs in any environment / CI / hook.
Exits non-zero if any skill is invalid.
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

SKILLS_DIR = Path(__file__).resolve().parent.parent
KEBAB = re.compile(r"^[a-z0-9]+(-[a-z0-9]+)*$")
DESC_MIN, DESC_MAX = 20, 600


def parse_frontmatter(text: str) -> dict | None:
    """Return a flat dict of top-level scalar frontmatter keys, or None."""
    if not text.startswith("---"):
        return None
    parts = text.split("---", 2)
    if len(parts) < 3:
        return None
    body = parts[1]

    data: dict[str, str] = {}
    key = None
    buf: list[str] = []

    def flush():
        nonlocal key, buf
        if key is not None:
            val = " ".join(s.strip() for s in buf if s.strip())
            data[key] = val.strip().strip('"').strip("'")
        key, buf = None, []

    for raw in body.splitlines():
        line = raw.rstrip()
        if not line.strip():
            continue
        # Top-level key (no leading indentation).
        m = re.match(r"^([A-Za-z0-9_]+):\s*(.*)$", line)
        if m and not raw.startswith((" ", "\t")):
            flush()
            key = m.group(1)
            rest = m.group(2).strip()
            # Skip block/scalar indicators like ">-" or "|".
            if rest and rest not in (">", "|", ">-", "|-", ">+", "|+"):
                buf = [rest]
            else:
                buf = []
        else:
            # Continuation line of a folded/literal scalar.
            if key is not None:
                buf.append(line)
    flush()
    return data


def validate_skill(skill_md: Path) -> list[str]:
    errors: list[str] = []
    rel = skill_md.relative_to(SKILLS_DIR.parent)
    dir_name = skill_md.parent.name
    text = skill_md.read_text(encoding="utf-8")

    fm = parse_frontmatter(text)
    if fm is None:
        return [f"{rel}: missing or malformed YAML frontmatter (--- ... ---)"]

    name = fm.get("name", "")
    if not name:
        errors.append(f"{rel}: missing required field 'name'")
    else:
        if not KEBAB.match(name):
            errors.append(f"{rel}: name '{name}' is not kebab-case")
        if name != dir_name:
            errors.append(
                f"{rel}: name '{name}' does not match directory '{dir_name}'"
            )

    desc = fm.get("description", "")
    if not desc:
        errors.append(f"{rel}: missing required field 'description'")
    elif not (DESC_MIN <= len(desc) <= DESC_MAX):
        errors.append(
            f"{rel}: description length {len(desc)} outside "
            f"[{DESC_MIN}, {DESC_MAX}]"
        )

    # Referenced reference files must exist.
    for ref in re.findall(r"references/[A-Za-z0-9_./-]+\.md", text):
        if not (skill_md.parent / ref).exists():
            errors.append(f"{rel}: references missing file '{ref}'")

    return errors


def main() -> int:
    skill_files = sorted(SKILLS_DIR.glob("*/SKILL.md"))
    if not skill_files:
        print(f"No SKILL.md files found under {SKILLS_DIR}", file=sys.stderr)
        return 1

    all_errors: list[str] = []
    for skill_md in skill_files:
        all_errors.extend(validate_skill(skill_md))

    if all_errors:
        print("Skill validation FAILED:\n", file=sys.stderr)
        for e in all_errors:
            print(f"  - {e}", file=sys.stderr)
        return 1

    print(f"OK: {len(skill_files)} skill(s) valid.")
    for skill_md in skill_files:
        print(f"  - {skill_md.parent.name}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
