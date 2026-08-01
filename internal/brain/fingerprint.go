package brain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

const cacheFileName = ".atlantis-cache.json"

// ContextResult is the agent context plus the vault fingerprint that produced it.
type ContextResult struct {
	Context     string
	Fingerprint string
}

type cacheFile struct {
	Fingerprint string `json:"fingerprint"`
	Context     string `json:"context"`
}

func isDerivedMarkdown(relSlash string) bool {
	return relSlash == "index.md" || relSlash == "plans/index.md"
}

// Fingerprint returns a deterministic content hash of source Markdown in the vault.
// Derived indexes (index.md, plans/index.md) are excluded.
func (v *Vault) Fingerprint() (string, error) {
	type entry struct {
		path string
		sum  string
	}
	entries := []entry{}
	err := walkMarkdown(v.Root, func(path string) error {
		relative, relErr := filepath.Rel(v.Root, path)
		if relErr != nil {
			return relErr
		}
		relSlash := filepath.ToSlash(relative)
		if isDerivedMarkdown(relSlash) {
			return nil
		}
		file, openErr := os.Open(path) //nolint:gosec // path was discovered by walking the selected vault
		if openErr != nil {
			return fmt.Errorf("open %s: %w", relSlash, openErr)
		}
		hasher := sha256.New()
		_, copyErr := io.Copy(hasher, file)
		closeErr := file.Close()
		if copyErr != nil {
			return fmt.Errorf("hash %s: %w", relSlash, copyErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close %s: %w", relSlash, closeErr)
		}
		entries = append(entries, entry{path: relSlash, sum: hex.EncodeToString(hasher.Sum(nil))})
		return nil
	})
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("inspect brain directory: %w", err)
		}
		return "", fmt.Errorf("fingerprint brain directory: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].path < entries[j].path })

	hasher := sha256.New()
	for _, item := range entries {
		_, _ = io.WriteString(hasher, item.path)
		_, _ = hasher.Write([]byte{0})
		_, _ = io.WriteString(hasher, item.sum)
		_, _ = hasher.Write([]byte{0})
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (v *Vault) cachePath() string {
	return filepath.Join(v.Root, cacheFileName)
}

func (v *Vault) readCache() (cacheFile, bool, error) {
	data, err := os.ReadFile(v.cachePath()) //nolint:gosec // cache path is constructed inside the selected vault
	if errors.Is(err, os.ErrNotExist) {
		return cacheFile{}, false, nil
	}
	if err != nil {
		return cacheFile{}, false, fmt.Errorf("read brain cache: %w", err)
	}
	var cached cacheFile
	if err := json.Unmarshal(data, &cached); err != nil {
		return cacheFile{}, false, nil
	}
	if cached.Fingerprint == "" || cached.Context == "" {
		return cacheFile{}, false, nil
	}
	return cached, true, nil
}

func (v *Vault) writeCache(cached cacheFile) error {
	data, err := json.Marshal(cached)
	if err != nil {
		return fmt.Errorf("encode brain cache: %w", err)
	}
	return writeAtomicBytes(v.cachePath(), data)
}

func (v *Vault) cacheHit(force bool, fingerprint string) (ContextResult, bool, error) {
	if force {
		return ContextResult{}, false, nil
	}
	cached, ok, err := v.readCache()
	if err != nil || !ok || cached.Fingerprint != fingerprint {
		return ContextResult{}, false, err
	}
	indexPath := filepath.Join(v.Root, "index.md")
	data, err := os.ReadFile(indexPath) //nolint:gosec // index path is constructed inside the selected vault
	if err != nil {
		return ContextResult{}, false, nil
	}
	if cached.Context != contextPrefix+string(data) {
		return ContextResult{}, false, nil
	}
	return ContextResult{Context: cached.Context, Fingerprint: fingerprint}, true, nil
}

func writeAtomicBytes(path string, content []byte) error {
	current, err := os.ReadFile(path) //nolint:gosec // cache targets are constructed inside the selected vault
	if err == nil && string(current) == string(content) {
		return nil
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read current cache: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create cache directory: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".atlantis-cache-*")
	if err != nil {
		return fmt.Errorf("create temporary cache: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if _, err := temporary.Write(content); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write temporary cache: %w", err)
	}
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("set cache permissions: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary cache: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace cache: %w", err)
	}
	return nil
}
