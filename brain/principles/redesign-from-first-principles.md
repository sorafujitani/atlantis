# Redesign From First Principles

When integrating a change, don't bolt it onto the existing design. Redesign as if the change had been a foundational assumption from the start. The result should be the most elegant solution that would have emerged if we'd known about this requirement on day one.

- Read all affected files and understand the current design holistically
- Ask: "if we were writing this from scratch with this new requirement, what would we build?"
- When the answer is not obvious, sketch or prototype two or three concrete alternatives and compare them before committing
- Propagate the change through every reference — types, docs, examples, rationale sections
- Think holistically, then deliver incrementally

This is the method for preserving option value when integrating changes into an existing design.
