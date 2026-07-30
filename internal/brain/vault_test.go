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

func TestIndexFollowsSymlinkedCommonCheckout(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	root := filepath.Join(base, "brain")
	common := filepath.Join(base, "common")
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, common, "principles.md", "# Principles\n- [[principles/prove-it-works]]\n")
	writeTestFile(t, common, "principles/prove-it-works.md", "# Prove It Works\n")
	writeTestFile(t, common, "workflow/git-push-safety.md", "# Git Push Safety\n")
	writeTestFile(t, common, "protocol/self-improvement.md", "# Self Improvement\n")

	if err := os.Symlink(common, filepath.Join(root, "common")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("common/principles", filepath.Join(root, "principles")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("common/principles.md", filepath.Join(root, "principles.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("common/protocol", filepath.Join(root, "protocol")); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "workflow"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../common/workflow/git-push-safety.md", filepath.Join(root, "workflow", "git-push-safety.md")); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, root, "workflow/local-only.md", "# Local Only\n")
	writeTestFile(t, root, "codebase/app.md", "# App\n")

	vault, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := vault.Index(); err != nil {
		t.Fatal(err)
	}

	index := readTestFile(t, root, "index.md")
	for _, expected := range []string{
		"[[principles]]",
		"[[protocol/self-improvement]]",
		"[[workflow/git-push-safety]]",
		"[[workflow/local-only]]",
		"[[codebase/app]]",
	} {
		if !strings.Contains(index, expected) {
			t.Fatalf("index.md does not contain %q:\n%s", expected, index)
		}
	}
	if strings.Contains(index, "[[common/") {
		t.Fatalf("common checkout leaked into index:\n%s", index)
	}

	report, err := vault.Check()
	if err != nil {
		t.Fatal(err)
	}
	if len(report.BrokenLinks) > 0 || len(report.Unreachable) > 0 {
		t.Fatalf("unexpected report: %#v\nindex:\n%s", report, index)
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

func TestContextRefreshesIndex(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTestFile(t, root, "principles.md", "# Principles\n")
	writeTestFile(t, root, "workflow/safety.md", "# Safety\n")

	vault, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	context, err := vault.Context()
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"Brain vault index.",
		"HARD SAFETY:",
		"[[workflow/safety]]",
	} {
		if !strings.Contains(context, expected) {
			t.Fatalf("context does not contain %q:\n%s", expected, context)
		}
	}
	if context != contextPrefix+readTestFile(t, root, "index.md") {
		t.Fatal("context differs from the canonical prefix plus refreshed index")
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
