package main

import (
	"runtime/debug"
	"testing"
)

func TestResolvedVersion(t *testing.T) {
	tests := []struct {
		name          string
		linkerVersion string
		buildVersion  string
		want          string
	}{
		{name: "GoReleaser linker version wins", linkerVersion: "1.2.3", buildVersion: "v9.9.9", want: "1.2.3"},
		{name: "Go install module version", linkerVersion: "dev", buildVersion: "v1.2.3", want: "v1.2.3"},
		{name: "pseudo-version install", linkerVersion: "dev", buildVersion: "v0.0.0-20260722120000-abcdef123456", want: "v0.0.0-20260722120000-abcdef123456"},
		{name: "local build", linkerVersion: "dev", buildVersion: "(devel)", want: "dev"},
		{name: "missing build info", linkerVersion: "dev", want: "dev"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var info *debug.BuildInfo
			if test.buildVersion != "" {
				info = &debug.BuildInfo{Main: debug.Module{Version: test.buildVersion}}
			}
			if got := resolvedVersion(test.linkerVersion, info); got != test.want {
				t.Fatalf("resolvedVersion(%q, %q) = %q, want %q", test.linkerVersion, test.buildVersion, got, test.want)
			}
		})
	}
}
