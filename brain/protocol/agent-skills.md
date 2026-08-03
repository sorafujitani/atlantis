# Agent skills lifecycle

Create, edit, and improve Agent Skills as structural memory — not as a brain prose dump.

## Prefer order

1. Lint / script / test / hook / CI (enforce without reading)
2. **Edit** an existing skill that already owns the process
3. **Create** a new skill when no owner exists
4. Brain note (fact/gotcha/principle), never a substitute for a procedure
5. Discard one-off noise

## Create when all hold

- Multi-step agent procedure with a clear trigger phrase or situation
- No existing skill owns that process (search personal + project skills first)
- Same friction twice, or once with high impact if missed
- Future invocations should change decisions by reading the skill

Do not create for: single facts, one-repo constraints without procedure, principles, or anything a lint/script can enforce.

## Edit when

- An existing skill was selected (or should have been) and its steps are wrong, missing, or stale
- A correction is about *how* the skill runs, not a new domain

Prefer the smallest patch. Delete prose that does not change a decision.

## Improve cadence

- **Hot path** (`reflect`): create or edit immediately when criteria match; do not park the procedure in brain
- **Weekly / on request** (`meditate`): promote brain process recipes → skill; prune skill bloat; flag structural replacements
- **Authoring task**: follow Atlantis playbook `authoring-a-skill` / harness `create-skill`

## Placement

- Portable personal: `~/.agents/skills/<name>/` (sync to harness dirs if the host does not read that path)
- Project-local: `<repo>/.agents/skills/<name>/` when the procedure is repo-specific
- Atlantis-shipped: edit in the atlantis repo under `.agents/skills/`, then `make install-skill`

## Anti-patterns

- Brain note that is really a runbook ("do A then B then C")
- New skill that duplicates an existing one under a new name
- Skill that only restates a principle without an executable procedure
- Creating a skill when a lint/script would close the loop
