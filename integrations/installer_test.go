package integrations

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallWritesEmbeddedAdapterAndRemovesLegacyFile(t *testing.T) {
	t.Parallel()
	for _, target := range Targets() {
		t.Run(string(target), func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			legacy := filepath.Join(dir, "atlantis-brain.ts")
			if err := os.WriteFile(legacy, []byte("legacy"), 0o600); err != nil {
				t.Fatal(err)
			}

			installed, err := Install(target, dir)
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(installed)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(data, brainContextJS) {
				t.Fatal("installed adapter differs from embedded adapter")
			}
			if _, err := os.Stat(legacy); !os.IsNotExist(err) {
				t.Fatalf("legacy adapter still exists: %v", err)
			}

			if _, err := Install(target, dir); err != nil {
				t.Fatalf("second install failed: %v", err)
			}
		})
	}
}

func TestInstallRejectsUnknownTarget(t *testing.T) {
	t.Parallel()
	if _, err := Install(Target("unknown"), t.TempDir()); err == nil {
		t.Fatal("unknown target was accepted")
	}
}
