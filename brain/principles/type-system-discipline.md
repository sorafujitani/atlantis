# Type System Discipline

The type checker is a proof assistant. Use it to eliminate impossible states, mismatched primitives, and unhandled variants at compile time.

**Why:** Anything let through as runtime data becomes a runtime failure the compiler could have stopped.

**Pattern:**
- Make illegal states unrepresentable. Model variants as sum types, not bags of optional fields. `{ completed: boolean; completedAt?: Date }` admits a meaningless combination; model `{ kind: 'open' } | { kind: 'done'; at: Date }` instead.
- Brand semantic primitives. `UserId` and `OrderId` are both strings underneath but must not be interchangeable. Validate once at creation, trust the type downstream.
- External data is untyped until parsed. Parse at every boundary: RPC payloads, JSON, CLI args, config, env vars, DB rows. See [[principles/boundary-discipline]] for where validation lives.
- Don't lie to the type system. Casts and unsafe assertions are deferred crashes; validate, narrow, or refine the model instead.
- Exhaustive matching is the compiler's job (`never`-typed binding in TS, unannotated `match` in Rust, sealed-class exhaustiveness in Kotlin).
- Derive types from authoritative schemas (protobuf, OpenAPI, GraphQL, migrations) instead of hand-rolling parallel shapes. See [[principles/encode-lessons-in-structure]].

**Tests:** Need a comment explaining when a field combination is valid → split into a sum type. Two same-typed args meaning different things → brand them. An `any` / `as` / `assertNotNull` → trace it to the boundary and validate there. A new variant next month wouldn't break compilation → the match isn't exhaustive.
