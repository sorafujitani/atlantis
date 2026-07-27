# Atlantis

[![CI](https://github.com/sorafujitani/atlantis/actions/workflows/ci.yml/badge.svg)](https://github.com/sorafujitani/atlantis/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/sorafujitani/atlantis)](https://go.dev/)
[![License](https://img.shields.io/github/license/sorafujitani/atlantis)](./LICENSE)

Atlantis is a portable operating mode and persistent-context maintenance tool for coding agents. The repository retains a bounded multi-model supervisor, but its model-routing CLI entry points are disabled. The bundled Agent Skill adds principle-grounded playbooks under the same `atlantis` name.

Provider credentials remain owned by existing agent CLIs. Atlantis does not copy API keys, OAuth tokens, or native session files.

## Components

```text
                         atlantis
                 /          |          \
        operating mode   orchestration   brain
          playbooks      dormant source   vault
                              |             |
                  Codex / Claude / OMP / others  Claude / Codex / Pi / OMP
```

- **Operating mode.** Task-specific playbooks for features, bugs, refactors, investigations, performance work, and autonomous runs.
- **Orchestration.** The `single`, `advisor`, `orchestrator`, and `hybrid` implementation remains in source, but model-starting CLI commands are disabled.
- **Brain.** Direct Markdown context reads with deterministic index maintenance, wikilink validation, and transient plan cleanup.

The vault is user data and stays outside this repository. The repository contains only the portable machinery and empty-vault behavior.

## Install

```bash
go install github.com/sorafujitani/atlantis/cmd/atlantis@latest
atlantis version
```

From a checkout, install the binary plus shared, OMP, and Pi integrations together:

```bash
make install-all
```

OMP, OpenCode, Claude Code, Codex, and Pi setup is documented in [docs/integrations.md](./docs/integrations.md). Existing `model-orchestrator` and `sora-mode` users should follow [docs/migration.md](./docs/migration.md).

## Orchestration

Model routing is intentionally disabled. These commands return a disabled error without starting a provider CLI:

```bash
atlantis run
atlantis resume
atlantis eval
```

The engine, adapters, profiles, fallback, and resume implementation remain tested in source. Historical append-only state can still be inspected or cancelled:

```bash
atlantis status <run-id>
atlantis inspect <run-id>
atlantis cancel <run-id>
```

`status` reports `interrupted` when events say a run was active but its supervisor process is gone. `cancel` signals a live supervisor or records an idempotent terminal cancellation for an interrupted run.

## Brain

The default vault is `~/brain`. Override it with `ATLANTIS_BRAIN_DIR` or `--dir`.

```bash
atlantis brain init
atlantis brain index
atlantis brain check
atlantis brain plan finish <slug>
```

Harness integrations read `index.md` and linked notes directly. The Go CLI owns only deterministic index maintenance, validation, and safe plan cleanup.

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

Run state defaults to `$XDG_STATE_HOME/atlantis` or `~/.local/state/atlantis`. The dormant routing configuration is retained in [config.example.toml](./config.example.toml) so the implementation remains testable.

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

The Oh My Pi (`omp`) adapter runs headless JSON mode (`-p --mode json`), captures session headers (`{"type":"session","id":...}`) and usage (`input`/`output`/`cost.total`), and supports `--model` and `--resume`. OMP has no `--json-schema` flag, so Atlantis injects the result schema through `--append-system-prompt`. Read assignments use a strict read-only tool allowlist with `--approval-mode always-ask`; write assignments use a coding tool allowlist that excludes `task`, `hub`, and memory tools (`retain`, `recall`, `reflect`, `memory_edit`, `learn`, `manage_skill`) with `--approval-mode yolo`, so Atlantis remains the orchestration and memory owner.

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
