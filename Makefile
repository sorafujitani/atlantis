.PHONY: build install install-skill install-omp install-pi install-opencode install-integrations install-all test race lint vet vuln snapshot

INSTALL_DIR ?= $(HOME)/.local/bin
SKILLS_DIR ?= $(HOME)/.agents/skills
OMP_SKILLS_DIR ?= $(HOME)/.omp/agent/skills
OMP_EXTENSIONS_DIR ?= $(HOME)/.omp/agent/extensions
PI_EXTENSIONS_DIR ?= $(HOME)/.pi/agent/extensions
OPENCODE_PLUGINS_DIR ?= $(HOME)/.config/opencode/plugins
ATLANTIS_DATA_DIR ?= $(HOME)/.local/share/atlantis
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
BUILD_DATE ?= $(shell git show -s --format=%cI HEAD 2>/dev/null || echo unknown)
VERSION_PACKAGE := github.com/sorafujitani/atlantis/internal/cli
LD_FLAGS := -X $(VERSION_PACKAGE).Version=$(VERSION) -X $(VERSION_PACKAGE).Commit=$(COMMIT) -X $(VERSION_PACKAGE).Date=$(BUILD_DATE)

build:
	go build -trimpath -ldflags "$(LD_FLAGS)" -o bin/atlantis ./cmd/atlantis

install: build
	install -d $(INSTALL_DIR)
	install -m 0755 bin/atlantis $(INSTALL_DIR)/atlantis

install-skill:
	rm -rf $(SKILLS_DIR)/atlantis
	install -d $(SKILLS_DIR)
	cp -R .agents/skills/atlantis $(SKILLS_DIR)/atlantis

install-omp:
	rm -rf $(OMP_SKILLS_DIR)/atlantis
	install -d $(OMP_SKILLS_DIR)
	cp -R .agents/skills/atlantis $(OMP_SKILLS_DIR)/atlantis
	install -d $(OMP_EXTENSIONS_DIR)
	install -m 0644 integrations/pi/brain-context.ts $(OMP_EXTENSIONS_DIR)/atlantis-brain.ts

install-pi:
	install -d $(PI_EXTENSIONS_DIR)
	install -m 0644 integrations/pi/brain-context.ts $(PI_EXTENSIONS_DIR)/atlantis-brain.ts

install-opencode:
	install -d $(OPENCODE_PLUGINS_DIR)
	install -m 0644 integrations/opencode/atlantis-brain.js $(OPENCODE_PLUGINS_DIR)/atlantis-brain.js

install-integrations: install-skill install-omp install-pi install-opencode
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
