### Opening a PR

1. Work from a clean branch or dedicated worktree based on the current remote default branch. Preserve unrelated local changes.
2. Simplify the diff before committing. Remove compatibility code, speculative abstractions, narration comments, and unrelated cleanup.
3. Run the full static and runtime verification required by the matched playbook.
4. Shape small ordered commits that tell the verification story. Stage only intended files.
5. Push an explicit refspec and open a focused PR. Do not add AI signatures or boilerplate footers.
6. Own follow-up. Watch checks, read review comments, and classify each as fix, dismiss with a concrete reason, or ask when product intent is genuinely ambiguous.

A delegate that opens a PR returns the URL and stops. The supervising agent owns CI and review follow-up.
