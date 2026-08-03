### Authoring or modifying a skill

**You own the skill's voice.** Agent-facing prose has a higher bar than human prose; unhelpful sentences become instructions.

Decide create vs edit first via brain `[[protocol/agent-skills]]`. Prefer editing the owner skill. Create only when no skill owns the procedure and create criteria match. Prefer lint/script over a new skill.

1. Search personal + project skills for an owner. Owner found → edit it. None → create.
2. Use the **create-skill** skill for authoring mechanics (frontmatter, layout, placement).
3. Validate: `name` + `description` present, referenced files exist, cross-skill links resolve, description triggers are distinctive.
4. Test cases if structural; skip if subjective.
5. Install/sync to the harness path that will load it (`make install-skill` for Atlantis-shipped; copy/sync for personal `~/.agents/skills`).
6. Run **Opening a PR** when the skill lives in a repository with that delivery policy.

When in doubt, delete; prose earns its keep by changing a decision. Match tone to scope. Point at structural sources (types, READMEs, config); hardcoded details go stale (the **encode-lessons-in-structure** principle). Delegate to other skills by path; don't restate.

**Reply:** create|edit, path, summary, key design decisions, validation notes.
