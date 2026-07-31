# Build the Lever

When the work isn't trivial, build the tool that does it or proves it (codemod, script, generator) instead of working by hand.

**Why:** Throughput — the tool does the work the same way every time and reruns for free. Confidence — the tool is one artifact a reviewer can read and rerun to check the work; hand-done changes can only be re-verified by redoing them. A deterministic script turns "trust me" into "run this".

**Pattern:**
- Do the first unit by hand to learn the recipe, then build the tool. Prove it by rerunning on that unit and diffing against the hand-done version. Make it safe to rerun.
- A deterministic lever beats fan-out. If the tool can process every unit in one pass, run it yourself; don't fan out delegates to hand-apply what a script can do.
- When fanning out to subagents, write the lever as a shared instruction file (recipe, verification contract, do-not-touch fences) outside the delegates' write scope, so every delegate inherits the same hardened version.
- Applying this principle produces a file. Cited it with no codemod, script, or generator in the diff → not applied.
- Commit the lever when the work outlives the session, so the next run reruns it instead of redoing it.

**Balance:** The bar is triviality, not repetition. Build the smallest script that does or proves the job, never a framework ([[principles/laziness-protocol]]). Distinct from [[principles/encode-lessons-in-structure]] (recurring instruction → durable guardrail); this is throughput and reviewability on the work in front of you. For scripting the verification itself, see [[principles/prove-it-works]].
