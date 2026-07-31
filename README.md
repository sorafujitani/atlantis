# Atlantis

[![CI](https://github.com/sorafujitani/atlantis/actions/workflows/ci.yml/badge.svg)](https://github.com/sorafujitani/atlantis/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/sorafujitani/atlantis)](https://go.dev/)
[![License](https://img.shields.io/github/license/sorafujitani/atlantis)](./LICENSE)

Atlantis lets you reuse the same coding-agent playbooks and local Markdown knowledge across Claude Code, Codex, Cursor, Pi, OMP, and OpenCode.

It does not run or manage agents. Execution, delegation, progress, inspection, and cancellation remain the responsibility of the active agent harness.

## What Atlantis provides

Atlantis has three layers with separate responsibilities:

| Layer | Responsibility |
| --- | --- |
| **CLI** | Initializes, seeds, indexes, validates, and reads a local Brain vault; cleans up completed plans; installs embedded host adapters. |
| **Agent Skill** | Provides task-specific playbooks and principle-grounded working conventions. |
| **Host integrations** | Refresh and inject Brain context at host lifecycle events by calling the CLI. |

This separation keeps durable knowledge portable and inspectable while leaving agent execution to the tool designed to perform it. Brain files are ordinary Markdown, so they remain readable without the Atlantis CLI.

## Quick start

Atlantis requires Go 1.25.8 or later. Install the CLI and create a validated Brain vault:

```bash
go install github.com/sorafujitani/atlantis/cmd/atlantis@latest
atlantis brain init
atlantis brain seed
atlantis brain check
```

`brain init` creates the local vault, `brain seed` installs the portable principles and protocols, and `brain check` validates links, reachability, and note size.

Optionally install Agent Skills (Atlantis playbooks plus verification meta-skills) so a compatible harness can select them:

```bash
make install-skill
# or:
npx skills add https://github.com/sorafujitani/atlantis --skill atlantis -g -y
```

Automatic context refresh and injection are host-specific. See [Agent integrations](./docs/integrations.md) for Claude Code, Codex, Cursor, Pi, OMP, OpenCode, and verification-skill setup.

## Brain data ownership

The Brain vault defaults to `~/brain`. Set `ATLANTIS_BRAIN_DIR` or pass `atlantis brain --dir <path> ...` to use another location.

| Content | Ownership |
| --- | --- |
| `principles.md`, `principles/`, `protocol/` | Repo-managed documents. `atlantis brain seed` replaces these paths with the versions embedded in the installed binary. |
| `workflow/`, `codebase/`, `env/`, `plans/` | Local knowledge owned by the user. Seeding leaves these paths untouched. |
| `index.md`, `plans/index.md` | Derived indexes generated from the Markdown vault. |

Markdown is the canonical data. `atlantis brain context` refreshes the derived indexes and prints the context consumed by integrations. Active plans belong in `plans/`; after the work is verified, extract any reusable knowledge and remove the plan with `atlantis brain plan finish <slug>`.

## What gets installed

| Action | Result |
| --- | --- |
| `go install ...@latest` | Installs the `atlantis` binary. |
| `atlantis brain init` and `brain seed` | Creates the local vault and installs the repo-managed Brain documents. |
| `npx skills add ...` | Installs the optional `atlantis` Agent Skill. |
| `make install-all` from a checkout | Updates the binary, repo-managed Brain documents, Agent Skill, and all supported host integration artifacts. |

To install everything available from a checkout, run:

```bash
make install-all
```

This does not merge `~/.claude/settings.json`, `~/.codex/hooks.json`, or `~/.cursor/hooks.json`. Follow [Agent integrations](./docs/integrations.md) to preserve existing hooks while adding Atlantis.

## CLI reference

| Command | Purpose |
| --- | --- |
| `atlantis brain init` | Create a Brain vault without replacing existing notes. |
| `atlantis brain seed` | Replace repo-managed Brain documents and rebuild indexes. |
| `atlantis brain index` | Regenerate derived indexes. |
| `atlantis brain context` | Refresh indexes and print agent context. |
| `atlantis brain check` | Validate links, reachability, and note size. |
| `atlantis brain plan finish <slug>` | Delete one completed plan and rebuild indexes. |
| `atlantis integrations install [omp\|pi\|opencode]` | Install one embedded adapter, or all three when no target is given. |
| `atlantis version` | Print build version information. |
| `atlantis completion <bash\|fish\|powershell\|zsh>` | Generate shell completion. |

Brain and integration commands support structured output through the global `--output json` flag. Run `atlantis <command> --help` for command-specific flags, including `brain --dir` and the integration install directory override.

## Documentation

- [Architecture](./docs/architecture.md)
- [Agent integrations](./docs/integrations.md)
- [Contributing](./CONTRIBUTING.md)
- [License](./LICENSE)

## License

[MIT](./LICENSE)
