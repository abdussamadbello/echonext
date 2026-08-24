# EchoNext Agent Skills

Framework-specific [Agent Skills](https://code.claude.com/docs) for building
APIs with [EchoNext](https://github.com/abdussamadbello/echonext) — a type-safe
Go web framework on top of Echo with automatic OpenAPI generation, validation,
and a Cobra-based code generator.

Each skill is a directory with a portable `SKILL.md` (standard frontmatter
only), so it loads in Claude Code, the Claude Agent SDK, and any other harness
that understands the `SKILL.md` convention.

## Skills

| Skill | Use when… |
|-------|-----------|
| [`echonext-handlers`](echonext-handlers/SKILL.md) | Writing/editing a type-safe handler, request struct, or route registration |
| [`echonext-domain`](echonext-domain/SKILL.md) | Adding a new resource/feature/domain (model + service + handler + DTO) |
| [`echonext-cli`](echonext-cli/SKILL.md) | Scaffolding, running the dev loop, building, or managing the project via the CLI |
| [`echonext-openapi-security`](echonext-openapi-security/SKILL.md) | Configuring OpenAPI metadata or authentication (bearer/apiKey/oauth2/OIDC) |
| [`echonext-database`](echonext-database/SKILL.md) | Working with GORM models, the `Repository[T]` pattern, Atlas migrations, or seeds |
| [`echonext-testing`](echonext-testing/SKILL.md) | Writing tests with the contrib testing helpers (`APIClient`, `Suite`, fixtures) |
| [`echonext-integrations`](echonext-integrations/SKILL.md) | Adding WebSocket, GraphQL, or file-upload endpoints |
| [`echonext-middleware-config`](echonext-middleware-config/SKILL.md) | Registering or writing middleware, or loading configuration from YAML/env |

## How harnesses discover these

- **Claude Code:** the repo ships symlinks at `.claude/skills/<name>` pointing
  back to `skills/<name>`, so skills are discovered automatically with no
  duplicated content. (On a platform without symlink support, copy the
  `skills/<name>` directories into `.claude/skills/` instead.)
- **Claude Agent SDK:** point the SDK at this directory as a settings/skills
  source, or package `skills/` as a plugin.
- **Other harnesses / custom loaders:** scan `skills/*/SKILL.md` and read the
  `name` + `description` frontmatter.

## Contributing

See [`AUTHORING.md`](AUTHORING.md) for the format and conventions, then run
`python3 skills/scripts/validate_skills.py` before committing.
