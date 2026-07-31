# Subtract Before You Add

Remove obsolete structure before introducing its replacement. This is a sequencing rule, not only a preference for smaller diffs.

**Why:** Building on top of dead wrappers, duplicated validation, or transitional APIs makes the new design inherit constraints that no longer need to exist. Deleting first reveals the actual shape the replacement must serve.

**Pattern:**
1. Inventory dead paths, redundant layers, and superseded assumptions.
2. Delete or collapse them while the existing behavior contract remains pinned.
3. Re-evaluate the remaining problem.
4. Add only the structure the reduced system still requires.

**Test:** If the proposed addition disappeared, which existing code would still be removable? Remove that code first.
