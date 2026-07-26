.PHONY: build install install-skill install-pi install-integrations install-all test race lint vet vuln snapshot

INSTALL_DIR ?= $(HOME)/.local/bin
SKILLS_DIR ?= $(HOME)/.agents/skills
PI_EXTENSIONS_DIR ?= $(HOME)/.pi/agent/extensions
ATLANTIS_DATA_DIR ?= $(HOME)/.local/share/atlantis

build:
	go build -o bin/atlantis ./cmd/atlantis

install: build
	install -d $(INSTALL_DIR)
	install -m 0755 bin/atlantis $(INSTALL_DIR)/atlantis

install-skill:
	rm -rf $(SKILLS_DIR)/atlantis
	install -d $(SKILLS_DIR)
	cp -R .agents/skills/atlantis $(SKILLS_DIR)/atlantis

install-pi:
	install -d $(PI_EXTENSIONS_DIR)
	install -m 0644 integrations/pi/brain-context.ts $(PI_EXTENSIONS_DIR)/atlantis-brain.ts

install-integrations: install-skill install-pi
	install -d $(ATLANTIS_DATA_DIR)/hooks
	install -m 0755 integrations/hooks/brain-inject.sh $(ATLANTIS_DATA_DIR)/hooks/brain-inject.sh
	install -m 0755 integrations/hooks/brain-index.sh $(ATLANTIS_DATA_DIR)/hooks/brain-index.sh

install-all: install install-integrations

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
