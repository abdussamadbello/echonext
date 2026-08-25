---
name: echonext-setup
description: >-
  Install, scope, and restore the EchoNext framework and its agent skills in a
  project. Use when setting up EchoNext in a new or existing repository,
  moving globally-installed skills into a project, restoring skills from a
  skills-lock.json after cloning, or adding the skills that are not yet
  installed.
license: MIT
metadata:
  version: 0.1.0
---

# EchoNext project setup

Prepare a repository so both the framework and its agent skills are present and
committed. The framework install and the skill install are separate concerns —
a project can have either without the other.

## What cannot be bootstrapped

A skill is only available once it is installed, so the very first install on a
machine is always a manual step:

```bash
npx skills add abdussamadbello/echonext
```

Everything below covers what happens after that: scoping an existing global
install into a project, adding EchoNext to a codebase, filling in a partial
install, and restoring from a committed lockfile.

## Check the current state first

Three things vary independently. Determine all three before changing anything:

```bash
command -v echonext                                  # CLI present?
grep -q abdussamadbello/echonext go.mod 2>/dev/null  # framework a dependency?
ls skills-lock.json .claude/skills .agents/skills 2>/dev/null   # skills present?
```

`npx skills list` reports which skills are installed and at which scope.

## Case 1 — scope a global install into a project

Skills installed globally (`-g`) follow the person, not the repository. A
teammate or CI agent cloning the project gets nothing. Installing at project
scope writes `skills-lock.json`, which is what makes the set reproducible:

```bash
npx skills add abdussamadbello/echonext -s '*' -y
```

Run this from the repository root. It writes `skills-lock.json` plus the
installed content under `.agents/skills/` and `.claude/skills/`. Commit the
lockfile and ignore the installed content:

```gitignore
.agents/
.claude/skills/
```

Project-scoped skills do not replace the global ones; both resolve.

## Case 2 — add EchoNext to an existing Go project

Order matters: the framework is a Go dependency, the CLI is a binary, and the
skills are neither.

```bash
go get github.com/abdussamadbello/echonext@latest
go mod tidy
```

EchoNext v1.5.0 requires **Go 1.26 or newer** and Echo v5; typed handlers take
`*echo.Context`. Check `go version` first and report a mismatch rather than
attempting the upgrade silently — moving a codebase from Echo v4 to v5 changes
every handler signature.

Install the CLI and the skills as described in `echonext-cli` and Case 1. An
existing Echo application does not need restructuring: EchoNext wraps the Echo
instance, and typed routes coexist with standard Echo handlers.

## Case 3 — fill in a partial install

Installing a single skill with `-s` leaves the rest absent, and a missing skill
fails silently — the work just proceeds without its conventions. Compare what is
installed against the eight skills in the repository:

```bash
npx skills list
npx skills add abdussamadbello/echonext -s '*' -y   # adds the missing ones
```

The full set is `echonext-cli`, `echonext-domain`, `echonext-handlers`,
`echonext-database`, `echonext-openapi-security`, `echonext-middleware-config`,
`echonext-integrations`, and `echonext-testing`.

## Case 4 — restore after cloning

A repository containing `skills-lock.json` but no installed content restores the
exact recorded set, including each skill's `references/` directory:

```bash
npx skills experimental_install
```

Use this rather than a fresh `add` when the lockfile exists — `add` resolves to
the latest published skills, while the lockfile pins what the project was
developed against.

## Scaffolding a new project

`echonext init <name> --with-skills` creates the project and installs its skills
in one step, and writes a `.gitignore` that already excludes the installed
content while keeping `skills-lock.json` tracked. See `echonext-cli` for the
CLI's own installation.

## Verifying

```bash
npx skills list          # eight echonext-* skills
echonext --help          # CLI resolves
go build ./...           # framework dependency resolves
```
