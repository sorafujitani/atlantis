# Multi-phase Plan

Create a temporary implementation plan. Planning is the deliverable; do not implement in the same invocation.

1. Skip the plan for an obvious one or two-file change.
2. Read the relevant brain principles and affected code, tests, conventions, and dependencies.
3. Resolve only ambiguities that materially change product intent or architecture.
4. Compare two or three concrete alternatives for consequential architecture decisions.
5. Write the plan under `brain/plans/<slug>/` or as `brain/plans/<slug>.md`.
6. Keep each phase independently verifiable and usually limited to two or three files.
7. Include context, scope, constraints, chosen alternative, applicable skills, phase links, and project-level verification in the overview.
8. Include the goal, affected files, key data contracts, static checks, runtime proof, and edge cases in each phase.
9. Present the plan and stop.

Plans are transient. Once implementation is verified complete, extract reusable lessons and run `atlantis brain plan finish <slug>`. Delete abandoned or superseded plans instead of archiving them.
