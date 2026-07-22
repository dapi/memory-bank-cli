package main

import (
	"os"
	"runtime/debug"

	"github.com/dapi/memory-bank-cli/internal/cli"
)

var version = "dev"

func main() {
	os.Exit(cli.Run(os.Args[1:], resolvedVersion(version, readBuildInfo()), os.Stdout, os.Stderr))
}

func readBuildInfo() *debug.BuildInfo {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return nil
	}
	return info
}

func resolvedVersion(linkerVersion string, info *debug.BuildInfo) string {
	if linkerVersion != "dev" {
		return linkerVersion
	}
	if info == nil || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return linkerVersion
	}
	return info.Main.Version
}
