package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCapabilitiesWireContract(t *testing.T) {
	componentCode, componentUnsupported, caps := 1, `["components/v1","adoption/v1"]`, `["source-format/v1","legacy/v1"]`
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		componentCode, componentUnsupported, caps = 0, `[]`, `["source-format/v1","legacy/v1","components/v1","adoption/v1"]`
	}
	for _, tc := range []struct {
		args        []string
		code        int
		unsupported string
	}{
		{nil, 0, `[]`},
		{[]string{"--require", "legacy/v1", "--require", "source-format/v1"}, 0, `[]`},
		{[]string{"--require", "components/v1", "--require", "adoption/v1"}, componentCode, componentUnsupported},
	} {
		var out, err bytes.Buffer
		code := Run(append([]string{"capabilities"}, tc.args...), "test-version", &out, &err)
		want := `{"schema_version":1,"cli_version":"test-version","capabilities":` + caps + `,"unsupported":` + tc.unsupported + "}\n"
		if code != tc.code || out.String() != want || err.Len() != 0 {
			t.Fatalf("code=%d stdout=%s stderr=%s", code, out.String(), err.String())
		}
	}
	var out, err bytes.Buffer
	if code := Run([]string{"capabilities", "unexpected"}, "test", &out, &err); code != 2 || out.Len() != 0 {
		t.Fatalf("invalid syntax code=%d stdout=%s", code, out.String())
	}
}

func TestDoctorRepairRejectsUnsupportedSourceBeforeMutation(t *testing.T) {
	for _, dry := range []bool{false, true} {
		repo, source := t.TempDir(), t.TempDir()
		readme := []byte("---\ndoc_function: index\npurpose: Fixture.\nstatus: active\n---\n# Memory Bank\n")
		for _, root := range []string{repo, source} {
			if err := os.MkdirAll(filepath.Join(root, "memory-bank"), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "memory-bank/README.md"), readme, 0644); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(filepath.Join(source, "memory-bank-source.json"), []byte(`{"schema_version":99}`), 0644); err != nil {
			t.Fatal(err)
		}
		ref := commitCLISource(t, source, "unsupported")
		args := []string{"doctor", "--fix", "--repo-root", repo, "--source", source, "--template-version", "fixture", "--source-ref", ref}
		if dry {
			args = append(args, "--dry-run")
		}
		var out, stderr bytes.Buffer
		if code := Run(args, "test", &out, &stderr); code != 1 || !strings.Contains(stderr.String(), "unsupported source format") {
			t.Fatalf("code=%d stderr=%s", code, stderr.String())
		}
		got, err := os.ReadFile(filepath.Join(repo, "memory-bank/README.md"))
		if err != nil || !bytes.Equal(got, readme) {
			t.Fatal("repair changed README")
		}
		entries, err := os.ReadDir(filepath.Join(repo, "memory-bank"))
		if err != nil || len(entries) != 1 {
			t.Fatalf("repair changed tree: %v %v", entries, err)
		}
		entries, err = os.ReadDir(repo)
		if err != nil || len(entries) != 1 {
			t.Fatalf("repair changed root: %v %v", entries, err)
		}
	}
}
