# Agent integrations

Atlantis keeps the vault outside the repository. Set `ATLANTIS_BRAIN_DIR` to override the default `~/brain`.

Portable documents (`principles/`, `protocol/`) are managed in this repository under `brain/` and embedded in the binary. Everything else in the vault is local to the machine.

Recommended vault shape:

```text
~/brain/
  principles.md principles/   # installed by `atlantis brain seed`
  protocol/                   # installed by `atlantis brain seed`
  workflow/ codebase/ env/ plans/   # local notes
  index.md                    # generated
```

Install or refresh the repo-managed documents:

```bash
atlantis brain seed
```

`brain seed` replaces the repo-managed files and leaves local notes untouched. `atlantis brain index` follows symlinks, so a vault may still link notes in from elsewhere.

The Go binary embeds one minimal JavaScript adapter because the host extension APIs require JavaScript modules. Context generation and installation live in Go; the adapter only maps Pi/OMP and OpenCode lifecycle events to `atlantis brain context`.

Install all embedded adapters from a release binary:

```bash
atlantis integrations install
```

## Oh My Pi (omp)

Install the Atlantis skill and Brain extension into OMP's user directories:

```bash
make install-omp
```

Restart OMP, then invoke the skill explicitly or let OMP select it for a matching task:

```text
/skill:atlantis Investigate this failure and implement a verified fix
```

The Brain extension runs `atlantis brain context` on session boundaries. Run `/reload` in an existing OMP session after installation.

## Pi

From a checkout:

```bash
make install-pi
```

The extension runs `atlantis brain context` on session start and after an agent settles. Run `/reload` after installation in an existing Pi session.

## OpenCode

From a checkout:

```bash
make install-opencode
```

The plugin loads freshly generated context through `atlantis brain context` for each prompt. Restart OpenCode after installation.

## Claude Code

Install the hook scripts with `make install-integrations`, then merge these hooks into `~/.claude/settings.json`:

```json
{
  "hooks": {
    "SessionStart": [
      {
        "matcher": "startup|resume",
        "hooks": [
          {
            "type": "command",
            "command": "~/.local/share/atlantis/hooks/brain-inject.sh"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Write|Edit|MultiEdit",
        "hooks": [
          {
            "type": "command",
            "command": "~/.local/share/atlantis/hooks/brain-index.sh"
          }
        ]
      }
    ],
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "~/.local/share/atlantis/hooks/brain-stop-reflect.sh",
            "timeout": 10
          }
        ]
      }
    ]
  }
}
```

Merge with existing hooks instead of replacing the entire object.

The SessionStart hook runs `atlantis brain context`. `PostToolUse` keeps the derived index current. `Stop` emits a short self-improvement capture nudge via `additionalContext` (no model execution inside the hook).

## Codex

Install the same hook scripts, then merge this into `~/.codex/hooks.json`:

```json
{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "~/.local/share/atlantis/hooks/brain-inject.sh",
            "timeout": 10
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "~/.local/share/atlantis/hooks/brain-index.sh",
            "timeout": 10
          }
        ]
      }
    ]
  }
}
```

The session hook runs `atlantis brain context`. `PostToolUse` keeps the derived index current through the separate maintenance hook.

## Cursor

Install the Cursor hook scripts and materialize the always-applied Brain rule:

```bash
make install-cursor
```

Then merge these hooks into `~/.cursor/hooks.json` (preserve unrelated hooks such as herdr):

```json
{
  "version": 1,
  "hooks": {
    "sessionStart": [
      {
        "command": "~/.local/share/atlantis/hooks/cursor-brain-inject.sh",
        "timeout": 15
      }
    ],
    "afterFileEdit": [
      {
        "command": "~/.local/share/atlantis/hooks/cursor-brain-index.sh",
        "timeout": 10
      }
    ]
  }
}
```

`cursor-brain-inject.sh` syncs Brain context into:

1. `~/.cursor/rules/atlantis-brain.mdc` (`alwaysApply: true`)
2. a generated section in `~/AGENTS.md` between `<!-- atlantis-brain:start -->` and `<!-- atlantis-brain:end -->`

then emits Cursor `additional_context` JSON. The rule / AGENTS.md sync is the reliable injection path: Agents Window has dropped `sessionStart` `additional_context` due to a composer timing bug even when the Hooks channel reports a successful merge. `afterFileEdit` refreshes the derived index and rewrites those surfaces when content changes.

## Skill

`make install-skill` installs every skill under `.agents/skills/` into `~/.agents/skills/` (Atlantis operating mode, verification meta-skills, and `verify-atlantis`). The Atlantis skill alone is also installable from GitHub:

```bash
npx skills add https://github.com/sorafujitani/atlantis --skill atlantis -g -y
```

## Verification skills

Portable counterparts to project-local "drive the real app" skills:

| Skill | Role |
| --- | --- |
| `create-verification-skill` | Interview the repo; write `.agents/skills/verify-<app>/` with Launch / Doctor / Drive / Evidence / Cleanup and a feature map; prove one feature live before handoff |
| `maintain-verification-skill` | Source wave + live pass over an existing map; outcomes `clean` / `changed` / `blocked` |
| `verify-atlantis` | Dogfood map for this CLI (disposable brain dirs only) |

Evidence defaults to `<repo>/.atlantis/verify-artifacts/` (gitignored). When a project has `verify-<app>`, Atlantis **control** means follow that skill. Durable gotchas that would help a different task go to the brain via reflect; session run notes stay under `.atlantis/` and are not archived in the vault.
