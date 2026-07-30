// Package brain manages a portable, link-indexed agent memory vault.
package brain

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	defaultDirectory = "brain"
	maxNoteLines     = 50
	contextPrefix    = "Brain vault index. Read only the linked notes relevant to the task before acting. " +
		"Capture lessons per [[protocol/self-improvement]] (common vs local). " +
		"HARD SAFETY: never create a GitHub repo unless the user explicitly asked in this conversation " +
		"(no gh repo create / API create / unsolicited remote bootstrap). " +
		"Never invent git/GitHub identity — use only existing git config, authenticated account data, " +
		"or explicit user values; no fabricated noreply emails, authors, or collaborators.\n\n"
)

var wikiLinkPattern = regexp.MustCompile(`\[\[([^]#|]+)`)

// Vault owns index generation and validation for one brain directory.
type Vault struct {
	Root string
}

// Report describes structural problems in a vault.
type Report struct {
	Files       int      `json:"files"`
	Reachable   int      `json:"reachable"`
	BrokenLinks []string `json:"broken_links"`
	Unreachable []string `json:"unreachable"`
	Oversized   []string `json:"oversized"`
}

// DefaultRoot resolves ATLANTIS_BRAIN_DIR, then defaults to ~/brain.
func DefaultRoot() (string, error) {
	if root := os.Getenv("ATLANTIS_BRAIN_DIR"); root != "" {
		return filepath.Abs(root)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, defaultDirectory), nil
}

// New validates and normalizes a vault root.
func New(root string) (*Vault, error) {
	if root == "" {
		var err error
		root, err = DefaultRoot()
		if err != nil {
			return nil, err
		}
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve brain directory: %w", err)
	}
	return &Vault{Root: absolute}, nil
}

// Init creates an empty vault without replacing existing notes.
func (v *Vault) Init() error {
	for _, name := range []string{"codebase", "env", "plans", "principles", "workflow"} {
		if err := os.MkdirAll(filepath.Join(v.Root, name), 0o700); err != nil {
			return fmt.Errorf("create %s directory: %w", name, err)
		}
	}
	principles := filepath.Join(v.Root, "principles.md")
	if _, err := os.Stat(principles); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(principles, []byte("# Principles\n"), 0o600); err != nil {
			return fmt.Errorf("create principles index: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("inspect principles index: %w", err)
	}
	return v.Index()
}

// Index regenerates plans/index.md and the root index deterministically.
func (v *Vault) Index() error {
	if _, err := os.Stat(v.Root); err != nil {
		return fmt.Errorf("inspect brain directory: %w", err)
	}
	if err := v.indexPlans(); err != nil {
		return err
	}
	return v.indexRoot()
}

// Context refreshes the derived index and returns the complete agent context.
func (v *Vault) Context() (string, error) {
	if err := v.Index(); err != nil {
		return "", err
	}
	data, err := os.ReadFile(filepath.Join(v.Root, "index.md"))
	if err != nil {
		return "", fmt.Errorf("read brain index: %w", err)
	}
	return contextPrefix + string(data), nil
}

func (v *Vault) indexPlans() error {
	plansDir := filepath.Join(v.Root, "plans")
	entries, err := os.ReadDir(plansDir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read plans directory: %w", err)
	}

	links := []string{}
	for _, entry := range entries {
		switch {
		case entry.Type().IsRegular() && strings.HasSuffix(entry.Name(), ".md") && entry.Name() != "index.md":
			links = append(links, "plans/"+strings.TrimSuffix(entry.Name(), ".md"))
		case entry.IsDir():
			overview := filepath.Join(plansDir, entry.Name(), "overview.md")
			if info, statErr := os.Stat(overview); statErr == nil && info.Mode().IsRegular() {
				links = append(links, "plans/"+entry.Name()+"/overview")
			} else if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
				return fmt.Errorf("inspect plan overview: %w", statErr)
			}
		}
	}
	sort.Strings(links)
	return writeIndex(filepath.Join(plansDir, "index.md"), "Active Plans", links)
}

func (v *Vault) indexRoot() error {
	entries, err := os.ReadDir(v.Root)
	if err != nil {
		return fmt.Errorf("read brain directory: %w", err)
	}

	sections := []string{}
	standalone := []string{}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") || name == "common" {
			continue
		}
		full := filepath.Join(v.Root, name)
		isDir, dirErr := dirEntryIsDir(entry, full)
		if dirErr != nil {
			return dirErr
		}
		if isDir {
			sections = append(sections, name)
			continue
		}
		if name != "index.md" && isMarkdownNote(entry, full) {
			standalone = append(standalone, strings.TrimSuffix(name, ".md"))
		}
	}
	sort.Strings(sections)
	sort.Strings(standalone)

	var output strings.Builder
	output.WriteString("# Brain\n")
	entrypoints := map[string]bool{}
	for _, section := range sections {
		links, entrypoint, sectionErr := v.sectionLinks(section)
		if sectionErr != nil {
			return sectionErr
		}
		if len(links) == 0 {
			continue
		}
		output.WriteString("\n## " + title(section) + "\n")
		for _, link := range links {
			output.WriteString("- [[" + link + "]]\n")
		}
		if entrypoint != "" {
			entrypoints[entrypoint] = true
		}
	}

	other := []string{}
	for _, link := range standalone {
		if !entrypoints[link] {
			other = append(other, link)
		}
	}
	if len(other) > 0 {
		output.WriteString("\n## Other\n")
		for _, link := range other {
			output.WriteString("- [[" + link + "]]\n")
		}
	}
	output.WriteString("\n")
	return writeAtomic(filepath.Join(v.Root, "index.md"), output.String())
}

func (v *Vault) sectionLinks(section string) ([]string, string, error) {
	rootEntrypoint := filepath.Join(v.Root, section+".md")
	if info, err := os.Stat(rootEntrypoint); err == nil && info.Mode().IsRegular() {
		return []string{section}, section, nil
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, "", fmt.Errorf("inspect section entrypoint: %w", err)
	}

	nestedEntrypoint := filepath.Join(v.Root, section, "index.md")
	if info, err := os.Stat(nestedEntrypoint); err == nil && info.Mode().IsRegular() {
		return []string{section + "/index"}, "", nil
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, "", fmt.Errorf("inspect nested entrypoint: %w", err)
	}

	links := []string{}
	sectionRoot := filepath.Join(v.Root, section)
	err := walkMarkdown(sectionRoot, func(path string) error {
		relative, relErr := filepath.Rel(v.Root, path)
		if relErr != nil {
			return relErr
		}
		links = append(links, filepath.ToSlash(strings.TrimSuffix(relative, ".md")))
		return nil
	})
	if err != nil {
		return nil, "", fmt.Errorf("walk section %q: %w", section, err)
	}
	sort.Strings(links)
	return links, "", nil
}

// Check validates links, reachability from index.md, and note size.
func (v *Vault) Check() (Report, error) {
	files, err := v.markdownFiles()
	if err != nil {
		return Report{}, err
	}
	report := Report{Files: len(files), BrokenLinks: []string{}, Unreachable: []string{}, Oversized: []string{}}
	graph := make(map[string][]string, len(files))
	for key, path := range files {
		data, readErr := os.ReadFile(path) //nolint:gosec // path was discovered by walking the selected vault
		if readErr != nil {
			return Report{}, fmt.Errorf("read %s: %w", key, readErr)
		}
		if key != "index" && key != "plans/index" && countLines(data) > maxNoteLines {
			report.Oversized = append(report.Oversized, key)
		}
		for _, match := range wikiLinkPattern.FindAllStringSubmatch(string(data), -1) {
			target := resolveLink(key, strings.TrimSuffix(match[1], ".md"), files)
			if target == "" {
				report.BrokenLinks = append(report.BrokenLinks, key+" -> "+match[1])
				continue
			}
			graph[key] = append(graph[key], target)
		}
	}

	seen := map[string]bool{}
	queue := []string{"index"}
	for len(queue) > 0 {
		key := queue[0]
		queue = queue[1:]
		if seen[key] {
			continue
		}
		seen[key] = true
		queue = append(queue, graph[key]...)
	}
	for key := range files {
		if !seen[key] {
			report.Unreachable = append(report.Unreachable, key)
		}
	}
	report.Reachable = len(seen)
	sort.Strings(report.BrokenLinks)
	sort.Strings(report.Unreachable)
	sort.Strings(report.Oversized)
	return report, nil
}

// FinishPlan removes one completed or abandoned plan and rebuilds indexes.
func (v *Vault) FinishPlan(slug string) error {
	if slug == "" || slug == "." || slug == ".." || filepath.Base(slug) != slug {
		return errors.New("plan slug must be a single safe path segment")
	}
	plansDir := filepath.Join(v.Root, "plans")
	candidates := []string{filepath.Join(plansDir, slug), filepath.Join(plansDir, slug+".md")}
	found := ""
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			if found != "" {
				return fmt.Errorf("plan %q is ambiguous", slug)
			}
			found = candidate
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("inspect plan %q: %w", slug, err)
		}
	}
	if found == "" {
		return fmt.Errorf("plan %q does not exist", slug)
	}
	if err := os.RemoveAll(found); err != nil {
		return fmt.Errorf("remove plan %q: %w", slug, err)
	}
	return v.Index()
}

func (v *Vault) markdownFiles() (map[string]string, error) {
	files := map[string]string{}
	err := walkMarkdown(v.Root, func(path string) error {
		relative, relErr := filepath.Rel(v.Root, path)
		if relErr != nil {
			return relErr
		}
		slash := filepath.ToSlash(relative)
		// common/ is the git checkout backing symlinked sections; index via principles/workflow/protocol only.
		if slash == "common" || strings.HasPrefix(slash, "common/") {
			return nil
		}
		files[strings.TrimSuffix(slash, ".md")] = path
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk brain directory: %w", err)
	}
	return files, nil
}

func dirEntryIsDir(entry os.DirEntry, fullPath string) (bool, error) {
	if entry.IsDir() {
		return true, nil
	}
	if entry.Type()&os.ModeSymlink == 0 {
		return false, nil
	}
	info, err := os.Stat(fullPath)
	if err != nil {
		return false, fmt.Errorf("stat %s: %w", fullPath, err)
	}
	return info.IsDir(), nil
}

func isMarkdownNote(entry os.DirEntry, fullPath string) bool {
	if !strings.HasSuffix(entry.Name(), ".md") {
		return false
	}
	if entry.Type().IsRegular() {
		return true
	}
	if entry.Type()&os.ModeSymlink == 0 {
		return false
	}
	info, err := os.Stat(fullPath)
	return err == nil && info.Mode().IsRegular()
}

// walkMarkdown visits .md files under root, following directory symlinks once.
func walkMarkdown(root string, visit func(path string) error) error {
	info, err := os.Stat(root)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		if strings.HasSuffix(root, ".md") {
			return visit(root)
		}
		return nil
	}

	seen := map[string]bool{}
	var walk func(string) error
	walk = func(dir string) error {
		real, err := filepath.EvalSymlinks(dir)
		if err != nil {
			return err
		}
		if seen[real] {
			return nil
		}
		seen[real] = true

		entries, err := os.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			name := entry.Name()
			if strings.HasPrefix(name, ".") {
				continue
			}
			full := filepath.Join(dir, name)
			// Skip nested common checkouts when walking the vault root.
			if name == "common" && filepath.Clean(dir) == filepath.Clean(root) {
				continue
			}
			isDir, err := dirEntryIsDir(entry, full)
			if err != nil {
				return err
			}
			if isDir {
				if err := walk(full); err != nil {
					return err
				}
				continue
			}
			if isMarkdownNote(entry, full) {
				if err := visit(full); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return walk(root)
}

func resolveLink(source, target string, files map[string]string) string {
	target = filepath.ToSlash(filepath.Clean(target))
	directory := filepath.ToSlash(filepath.Dir(source))
	if directory != "." {
		relative := filepath.ToSlash(filepath.Clean(filepath.Join(directory, target)))
		if _, ok := files[relative]; ok {
			return relative
		}
	}
	if _, ok := files[target]; ok {
		return target
	}
	return ""
}

func writeIndex(path, heading string, links []string) error {
	var output strings.Builder
	output.WriteString("# " + heading + "\n")
	for _, link := range links {
		output.WriteString("- [[" + link + "]]\n")
	}
	output.WriteString("\n")
	return writeAtomic(path, output.String())
}

func writeAtomic(path, content string) error {
	current, err := os.ReadFile(path) //nolint:gosec // index targets are constructed inside the selected vault
	if err == nil && string(current) == content {
		return nil
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read current index: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create index directory: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".atlantis-index-*")
	if err != nil {
		return fmt.Errorf("create temporary index: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if _, err := temporary.WriteString(content); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write temporary index: %w", err)
	}
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("set index permissions: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary index: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace index: %w", err)
	}
	return nil
}

func countLines(data []byte) int {
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	lines := 0
	for scanner.Scan() {
		lines++
	}
	return lines
}

func title(value string) string {
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + value[1:]
}
