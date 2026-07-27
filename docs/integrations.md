# Agent integrations

Atlantis keeps the vault outside the repository. Set `ATLANTIS_BRAIN_DIR` to override the default `~/brain`.

## Oh My Pi (omp)

Install and authenticate the `omp` CLI yourself (Atlantis does not install it or rewrite `~/.omp`):

```bash
# Choose one installation method.
curl -fsSL https://omp.sh/install | sh
# brew install can1357/tap/omp
# bun install -g @oh-my-pi/pi-coding-agent

omp --version
```

Authenticate in OMP itself when using OMP directly or testing the dormant adapter source. Atlantis does not copy credentials or native session files.

### Dormant Atlantis → OMP adapter

The built-in adapter/model/profile named `omp` remains in source, but `atlantis run`, `atlantis resume`, and `atlantis eval` are disabled. It cannot launch OMP through the installed CLI.

The retained adapter invokes `omp -p --mode json` with `--model` and `--resume` when configured, injects the Atlantis result JSON schema via `--append-system-prompt`, and maps permissions to OMP tool allowlists and unattended approval modes. Its tests keep the dormant implementation from silently decaying.

### OMP → Atlantis

Install the Atlantis skill and Brain extension into OMP's user directories:

```bash
make install-omp
```

Restart OMP, then invoke the skill explicitly or let OMP select it for a matching task:

```text
/skill:atlantis Investigate this failure and implement a verified fix
```

The Atlantis skill does not route work to another model. If the user explicitly asks for delegation, use OMP's own facilities directly.

The Brain extension refreshes derived indexes on session boundaries and reads `brain/index.md` directly. Run `/reload` in an existing OMP session after installation.

## Pi

From a checkout:

```bash
make install-pi
```

The extension asks Atlantis to refresh derived indexes on session start and after an agent settles, then reads `brain/index.md` directly with the host filesystem API. If the maintenance command is unavailable, the existing durable index remains readable. Run `/reload` after installation in an existing Pi session.

## OpenCode

From a checkout:

```bash
make install-opencode
```

The plugin refreshes derived indexes on startup and after tool execution, then reads `brain/index.md` directly with Node's filesystem API. Index refresh is best-effort; an unavailable Atlantis binary does not prevent existing brain context from loading. Restart OpenCode after installation.

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

The hook reads `brain/index.md` directly. `PostToolUse` keeps the derived index current through the separate maintenance hook.

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

The session hook reads `brain/index.md` directly. `PostToolUse` keeps the derived index current through the separate maintenance hook.

## Skill

`make install-skill` installs the complete `atlantis` skill into `~/.agents/skills/atlantis`. It is also installable from GitHub:

```bash
npx skills add https://github.com/sorafujitani/atlantis --skill atlantis -g -y
```
