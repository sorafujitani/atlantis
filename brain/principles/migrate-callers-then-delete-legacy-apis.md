# Migrate Callers Then Delete Legacy APIs

When replacing an internal API, inventory every caller, migrate them in one controlled wave, and delete the legacy API in that same change sequence.

**Why:** Parallel old and new paths create two contracts, two test surfaces, and an indefinite cleanup promise. Compatibility code is justified only by an external consumer that cannot migrate atomically.

**Pattern:**
1. Pin the behavior shared by the old and target APIs.
2. Find callers in code, configuration, tests, strings, and generated references.
3. Introduce the target contract.
4. Migrate and verify every caller.
5. Delete the old API and prove no references remain.

**Exception:** A real external compatibility boundary may require a shim. Give it an owner, observable usage, and an explicit removal condition.

**Test:** If both APIs remain at the end of the wave, identify the external constraint. Without one, the migration is incomplete.
