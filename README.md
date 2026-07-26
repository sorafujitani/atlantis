# Atlantis

[![CI](https://github.com/sorafujitani/atlantis/actions/workflows/ci.yml/badge.svg)](https://github.com/sorafujitani/atlantis/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/sorafujitani/atlantis)](https://go.dev/)
[![License](https://img.shields.io/github/license/sorafujitani/atlantis)](./LICENSE)

Atlantis is a local control plane for coding agents. One binary provides bounded multi-model orchestration and a portable persistent-context vault. The bundled Agent Skill adds principle-grounded playbooks under the same `atlantis` name.

Provider credentials remain owned by existing agent CLIs. Atlantis does not copy API keys, OAuth tokens, or native session files.

## Components

```text
                         atlantis
                 /          |          \
        operating mode   orchestration   brain
          playbooks       supervisor      vault
                              |             |
                  Codex / Claude / others  Claude / Codex / Pi
```

- **Operating mode.** Task-specific playbooks for features, bugs, refactors, investigations, performance work, and autonomous runs.
- **Orchestration.** `single`, `advisor`, `orchestrator`, and `hybrid` execution over locally authenticated agent CLIs.
- **Brain.** Deterministic indexes, wikilink validation, compact context injection, and transient plan cleanup.

The vault is user data and stays outside this repository. The repository contains only the portable machinery and empty-vault behavior.

## Install

```bash
go install github.com/sorafujitani/atlantis/cmd/atlantis@latest
atlantis version
atlantis doctor
```

From a checkout, install the binary, skill, and Pi integration together:

```bash
make install-all
```

Claude Code, Codex, and Pi setup is documented in [docs/integrations.md](./docs/integrations.md). Existing `model-orchestrator` and `sora-mode` users should follow [docs/migration.md](./docs/migration.md).

## Orchestration

Atlantis delegates execution to existing CLIs and normalizes their structured results. Built-in adapters cover Codex CLI, Claude Code, Grok Build, OpenCode, Cursor Agent, and GitHub Copilot.

```bash
atlantis run --mode single --permission read \
  "Summarize this repository in three lines"

atlantis run --mode advisor --permission read \
  "Compare two designs and escalate only the consequential decision"

atlantis run --mode orchestrator \
  "Delegate independent investigations and synthesize the result"

atlantis run --profile grok \
  "Run this assignment with the locally authenticated Grok Build CLI"
```

JSON output is available for host agents:

```bash
atlantis --output json run --mode hybrid "<objective>"
```

Interrupted runs use append-only state and can be inspected or resumed:

```bash
atlantis status <run-id>
atlantis inspect <run-id>
atlantis resume <run-id>
atlantis cancel <run-id>
```

## Brain

The default vault is `~/brain`. Override it with `ATLANTIS_BRAIN_DIR` or `--dir`.

```bash
atlantis brain init
atlantis brain index
atlantis brain check
atlantis brain inject
atlantis brain plan finish <slug>
```

A vault has this shape:

```text
brain/
├── index.md
├── principles.md
├── principles/
├── workflow/
├── env/
├── codebase/
└── plans/
```

Only active work belongs in `plans/`. Once implementation is verified complete, extract reusable knowledge and delete the plan with `atlantis brain plan finish`.

## Configuration

Resolution order is flag, environment, config file, then defaults.

```text
$ATLANTIS_CONFIG
$XDG_CONFIG_HOME/atlantis/config.toml
~/.config/atlantis/config.toml
```

Run state defaults to `$XDG_STATE_HOME/atlantis` or `~/.local/state/atlantis`. See [config.example.toml](./config.example.toml).

```toml
default_profile = "default"

[models.premium]
adapter = "codex"
model = ""

[models.standard]
adapter = "claude"
model = ""
fallback = ["premium"]

[profiles.default]
mode = "hybrid"
orchestrator = "premium"
executor = "standard"
advisor = "premium"
worker = "standard"
reviewer = "premium"
max_workers = 4
max_calls = 20
max_advisor_calls = 1
max_retries = 1
max_duration = "30m"
```

The Grok Build adapter uses native JSON Schema output, captures usage and session IDs, and supports native resume. Read assignments run in `plan` permission mode; write assignments use `acceptEdits`. Grok's own memory and nested subagents are disabled so Atlantis remains the context and delegation owner.

Custom adapters execute a command directly without a shell. Supported placeholders are `{prompt}`, `{cwd}`, `{model}`, and `{session}`.

## Safety boundaries

- Credentials and raw provider reasoning are not persisted.
- Child-process environment variables use an allowlist.
- Read assignments are bounded and parallelizable.
- Write assignments use an exclusive lane.
- Commands are executed as argument arrays, not shell strings.
- Run state is private by default and supports convergent resume.
- Brain plan deletion accepts only a single safe slug inside the vault's `plans/` directory.

## Development

```bash
make test
make race
make vet
make lint
make build
```

See [docs/architecture.md](./docs/architecture.md), [docs/integrations.md](./docs/integrations.md), and [CONTRIBUTING.md](./CONTRIBUTING.md).

## License

MIT
