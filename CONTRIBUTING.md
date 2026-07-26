# Contributing

## Prerequisites

- Go 1.25.2
- golangci-lint 2.12.0
- GoReleaser 2.x（release検証時）

## Setup

```bash
git clone https://github.com/sorafujitani/atlantis
cd atlantis
go mod download
make test
```

## Required checks

```bash
make test
make race
make vet
make lint
make build
```

外部CLIを使うintegration testではcredentialをfixtureやlogへ保存しないでください。新しいadapterは共通contract、cancel、invalid output、permission capabilityを検証してください。
