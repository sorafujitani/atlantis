.PHONY: build install install-skill test race lint vet vuln snapshot

INSTALL_DIR ?= $(HOME)/.local/bin
SKILLS_DIR ?= $(HOME)/.agents/skills

build:
	go build -o bin/model-orchestrator ./cmd/model-orchestrator

install: build
	install -d $(INSTALL_DIR)
	install -m 0755 bin/model-orchestrator $(INSTALL_DIR)/model-orchestrator

install-skill:
	install -d $(SKILLS_DIR)/model-orchestrator/agents
	install -m 0644 .agents/skills/model-orchestrator/SKILL.md $(SKILLS_DIR)/model-orchestrator/SKILL.md
	install -m 0644 .agents/skills/model-orchestrator/agents/openai.yaml $(SKILLS_DIR)/model-orchestrator/agents/openai.yaml

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

vuln:
	govulncheck ./...

snapshot:
	goreleaser release --snapshot --clean
