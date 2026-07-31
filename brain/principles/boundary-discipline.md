# Boundary Discipline

Place validation, type narrowing, and error handling at system boundaries. Trust internal code unconditionally. Business logic lives in pure functions; the shell is thin and mechanical.

**Why:** Validation scattered throughout is noisy, redundant, and gives a false sense of safety. Concentrating it at boundaries means each piece of data is validated exactly once. Logic tangled with framework wiring can't be tested without the framework.

**The Pattern:**
- **At boundaries** (CLI args, config files, external APIs, network protocols): validate, return errors, handle defensively
- **Inside the system**: typed data, error propagation, no re-validation. Trust the types.

**Applications:**
- Parse raw input into typed state when it enters the system; do not carry unvalidated representations inward
- Validate config at load time, not inside business logic
- Do not repeat nil or shape checks after the boundary established the invariant
- Keep boundary adapters thin; business logic should be pure and framework-independent

**The Tests:**
- "Is this data crossing a system boundary right now?" If not, validation is redundant
- "Can this be a pure function that the shell just calls?" If yes, extract it
