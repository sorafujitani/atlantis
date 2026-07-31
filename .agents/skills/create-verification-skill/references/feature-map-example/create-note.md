# Create a note

Create note lets a user save a titled note from the CLI, reject incomplete input, and confirm the saved note from a second user-facing view.

## Sub-features

- `create-save` persists a title and body.
- `create-reject` rejects missing title or body with a non-zero exit.
- `create-list` shows the new note from a list or get command.

## How to get to it (user POV)

- Run `notes create --title <title> --body <body>` in a terminal.
- Run `notes create --title <title>` without `--body` to see validation fail.
- Run `notes list` or `notes get <id>` to confirm persistence.

## Driving it with notes CLI

Preconditions:

- `./bin/notes` is built and doctor reports a disposable `NOTES_DATA_DIR`.
- No note is titled `Release checklist`.

- **Save note.** Create a note. Run `NOTES_DATA_DIR=$DIR ./bin/notes create --title "Release checklist" --body "Tag and publish" --format json`. Exit code `0` and stdout contain the new note ID and title.
- **Confirm persistence.** List notes. Run `NOTES_DATA_DIR=$DIR ./bin/notes list --format json`. Stdout contains an object whose title is `Release checklist`.
- **Reject incomplete.** Omit body. Run `NOTES_DATA_DIR=$DIR ./bin/notes create --title "Discard me" --format json`. Exit code non-zero; `notes list` has no `Discard me`.
- **Proof.** Write stdout/stderr/exit files under `.atlantis/verify-artifacts/create-note/`. Re-run `notes get <id> --format json` and keep that output as the second view.

## Gotchas

- Titles are trimmed on save. Assert the stored title, not the raw argv string.
- A create exit code alone is insufficient proof. Re-read via `list` or `get`.
- Remove `Release checklist` during fixture cleanup, but retain proof artifacts.
