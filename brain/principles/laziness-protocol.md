# Laziness Protocol

Writing code is cheap for the agent, which makes over-engineering easy. Counter it by borrowing a human maintainer's fatigue: aim for the most result with the least code and complexity.

**Why:** If a human developer would find the code exhausting to maintain, it is a bad solution — regardless of how little it cost to write.

**Pattern:**
- Subtract before adding. Remove dead weight and speculative edge cases before improving what remains.
- Maintain a flat hierarchy. If answering a question requires tracing more than 3 files or layers, flatten it.
- Consolidate decisions. Don't repeat the same choice in several places; one source of truth, pass the result as a simple flag.
- Minimize the diff. The smallest change that solves the problem; fewer lines beat "elegant" boilerplate.
- Question the threading. A task that passes a new signal through types, schemas, and pipelines usually has a more direct path.
- When a reference or abstraction has no novel value, delete it rather than preserving a stub.

Related: [[principles/minimize-reader-load]].
