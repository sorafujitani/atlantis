# Migration from model-orchestrator and sora-mode

Atlantis replaces both names. There are no compatibility aliases.

## Binary and configuration

```bash
rm -f ~/.local/bin/model-orchestrator
go install github.com/sorafujitani/atlantis/cmd/atlantis@latest
```

Rename configuration and state only when they exist:

```bash
mv ~/.config/model-orchestrator ~/.config/atlantis
mv ~/.local/state/model-orchestrator ~/.local/state/atlantis
```

Environment variables use the `ATLANTIS_` prefix instead of `MODEL_ORCHESTRATOR_`.

## Skill

Remove the old `model-orchestrator` and `sora-mode` skill directories, then install the unified skill:

```bash
rm -rf ~/.agents/skills/model-orchestrator ~/.claude/skills/sora-mode
npx skills add https://github.com/sorafujitani/atlantis --skill atlantis -g -y
```

## Brain integrations

Replace local brain hook implementations with the Atlantis hook scripts and Pi extension described in [integrations.md](./integrations.md). The vault content itself does not move unless `ATLANTIS_BRAIN_DIR` points somewhere else.
