# Minimize Reader Load

Maintainability is the work a reader must do to understand code. Two independent axes: layers to trace (indirections between question and answer) and state to hold (hidden or mutable context in the reader's head).

**Why:** Code is read far more than written. LOC, cyclomatic complexity, and "clean architecture" are proxies; reader load is the thing itself. A flat file with 50 globals can be as bad as a 6-layer adapter stack — guard both axes. This is the human analog of [[principles/guard-the-context-window]]: working memory is finite for readers too.

**Pattern:**
- Collapse layers that don't earn their keep: one-caller wrappers, adapters with no second implementation, indirection for a future that never came. Inline them.
- Shrink state scope: pure functions over mutations, locals over fields, fields over module state, module state over globals. Derive instead of sync.
- Name the invariant at the boundary, not in every consumer, so the reader learns it once.
- Before adding a layer or state, ask: does this reduce reader load somewhere else by at least as much?

**Test:** Can a new reader answer "where does X come from?" and "what can change X?" in under 30 seconds? If not, cut layers or cut state.
