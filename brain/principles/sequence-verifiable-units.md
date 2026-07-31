# Sequence Verifiable Units

Break multi-step work into the smallest ordered units that each end in a real check. Keep every unit green before starting the next.

**Why:** A final test after a large batch proves only the batch. It does not reveal which change established or broke the invariant, and it makes review and rollback unnecessarily expensive.

**Pattern:**
- Pin the starting contract.
- Order prerequisites before consumers.
- End each unit with static and runtime evidence appropriate to that unit.
- Commit accepted units separately when the history is part of the delivery.
- Stop and repair the first failing unit instead of accumulating more changes.

**Canonical sequences:**
- Bug fix: failing regression test, then the fix.
- Refactor: behavioral pin, subtraction, reshape, equivalence proof.
- Migration: shared contract, callers, legacy deletion, end-to-end verification.

**Test:** Could one unit be reviewed, reverted, or resumed without reconstructing the whole batch? If not, split it again.
