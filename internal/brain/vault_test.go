package brain

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestVaultLifecycle(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	vault, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := vault.Init(); err != nil {
		t.Fatal(err)
	}

	writeTestFile(t, root, "principles.md", "# Principles\n- [[principles/prove-it-works]]\n")
	writeTestFile(t, root, "principles/prove-it-works.md", "# Prove It Works\n")
	writeTestFile(t, root, "plans/current.md", "# Current\n")
	if err := vault.Index(); err != nil {
		t.Fatal(err)
	}

	rootIndex := readTestFile(t, root, "index.md")
	for _, expected := range []string{"[[plans/index]]", "[[principles]]"} {
		if !strings.Contains(rootIndex, expected) {
			t.Fatalf("index.md does not contain %q:\n%s", expected, rootIndex)
		}
	}
	if !strings.Contains(readTestFile(t, root, "plans/index.md"), "[[plans/current]]") {
		t.Fatal("active plan is missing from plans index")
	}

	report, err := vault.Check()
	if err != nil {
		t.Fatal(err)
	}
	if len(report.BrokenLinks) > 0 || len(report.Unreachable) > 0 {
		t.Fatalf("unexpected report: %#v", report)
	}

	if err := vault.FinishPlan("current"); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(readTestFile(t, root, "plans/index.md"), "current") {
		t.Fatal("finished plan remains indexed")
	}
	if _, err := os.Stat(filepath.Join(root, "plans", "current.md")); !os.IsNotExist(err) {
		t.Fatalf("finished plan still exists: %v", err)
	}
}

func TestCheckReportsStructuralProblems(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	vault, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := vault.Init(); err != nil {
		t.Fatal(err)
	}

	writeTestFile(t, root, "orphan.md", "# Orphan\n- [[missing]]\n")
	report, err := vault.Check()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(report.BrokenLinks, []string{"orphan -> missing"}) {
		t.Fatalf("broken links = %#v", report.BrokenLinks)
	}
	if !reflect.DeepEqual(report.Unreachable, []string{"orphan"}) {
		t.Fatalf("unreachable = %#v", report.Unreachable)
	}
}

func TestFinishPlanRejectsTraversal(t *testing.T) {
	t.Parallel()
	vault, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := vault.FinishPlan("../outside"); err == nil {
		t.Fatal("FinishPlan accepted path traversal")
	}
}

func writeTestFile(t *testing.T, root, relative, content string) {
	t.Helper()
	path := filepath.Join(root, relative)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func readTestFile(t *testing.T, root, relative string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, relative)) //nolint:gosec // test path is confined to t.TempDir
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
