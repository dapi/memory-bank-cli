package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestLintRecognizesSourceComponentProjection(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"template/memory-bank", "memory-bank"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "template/memory-bank/components.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(root, "memory-bank/components.json")
	if err := os.Symlink("../template/memory-bank/components.json", marker); err != nil {
		t.Skip(err)
	}
	if err := os.WriteFile(filepath.Join(root, "memory-bank/README.md"), []byte("---\nstatus: draft\ndoc_function: index\npurpose: Navigate project.\n---\n# Memory Bank\n"), 0644); err != nil {
		t.Fatal(err)
	}
	var out, stderr bytes.Buffer
	if code := Run([]string{"lint", "--repo-root", root}, "test", &out, &stderr); code != 0 {
		t.Fatalf("source lint: %d %s %s", code, out.String(), stderr.String())
	}
	if err := os.WriteFile(filepath.Join(root, "memory-bank/.lock"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	stderr.Reset()
	if code := Run([]string{"lint", "--repo-root", root}, "test", &out, &stderr); code == 0 {
		t.Fatal("installed projection bypassed component state")
	}
}
