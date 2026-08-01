package brain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFingerprintStableForIdenticalSources(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTestFile(t, root, "principles.md", "# Principles\n")
	writeTestFile(t, root, "workflow/safety.md", "# Safety\n")
	vault, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	first, err := vault.Fingerprint()
	if err != nil {
		t.Fatal(err)
	}
	second, err := vault.Fingerprint()
	if err != nil {
		t.Fatal(err)
	}
	if first == "" || first != second {
		t.Fatalf("fingerprint unstable: %q vs %q", first, second)
	}
}

func TestFingerprintIgnoresDerivedIndexesAndMtime(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTestFile(t, root, "principles.md", "# Principles\n")
	vault, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	before, err := vault.Fingerprint()
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, root, "index.md", "# Brain\n\n## Principles\n- [[principles]]\n\n")
	writeTestFile(t, root, "plans/index.md", "# Active Plans\n\n")
	now := time.Now()
	if err := os.Chtimes(filepath.Join(root, "principles.md"), now, now); err != nil {
		t.Fatal(err)
	}
	after, err := vault.Fingerprint()
	if err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatalf("fingerprint changed for derived indexes or mtime: %q vs %q", before, after)
	}
}

func TestCacheHitSkipsIndexRewrite(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTestFile(t, root, "principles.md", "# Principles\n")
	writeTestFile(t, root, "workflow/safety.md", "# Safety\n")
	vault, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	first, err := vault.ContextResult(false)
	if err != nil {
		t.Fatal(err)
	}
	indexPath := filepath.Join(root, "index.md")
	infoBefore, err := os.Stat(indexPath)
	if err != nil {
		t.Fatal(err)
	}
	second, err := vault.ContextResult(false)
	if err != nil {
		t.Fatal(err)
	}
	if first.Context != second.Context || first.Fingerprint != second.Fingerprint {
		t.Fatal("cache hit returned different context or fingerprint")
	}
	infoAfter, err := os.Stat(indexPath)
	if err != nil {
		t.Fatal(err)
	}
	if !infoBefore.ModTime().Equal(infoAfter.ModTime()) {
		t.Fatal("cache hit rewrote index.md")
	}
}

func TestCacheMissOnNoteChangeAndForceRebuild(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTestFile(t, root, "principles.md", "# Principles\n")
	vault, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	before, err := vault.ContextResult(false)
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, root, "workflow/new.md", "# New\n")
	after, err := vault.ContextResult(false)
	if err != nil {
		t.Fatal(err)
	}
	if before.Fingerprint == after.Fingerprint {
		t.Fatal("fingerprint did not change after adding a note")
	}
	if after.Context == before.Context {
		t.Fatal("context did not change after adding a note")
	}
	if !containsLink(after.Context, "workflow/new") {
		t.Fatalf("updated context missing new note:\n%s", after.Context)
	}

	// Ensure mtime can move forward on filesystems with coarse resolution.
	time.Sleep(10 * time.Millisecond)
	forced, err := vault.ContextResult(true)
	if err != nil {
		t.Fatal(err)
	}
	if forced.Fingerprint != after.Fingerprint || forced.Context != after.Context {
		t.Fatal("force rebuild changed fingerprint or context without source edits")
	}
	if _, err := os.Stat(filepath.Join(root, cacheFileName)); err != nil {
		t.Fatalf("cache missing after force rebuild: %v", err)
	}
}

func containsLink(context, link string) bool {
	return strings.Contains(context, "[["+link+"]]")
}
