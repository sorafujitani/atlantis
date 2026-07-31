# Contributing

## Prerequisites

- Go 1.25.8
- golangci-lint 2.12.0
- GoReleaser 2.x（release検証時）

## Setup

```bash
git clone https://github.com/sorafujitani/atlantis
cd atlantis
go mod download
make install-git-hooks
make test
```

`make install-git-hooks` links `.githooks/pre-push` into `.git/hooks/`. Pushing `main` runs `make install install-skill` so `~/.local/bin/atlantis` and `~/.agents/skills/` match the checkout (build failure aborts the push).

## Required checks

```bash
make test
make race
make vet
make lint
make build
```

Harness integrationは、Atlantis CLIが利用できない場合にも既存のBrain Markdownを直接読めることを検証してください。
