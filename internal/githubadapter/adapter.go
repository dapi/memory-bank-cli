// Package githubadapter installs the optional GitHub workflow surface.
package githubadapter

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	Create   = "create"
	Update   = "update"
	Preserve = "preserve"
	Conflict = "conflict"
)

type Decision struct {
	Path   string `json:"path"`
	Action string `json:"action"`
	Reason string `json:"reason"`
}

type Report struct {
	FormatVersion int        `json:"format_version"`
	DryRun        bool       `json:"dry_run"`
	Applied       bool       `json:"applied"`
	Decisions     []Decision `json:"decisions"`
	ConflictCount int        `json:"conflict_count"`
}

type Options struct {
	RepoRoot string
	DryRun   bool
}

type asset struct{ id, path, content string }

// Run plans and, unless dry-run or a conflict is present, applies the adapter.
// `init` and `update` have identical ownership semantics: markers identify the
// adapter's portion, while unmarked files are always user-owned.
func Run(options Options) (Report, error) {
	if options.RepoRoot == "" {
		return Report{}, fmt.Errorf("repo root is required")
	}
	assets := defaultAssets()
	report := Report{FormatVersion: 1, DryRun: options.DryRun}
	type mutation struct{ path, data string }
	mutations := make([]mutation, 0, len(assets))
	for _, item := range assets {
		path, err := safePath(options.RepoRoot, item.path)
		if err != nil {
			return Report{}, err
		}
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			report.Decisions = append(report.Decisions, Decision{Path: item.path, Action: Create, Reason: "adapter-managed file is missing"})
			mutations = append(mutations, mutation{path: path, data: render(item)})
			continue
		}
		if err != nil {
			return Report{}, fmt.Errorf("read %s: %w", item.path, err)
		}
		next, action, reason := reconcile(item, string(data))
		report.Decisions = append(report.Decisions, Decision{Path: item.path, Action: action, Reason: reason})
		if action == Conflict {
			report.ConflictCount++
		} else if action == Update {
			mutations = append(mutations, mutation{path: path, data: next})
		}
	}
	sort.Slice(report.Decisions, func(i, j int) bool { return report.Decisions[i].Path < report.Decisions[j].Path })
	if report.DryRun || report.ConflictCount > 0 {
		return report, nil
	}
	for _, mutation := range mutations {
		if err := os.MkdirAll(filepath.Dir(mutation.path), 0o755); err != nil {
			return Report{}, err
		}
		if err := os.WriteFile(mutation.path, []byte(mutation.data), 0o644); err != nil {
			return Report{}, err
		}
	}
	report.Applied = len(mutations) > 0
	return report, nil
}

func safePath(root, relative string) (string, error) {
	if filepath.IsAbs(relative) || strings.Contains(relative, "..") || strings.Contains(relative, "\\") {
		return "", fmt.Errorf("unsafe adapter path %q", relative)
	}
	current := root
	for _, part := range strings.Split(filepath.FromSlash(relative), string(filepath.Separator)) {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			return filepath.Join(root, filepath.FromSlash(relative)), nil
		}
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("unsafe adapter path %q: symlink component", relative)
		}
	}
	return current, nil
}

func render(item asset) string {
	digest := digest(item.content)
	return fmt.Sprintf("<!-- MB-CLI GITHUB ADAPTER START: %s sha256:%s -->\n%s<!-- MB-CLI GITHUB ADAPTER END: %s -->\n", item.id, digest, item.content, item.id)
}

func reconcile(item asset, existing string) (string, string, string) {
	startPrefix := "<!-- MB-CLI GITHUB ADAPTER START: " + item.id + " sha256:"
	end := "<!-- MB-CLI GITHUB ADAPTER END: " + item.id + " -->"
	start := strings.Index(existing, startPrefix)
	if start < 0 {
		if strings.Contains(existing, "MB-CLI GITHUB ADAPTER") {
			return existing, Conflict, "adapter markers are malformed or belong to another asset"
		}
		return existing, Preserve, "existing unmarked GitHub file is user-owned"
	}
	lineEnd := strings.Index(existing[start:], " -->\n")
	endAt := strings.Index(existing[start:], end)
	if lineEnd < 0 || endAt < 0 || strings.Count(existing, startPrefix) != 1 || strings.Count(existing, end) != 1 {
		return existing, Conflict, "adapter markers are malformed or ambiguous"
	}
	lineEnd += start + len(" -->\n")
	endStart := start + endAt
	recorded := strings.TrimSuffix(strings.TrimPrefix(existing[start:lineEnd], startPrefix), " -->\n")
	body := existing[lineEnd:endStart]
	if recorded != digest(body) {
		return existing, Conflict, "managed adapter block has downstream drift"
	}
	nextBlock := render(item)
	if existing[start:endStart+len(end)] == strings.TrimSuffix(nextBlock, "\n") {
		return existing, Preserve, "managed adapter block is current"
	}
	return existing[:start] + strings.TrimSuffix(nextBlock, "\n") + existing[endStart+len(end):], Update, "update clean managed adapter block"
}

func digest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func defaultAssets() []asset {
	return []asset{
		{"small-change", ".github/ISSUE_TEMPLATE/memory-bank-small-change.yml", issueForm("Small Change", "small-change")},
		{"feature", ".github/ISSUE_TEMPLATE/memory-bank-feature.yml", issueForm("Feature", "feature")},
		{"pull-request", ".github/pull_request_template.md", prTemplate},
		{"validation", ".github/memory-bank-validation.md", validationGuidance},
		{"agent-guidance", ".github/memory-bank-agent-guidance.md", agentGuidance},
	}
}

func issueForm(name, flow string) string {
	return fmt.Sprintf("name: Memory Bank %s\ndescription: Request work through the Memory Bank %s flow.\ntitle: \"[%s] \"\nbody:\n  - type: textarea\n    id: outcome\n    attributes:\n      label: Expected outcome\n    validations:\n      required: true\n  - type: input\n    id: owner_docs\n    attributes:\n      label: Canonical owner documents\n      description: Link brief, design, plan or ADR when applicable.\n  - type: input\n    id: validation_profile\n    attributes:\n      label: Validation profile\n  - type: textarea\n    id: verify\n    attributes:\n      label: Acceptance and verification evidence\n    validations:\n      required: true\n  - type: input\n    id: feature_or_epic\n    attributes:\n      label: Feature or epic identifier\n", name, flow, flow)
}

const prTemplate = `## What changed

## Canonical issue and owner documents

## Validation evidence

## Risks, manual steps, and known gaps

## Closure

- Use a closing keyword only when the issue's acceptance/evidence and terminal flow state are complete.
- Use a non-closing reference for partial delivery; do not silently change owner-document lifecycle status.
`

const validationGuidance = `# Memory Bank validation configuration

Configure this repository's CI to run a pinned ` + "`mb-cli doctor`" + ` command. This adapter does not install a moving CLI reference or create GitHub state outside this repository. Record the pinned version and CI evidence in the relevant issue or feature verify contract.
`

const agentGuidance = `# GitHub delivery guidance

GitHub Issues own delivery request and workflow state. Memory Bank owner documents remain canonical for problem space (` + "`brief.md`" + `), selected solution (` + "`design.md`" + ` when required), execution (` + "`implementation-plan.md`" + `), and architecture decisions (ADR).

For Small Change, record routing, validation profile, and concrete verification in the issue. For Feature, link the feature package and use its acceptance/evidence contract. A PR alone is not closure: close only after required evidence and terminal flow state; use a non-closing reference for partial delivery.
`
