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

Harness integrationは、Atlantis CLIが利用できない場合にも既存のBrain Markdownを直接読めることを検証してください。
