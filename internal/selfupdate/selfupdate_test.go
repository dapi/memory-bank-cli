package selfupdate

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testAsset = "memory-bank-cli-linux-amd64"

func TestCurrentReleaseDoesNotDownloadOrReplace(t *testing.T) {
	server, calls := releaseServer(t, "v1.0.0", []byte("ignored"), true)
	defer server.Close()
	destination := writeDestination(t, []byte("old"))
	var stdout, stderr bytes.Buffer
	if code := service(server, destination, &stdout, &stderr).Run(context.Background()); code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), "already up to date") || calls.asset != 0 || calls.sums != 0 {
		t.Fatalf("code=%d stdout=%q stderr=%q calls=%+v", code, stdout.String(), stderr.String(), calls)
	}
	assertBytes(t, destination, []byte("old"))
}

func TestVerifiedUpdateReplacesRunningExecutable(t *testing.T) {
	server, _ := releaseServer(t, "v1.1.0", []byte("#!/bin/sh\nprintf 'memory-bank-cli v1.1.0\\n'\n"), true)
	defer server.Close()
	destination := writeDestination(t, []byte("old"))
	var stdout, stderr bytes.Buffer
	if code := service(server, destination, &stdout, &stderr).Run(context.Background()); code != 0 || stderr.Len() != 0 {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if got, err := os.ReadFile(destination); err != nil || !bytes.Contains(got, []byte("v1.1.0")) {
		t.Fatalf("updated executable=%q err=%v", got, err)
	}
	if !strings.Contains(stdout.String(), "Updated memory-bank-cli to v1.1.0") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestFailuresPreserveOriginalExecutable(t *testing.T) {
	for _, test := range []struct {
		name  string
		setup func(*Service)
		valid bool
	}{
		{"checksum mismatch", func(*Service) {}, false},
		{"rename failure", func(s *Service) { s.Rename = func(string, string) error { return fmt.Errorf("permission denied") } }, true},
		{"unsupported platform", func(s *Service) { s.GOARCH = "386" }, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			server, _ := releaseServer(t, "v1.1.0", []byte("#!/bin/sh\nprintf 'memory-bank-cli v1.1.0\\n'\n"), test.valid)
			defer server.Close()
			destination := writeDestination(t, []byte("old"))
			var stdout, stderr bytes.Buffer
			s := service(server, destination, &stdout, &stderr)
			test.setup(&s)
			if code := s.Run(context.Background()); code != 1 || !strings.Contains(stderr.String(), "memory-bank-cli update:") {
				t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
			}
			assertBytes(t, destination, []byte("old"))
		})
	}
}

func TestWindowsRequiresManualReplacement(t *testing.T) {
	server, _ := releaseServer(t, "v1.1.0", []byte("ignored"), true)
	defer server.Close()
	destination := writeDestination(t, []byte("old"))
	var stdout, stderr bytes.Buffer
	s := service(server, destination, &stdout, &stderr)
	s.GOOS = "windows"
	if code := s.Run(context.Background()); code != 1 || !strings.Contains(stderr.String(), "manual") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	assertBytes(t, destination, []byte("old"))
}

type counters struct{ asset, sums int }

func releaseServer(t *testing.T, tag string, binary []byte, validChecksum bool) (*httptest.Server, *counters) {
	t.Helper()
	calls := &counters{}
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/latest":
			fmt.Fprintf(writer, `{"tag_name":%q,"assets":[{"name":%q,"browser_download_url":%q},{"name":"checksums.txt","browser_download_url":%q}]}`,
				tag, testAsset, server.URL+"/asset", server.URL+"/checksums")
		case "/asset":
			calls.asset++
			_, _ = writer.Write(binary)
		case "/checksums":
			calls.sums++
			sum := sha256.Sum256(binary)
			if validChecksum {
				fmt.Fprintf(writer, "%x  %s\n", sum, testAsset)
			} else {
				fmt.Fprintf(writer, "%s  %s\n", strings.Repeat("0", 64), testAsset)
			}
		default:
			http.NotFound(writer, request)
		}
	}))
	return server, calls
}

func service(server *httptest.Server, destination string, stdout, stderr *bytes.Buffer) Service {
	return Service{Version: "v1.0.0", GOOS: "linux", GOARCH: "amd64", Client: server.Client(), LatestURL: server.URL + "/latest", Executable: func() (string, error) { return destination, nil }, Stdout: stdout, Stderr: stderr}
}

func writeDestination(t *testing.T, contents []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "memory-bank-cli")
	if err := os.WriteFile(path, contents, 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func assertBytes(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(got, want) {
		t.Fatalf("read %s = %q, %v; want %q", path, got, err, want)
	}
}
