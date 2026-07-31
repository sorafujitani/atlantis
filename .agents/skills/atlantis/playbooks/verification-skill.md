### Verification skill

**You own scripted proof on the real user surface.** For generating or auditing a project-local `verify-<app>` skill with a feature map.

1. Creating or regenerating → run the **create-verification-skill** skill. Output lives under `.agents/skills/verify-<app>/`. Prove one mapped feature live before calling it done.
2. Auditing an existing map → run the **maintain-verification-skill** skill. Outcomes are `clean`, `changed`, or `blocked`. Ship corrections only inside the verification skill directory via **Opening a PR**.
3. When a project already has `verify-<app>`, **control** means follow that skill's Launch / Doctor / Drive / Evidence / Cleanup — do not invent a parallel drive path.
4. Evidence stays under `.atlantis/verify-artifacts/`. Durable gotchas that would improve a different task go to the brain via reflect; do not archive session run notes in the vault.

**Reply:** which meta-skill ran, the verify skill path, outcome (and PR URL if changed), evidence location for the proof that was driven.
