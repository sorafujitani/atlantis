# Agent integrations

Atlantis keeps the vault outside the repository. Set `ATLANTIS_BRAIN_DIR` to override the default `~/brain`.

## Pi

From a checkout:

```bash
make install-pi
```

The extension runs `atlantis brain index` on session start and after an agent settles, then injects `atlantis brain inject` into the system prompt. Run `/reload` after installation in an existing Pi session.

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
    ]
  }
}
```

Merge with existing hooks instead of replacing the entire object.

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

## Skill

`make install-skill` installs the complete `atlantis` skill into `~/.agents/skills/atlantis`. It is also installable from GitHub:

```bash
npx skills add https://github.com/sorafujitani/atlantis --skill atlantis -g -y
```
