# Agent integrations

Atlantis keeps the vault outside the repository. Set `ATLANTIS_BRAIN_DIR` to override the default `~/brain`.

Recommended vault shape:

```text
~/brain/
  common/          # git clone of agent-brain-common (principles, portable workflow, protocol)
  principles -> common/principles
  protocol -> common/protocol
  workflow/        # common file symlinks + local-only notes
  codebase/ env/ plans/   # local
  index.md         # generated
```

`atlantis brain index` follows symlinks and skips the `common/` checkout directory itself so notes are indexed via stable paths (`principles/`, `workflow/`, `protocol/`).

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

## Skill

`make install-skill` installs the complete `atlantis` skill into `~/.agents/skills/atlantis`. It is also installable from GitHub:

```bash
npx skills add https://github.com/sorafujitani/atlantis --skill atlantis -g -y
```
