// Package selfupdate discovers and safely installs compatible memory-bank-cli releases.
package selfupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const latestReleaseURL = "https://api.github.com/repos/dapi/memory-bank-cli/releases/latest"

// Service implements the self-update command. Dependencies are fields so the
// release and filesystem failure paths can be deterministic in tests.
type Service struct {
	Version    string
	GOOS       string
	GOARCH     string
	Executable func() (string, error)
	Client     *http.Client
	LatestURL  string
	Stdout     io.Writer
	Stderr     io.Writer
	Rename     func(string, string) error
}

type release struct {
	TagName    string  `json:"tag_name"`
	Prerelease bool    `json:"prerelease"`
	Assets     []asset `json:"assets"`
}

type asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type semver struct{ major, minor, patch int }

func (v semver) String() string { return fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.patch) }

func (v semver) compare(other semver) int {
	if v.major != other.major {
		if v.major < other.major {
			return -1
		}
		return 1
	}
	if v.minor != other.minor {
		if v.minor < other.minor {
			return -1
		}
		return 1
	}
	if v.patch < other.patch {
		return -1
	}
	if v.patch > other.patch {
		return 1
	}
	return 0
}

// Run returns 0 after a successful update or an already-current result, and
// 1 for every operational failure. It changes the executable only at rename.
func (s Service) Run(ctx context.Context) int {
	out, errOut := s.Stdout, s.Stderr
	if out == nil {
		out = os.Stdout
	}
	if errOut == nil {
		errOut = os.Stderr
	}
	fail := func(format string, args ...any) int {
		fmt.Fprintf(errOut, "memory-bank-cli update: "+format+"\n", args...)
		return 1
	}

	current, err := parseVersion(s.Version)
	if err != nil {
		return fail("running version: %v", err)
	}
	osName, arch, err := target(s.GOOS, s.GOARCH)
	if err != nil {
		return fail("%v", err)
	}
	executable := s.Executable
	if executable == nil {
		executable = os.Executable
	}
	destination, err := executable()
	if err != nil {
		return fail("running executable: %v", err)
	}
	if resolved, resolveErr := filepath.EvalSymlinks(destination); resolveErr == nil {
		destination = resolved
	}

	rel, err := s.latest(ctx)
	if err != nil {
		return fail("latest release: %v", err)
	}
	if rel.Prerelease {
		return fail("latest release is a prerelease")
	}
	targetVersion, err := parseVersion(rel.TagName)
	if err != nil {
		return fail("latest release version: %v", err)
	}
	if targetVersion.compare(current) <= 0 {
		fmt.Fprintln(out, "memory-bank-cli is already up to date.")
		return 0
	}

	assetName := fmt.Sprintf("memory-bank-cli-%s-%s", osName, arch)
	if osName == "windows" {
		assetName += ".exe"
	}
	assetURL, sumsURL, ok := releaseAssets(rel.Assets, assetName)
	if !ok {
		return fail("release %s is missing %s or checksums.txt", rel.TagName, assetName)
	}
	if err := s.install(ctx, assetURL, sumsURL, assetName, destination, rel.TagName); err != nil {
		return fail("install %s: %v", rel.TagName, err)
	}
	fmt.Fprintf(out, "Updated memory-bank-cli to %s at %s\n", rel.TagName, destination)
	return 0
}

func (s Service) latest(ctx context.Context) (release, error) {
	var rel release
	url := s.LatestURL
	if url == "" {
		url = latestReleaseURL
	}
	if err := s.getJSON(ctx, url, &rel); err != nil {
		return rel, err
	}
	return rel, nil
}

func (s Service) getJSON(ctx context.Context, targetURL string, value any) error {
	response, err := s.get(ctx, targetURL)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: HTTP %s", targetURL, response.Status)
	}
	return json.NewDecoder(response.Body).Decode(value)
}

func (s Service) get(ctx context.Context, targetURL string) (*http.Response, error) {
	client := s.Client
	if client == nil {
		client = http.DefaultClient
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "memory-bank-cli-self-update")
	return client.Do(request)
}

func (s Service) install(ctx context.Context, assetURL, sumsURL, assetName, destination, expectedTag string) error {
	staged, err := os.CreateTemp(filepath.Dir(destination), ".memory-bank-cli-update-")
	if err != nil {
		return fmt.Errorf("stage update: %w", err)
	}
	stagedPath := staged.Name()
	defer os.Remove(stagedPath)
	if err := staged.Close(); err != nil {
		return err
	}
	if err := s.download(ctx, assetURL, stagedPath); err != nil {
		return fmt.Errorf("download asset: %w", err)
	}
	checksums, err := s.downloadBytes(ctx, sumsURL)
	if err != nil {
		return fmt.Errorf("download checksums.txt: %w", err)
	}
	expected, err := checksumFor(string(checksums), assetName)
	if err != nil {
		return err
	}
	if err := verifyFile(stagedPath, expected); err != nil {
		return err
	}
	if err := os.Chmod(stagedPath, 0o755); err != nil {
		return fmt.Errorf("make staged executable: %w", err)
	}
	if err := verifyStagedVersion(stagedPath, expectedTag); err != nil {
		return err
	}
	rename := s.Rename
	if rename == nil {
		rename = os.Rename
	}
	if err := rename(stagedPath, destination); err != nil {
		return fmt.Errorf("replace executable: %w", err)
	}
	return nil
}

func (s Service) download(ctx context.Context, source, destination string) error {
	data, err := s.downloadBytes(ctx, source)
	if err != nil {
		return err
	}
	return os.WriteFile(destination, data, 0o600)
}

func (s Service) downloadBytes(ctx context.Context, source string) ([]byte, error) {
	response, err := s.get(ctx, source)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: HTTP %s", source, response.Status)
	}
	return io.ReadAll(response.Body)
}

func target(osName, arch string) (string, string, error) {
	if osName == "" {
		osName = runtime.GOOS
	}
	if arch == "" {
		arch = runtime.GOARCH
	}
	switch osName + ":" + arch {
	case "darwin:amd64", "darwin:arm64", "linux:amd64", "linux:arm64":
		return osName, arch, nil
	case "windows:amd64":
		return "", "", fmt.Errorf("automatic replacement is unsupported on windows; download memory-bank-cli-windows-amd64.exe from the latest release and replace the executable manually")
	default:
		return "", "", fmt.Errorf("unsupported platform %s/%s", osName, arch)
	}
}

func parseVersion(value string) (semver, error) {
	value = strings.TrimPrefix(strings.TrimSpace(value), "v")
	var version semver
	var extra string
	if _, err := fmt.Sscanf(value, "%d.%d.%d%s", &version.major, &version.minor, &version.patch, &extra); err == nil || err != io.EOF || extra != "" || version.major < 0 || version.minor < 0 || version.patch < 0 || version.String() != value {
		return semver{}, fmt.Errorf("invalid semantic version %q", value)
	}
	return version, nil
}

func releaseAssets(assets []asset, assetName string) (string, string, bool) {
	var assetURL, sumsURL string
	for _, asset := range assets {
		if asset.Name == assetName {
			assetURL = asset.BrowserDownloadURL
		}
		if asset.Name == "checksums.txt" {
			sumsURL = asset.BrowserDownloadURL
		}
	}
	return assetURL, sumsURL, assetURL != "" && sumsURL != ""
}

func checksumFor(checksums, assetName string) (string, error) {
	var found string
	for _, line := range strings.Split(checksums, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || strings.TrimPrefix(fields[1], "*") != assetName {
			continue
		}
		if found != "" || len(fields[0]) != sha256.Size*2 {
			return "", fmt.Errorf("invalid checksum entry for %s", assetName)
		}
		if _, err := hex.DecodeString(fields[0]); err != nil {
			return "", fmt.Errorf("invalid checksum entry for %s", assetName)
		}
		found = strings.ToLower(fields[0])
	}
	if found == "" {
		return "", fmt.Errorf("missing checksum for %s", assetName)
	}
	return found, nil
}

func verifyFile(path, expected string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}
	if actual := hex.EncodeToString(hash.Sum(nil)); actual != expected {
		return fmt.Errorf("checksum mismatch for %s", filepath.Base(path))
	}
	return nil
}

func verifyStagedVersion(stagedPath, expectedTag string) error {
	output, err := exec.Command(stagedPath, "--version").Output()
	if err != nil {
		return fmt.Errorf("run staged --version: %w", err)
	}
	want := "memory-bank-cli " + expectedTag
	if got := strings.TrimSpace(string(output)); got != want {
		return fmt.Errorf("staged --version = %q, want %q", got, want)
	}
	return nil
}
