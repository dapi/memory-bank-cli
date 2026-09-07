package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/dapi/memory-bank-cli/internal/ownership"
)

func runCapabilities(arguments []string, version string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("memory-bank-cli capabilities", flag.ContinueOnError)
	flags.SetOutput(stderr)
	var required entrypointFlags
	flags.Var(&required, "require", "required capability (repeatable)")
	if err := flags.Parse(arguments); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitSuccess
		}
		return exitUsage
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "capabilities: unexpected positional arguments")
		return exitUsage
	}
	report := struct {
		SchemaVersion int      `json:"schema_version"`
		CLIVersion    string   `json:"cli_version"`
		Capabilities  []string `json:"capabilities"`
		Unsupported   []string `json:"unsupported"`
	}{1, version, ownership.SupportedCapabilities(), []string{}}
	for _, requested := range required {
		found := false
		for _, available := range report.Capabilities {
			if requested == available {
				found = true
			}
		}
		if !found {
			report.Unsupported = append(report.Unsupported, requested)
		}
	}
	if err := json.NewEncoder(stdout).Encode(report); err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	if len(report.Unsupported) != 0 {
		return exitFailure
	}
	return exitSuccess
}
