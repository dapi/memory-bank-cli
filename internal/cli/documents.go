package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/dapi/memory-bank-cli/internal/ownership"
	"github.com/dapi/memory-bank-cli/internal/repository"
)

func runDocument(arguments []string, stdout, stderr io.Writer) int {
	if len(arguments) == 0 {
		fmt.Fprintln(stderr, "Usage: memory-bank-cli document <create|adopt|transition|move> [options]")
		return exitUsage
	}
	op := arguments[0]
	switch op {
	case "create", "adopt", "transition", "move":
	default:
		fmt.Fprintln(stderr, "unknown document operation")
		return exitUsage
	}
	flags := flag.NewFlagSet("memory-bank-cli document "+op, flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := addRepoRootFlag(flags)
	o := ownership.DocumentOptions{Operation: op}
	flags.StringVar(&o.From, "from", "", "prepared local draft for atomic creation (repository-relative)")
	flags.StringVar(&o.Type, "type", "", "installed document type")
	flags.StringVar(&o.Path, "path", "", "project Markdown path under memory-bank")
	flags.StringVar(&o.To, "to", "", "move destination within the original context")
	flags.StringVar(&o.ID, "id", "", "required stable identity for move")
	flags.StringVar(&o.Contract, "contract", "", "installed flow contract to adopt")
	flags.BoolVar(&o.LegacyFlow, "legacy-flow", false, "use the installation's pinned compatibility contract")
	flags.BoolVar(&o.DryRun, "dry-run", false, "validate and preview without mutation")
	var evidence entrypointFlags
	flags.Var(&evidence, "evidence", "transition evidence reference (repeatable)")
	jsonOutput := addJSONOutputFlag(flags)
	if err := flags.Parse(arguments[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitSuccess
		}
		return exitUsage
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "unexpected document arguments")
		return exitUsage
	}
	var err error
	o.RepoRoot, err = repository.ResolveRoot(*root)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	o.Evidence = evidence
	report, err := ownership.DocumentOperation(o)
	if err != nil {
		if report.Applied {
			_ = writeResult(stdout, *jsonOutput, report, func(w io.Writer) { printOwnershipReport(w, report) })
		}
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	if err = writeResult(stdout, *jsonOutput, report, func(w io.Writer) { printOwnershipReport(w, report) }); err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	return exitSuccess
}
