# Authoring EchoNext Agent Skills

These skills follow the open **Agent Skills** format: a directory containing a
`SKILL.md` file with YAML frontmatter and a markdown body. They are authored to
be **portable** — loadable by Claude Code, the Claude Agent SDK, and any other
harness that understands the `SKILL.md` convention.

## Rules

1. **One directory per skill.** The directory name must equal the `name` field
   in frontmatter, in `kebab-case` (e.g. `echonext-handlers`).

2. **Standard frontmatter only.** Use just these fields so non-Claude harnesses
   parse cleanly:
   ```yaml
   ---
   name: echonext-handlers
   description: >-
     Write and register type-safe EchoNext HTTP handlers. Use when adding,
     editing, or debugging a handler, request struct, or route registration.
   license: MIT            # optional
   metadata:               # optional
     version: 0.1.0
   ---
   ```
   - `name` (required): kebab-case, matches the directory.
   - `description` (required): third person, ≤ ~2 sentences. **Lead with what
     the skill does, then the trigger** ("Use when …"). The description is the
     only text most harnesses always have in context, so its trigger cues
     decide whether the skill fires.

3. **Progressive disclosure.** Keep each `SKILL.md` body focused (aim for under
   ~500 lines). Move long examples and exhaustive references into
   `references/*.md` and link them by relative path; harnesses load those only
   when needed.

4. **Real, compiling snippets.** Every Go example must match the current
   echonext API. Verify against the source before writing — signatures drift.

5. **No harness-specific instructions** in the body. Describe the framework, not
   how a particular agent should behave.

## Validating

Run the linter before committing:

```bash
python3 skills/scripts/validate_skills.py
```

It checks that each skill has the required frontmatter, the `name` matches its
directory, the `name` is kebab-case, the `description` is present and within
length bounds, and that any `references/` files mentioned actually exist.
