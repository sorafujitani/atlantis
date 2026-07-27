# Atlantis

[![CI](https://github.com/sorafujitani/atlantis/actions/workflows/ci.yml/badge.svg)](https://github.com/sorafujitani/atlantis/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/sorafujitani/atlantis)](https://go.dev/)
[![License](https://img.shields.io/github/license/sorafujitani/atlantis)](./LICENSE)

Atlantis provides a portable operating mode and persistent Markdown context for coding agents.

## Components

- **Agent Skill.** Task-specific playbooks and principle-grounded working conventions.
- **Brain.** Direct Markdown context reads with deterministic index maintenance, wikilink validation, and transient plan cleanup.

The Brain vault is user data and stays outside this repository. Atlantis does not start or manage coding-agent executions; the active harness owns execution, progress, inspection, and cancellation.

## Install

```bash
go install github.com/sorafujitani/atlantis/cmd/atlantis@latest
atlantis version
```

From a checkout, install the binary, shared Skill, and harness integrations:

```bash
make install-all
```

Claude Code, Codex, Pi, OMP, and OpenCode setup is documented in [docs/integrations.md](./docs/integrations.md).

## Brain

The default vault is `~/brain`. Override it with `ATLANTIS_BRAIN_DIR` or `--dir`.

```bash
atlantis brain init
atlantis brain index
atlantis brain check
atlantis brain plan finish <slug>
```

Harness integrations read `index.md` and linked notes directly. The Go CLI owns only deterministic index maintenance, validation, and safe plan cleanup.

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

## Safety boundaries

- Brain Markdown remains the canonical data.
- Context reads do not depend on the Atlantis binary.
- Index refresh is best-effort for host integrations.
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
