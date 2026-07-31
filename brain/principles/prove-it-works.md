# Prove It Works

Every task output must be verified by checking the real thing directly — not by inferring from proxies, self-reports, or "it compiles."

**Why:** Unverified work has unknown correctness. Indirect verification (file mtimes, output freshness, agent self-reports, cached screenshots) feels cheaper than direct observation, but acting on a wrong inference costs far more than checking the source.

**Pattern:** After completing any task, ask: "How do I prove this actually works?"

Check the real thing, not a proxy:
- Check process liveness directly, not indirectly through derived state
- Read the actual value, not a cached or derived representation
- When verification fails, suspect the observation method before suspecting the system

Match the tool to the evidence:
- Start with the highest-level tool designed for the signal being investigated.
- Escalate to a lower-level technique only when the current tool cannot answer a decision-relevant question.
- Before inspecting raw bytes, machine instructions, or memory, state the unanswered question, why the available semantic tools cannot answer it, and what result would distinguish the remaining hypotheses.
- If those three points are unclear, stop. The investigation has likely drifted from its goal or skipped a provided capability.
- Low-level inspection is a last resort, not proof of rigor.

Code / Features:
1. Build it (necessary but not sufficient)
2. Run it and exercise the actual feature path
3. Check the full chain: does data flow from input to output?
4. For integrations, test the full communication path end-to-end

Sequence work into the smallest units that end in a real check. Keep each unit green before advancing; for a bug, the canonical delivery order is a failing regression test followed by the fix.

Delegation: trust artifacts, not self-reports. Inspect the actual output artifact (git diff, file contents, runtime behavior), because agents report what they intended, not always what happened.
