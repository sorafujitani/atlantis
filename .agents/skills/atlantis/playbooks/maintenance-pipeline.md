### Maintenance pipeline

**You own the line, not the stages. Each stage proves its work or stops the line.** For recurring inbound reports (bug queues, crash reports, user feedback) worked as a staged flow: triage → reproduce → fix → verify.

Stages hand off an evidence bundle, never a summary. A stop verdict is a successful output: a defect caught here is cheap, the same defect after the next stage costs more. No stage passes work forward on plausibility.

1. **Triage.** Classify the report: read the attachments, check the reporter's version, search duplicates and recent history over the affected code (**how** + **why**). Output a ticket plus a handoff bundle (suspected component, affected versions, duplicate links, evidence read), or a stop verdict: expected behavior, duplicate, or not actionable, each backed by the evidence that justifies it.
2. **Reproduce.** Drive the real surface via **control** along the reporter's path. Reach the broken state twice; capture screenshots or recordings each time. Output the repro recipe plus captures, or a stop verdict: cannot reproduce, with what was tried. If a fix is already in flight as an open PR, run the repro before/after against it and report the result instead of duplicating the work.
3. **Hold for override.** Leave the repro verdict visible where a human can correct or reject it before the fix starts. Proceed without waiting only when the root cause is unambiguous and the change is reversible.
4. **Fix.** Run **Bug fix** seeded with the inherited repro and evidence; do not re-derive them. Stop verdict: change too risky, with the risk named.
5. **Verify.** Re-run the inherited repro before and after on the same surface, pair the evidence, and run **Opening a PR** with the bundle in the PR body.

Deterministic stages (attachment download, version checks, diff scripts, media capture) belong in scripts per the **build-the-lever** principle; spend agent judgment only where interpretation is required.

**Reply:** per stage, the verdict and its evidence; for a stopped line, which stage stopped and why that verdict holds.
