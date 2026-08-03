# Fail Fast Required Config

When a dependency is required in an environment, missing credentials or config are a **deploy error**. Fail at process start, not on the first request that happens to construct the client.

**Why:** Empty defaults (`?? ""`, optional env) plus lazy construction (transient DI, first-use init) let the process look healthy while the real failure waits in the request path. A best-effort `try/catch` around the call site does not cover constructor or DI resolution failures — those run before the method body. The user-visible symptom becomes a hard 500 on a core mutation, not a degraded side effect.

**Pattern:**
- If `enabled` / production requires the integration, validate secrets at boot the same way sibling required secrets are validated
- Treat "best-effort side effect" and "required to construct the handler" as different failure domains; catch only the former
- When DI lifetime is per-request, assume constructor errors surface as request errors until proven otherwise

**Anti-patterns:**
- Soft-defaulting a required key to `""` while leaving the feature enabled
- Logging "side effect failed" while the throw actually aborts the primary workflow
- Relying on health checks that never resolve the failing dependency

**Test:** Remove the secret in a staging-like boot. Does the process refuse to start, or does only one endpoint blow up later?
