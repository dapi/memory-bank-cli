// Package handoff builds a deterministic, read-only projection of explicit
// project evidence for resuming one concrete task.
package handoff

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const FormatVersion = 1

type Options struct {
	RepoRoot    string
	From        string
	GitRange    string
	TestReports []string
	Out         string
}

type SourceRef struct {
	Kind string `json:"kind"`
	Ref  string `json:"ref"`
	Line int    `json:"line,omitempty"`
}

type Item struct {
	Type    string    `json:"type"`
	Context string    `json:"context"`
	Value   string    `json:"value"`
	Source  SourceRef `json:"source"`
}

type Document struct {
	Path   string    `json:"path"`
	Role   string    `json:"role"`
	Source SourceRef `json:"source"`
}

type Commit struct {
	Hash         string        `json:"hash"`
	AuthorDate   string        `json:"author_date"`
	Subject      string        `json:"subject"`
	Source       SourceRef     `json:"source"`
	ChangedFiles []ChangedFile `json:"changed_files"`
}

type ChangedFile struct {
	Path   string    `json:"path"`
	Status string    `json:"status"`
	Source SourceRef `json:"source"`
}

type Verification struct {
	Path   string    `json:"path"`
	Format string    `json:"format"`
	Result any       `json:"result"`
	Source SourceRef `json:"source"`
}

type DeclaredPriming struct {
	Documents     []Document `json:"documents"`
	Dependencies  []Document `json:"dependencies"`
	Decisions     []Item     `json:"decisions"`
	OpenQuestions []Item     `json:"open_questions"`
	Blockers      []Item     `json:"blockers"`
	NextSteps     []Item     `json:"next_steps"`
}

type ObservedExecution struct {
	GitRange     string         `json:"git_range,omitempty"`
	Commits      []Commit       `json:"commits"`
	ChangedFiles []ChangedFile  `json:"changed_files"`
	Verification []Verification `json:"verification"`
}

type Unresolved struct {
	Kind      string    `json:"kind"`
	Reference string    `json:"reference"`
	Source    SourceRef `json:"source"`
	Reason    string    `json:"reason"`
}

type Report struct {
	FormatVersion     int               `json:"format_version"`
	ReadOnly          bool              `json:"read_only"`
	StartingDocument  string            `json:"starting_document"`
	DeclaredPriming   DeclaredPriming   `json:"declared_priming"`
	ObservedExecution ObservedExecution `json:"observed_execution"`
	Items             []Item            `json:"items"`
	Unresolved        []Unresolved      `json:"unresolved"`
}

var markdownLinkRE = regexp.MustCompile(`(!)?(?:\[[^\]]*\])\(([^)]+)\)`)
var headingRE = regexp.MustCompile(`^#{1,6}\s+(.+?)\s*$`)

func Build(options Options) (Report, error) {
	report := emptyReport()
	root, err := filepath.Abs(options.RepoRoot)
	if err != nil {
		return report, fmt.Errorf("resolve repository root: %w", err)
	}
	if options.From == "" {
		addUnresolved(&report, "input", "--from", SourceRef{Kind: "input", Ref: "--from"}, "required input is missing")
		return report, errors.New("handoff build: --from is required")
	}

	fromPath, fromAbsolute, err := resolveInputPath(root, options.From, true)
	if err != nil {
		addUnresolved(&report, "path", options.From, SourceRef{Kind: "input", Ref: "--from"}, err.Error())
		return report, buildError(&report)
	}
	report.StartingDocument = fromPath

	if options.Out != "" {
		if outputPath, outputErr := resolveOutputPath(root, options.Out); outputErr == nil {
			if samePath(outputPath, fromAbsolute) {
				addUnresolved(&report, "output", options.Out, SourceRef{Kind: "input", Ref: "--out"}, "output must not overwrite the starting document")
				return report, buildError(&report)
			}
		} else {
			addUnresolved(&report, "output", options.Out, SourceRef{Kind: "input", Ref: "--out"}, outputErr.Error())
			return report, buildError(&report)
		}
	}

	documents, dependencies, docItems, unresolved := collectDocuments(root, fromPath, fromAbsolute)
	report.DeclaredPriming.Documents = documents
	report.DeclaredPriming.Dependencies = dependencies
	report.Unresolved = append(report.Unresolved, unresolved...)
	report.DeclaredPriming.Decisions = docItems.decisions
	report.DeclaredPriming.OpenQuestions = docItems.openQuestions
	report.DeclaredPriming.Blockers = docItems.blockers
	report.DeclaredPriming.NextSteps = docItems.nextSteps
	if options.Out != "" {
		outputPath, outputErr := resolveOutputPath(root, options.Out)
		if outputErr != nil {
			addUnresolved(&report, "output", options.Out, SourceRef{Kind: "input", Ref: "--out"}, outputErr.Error())
		} else {
			for _, document := range report.DeclaredPriming.Documents {
				if samePath(outputPath, filepath.Join(root, filepath.FromSlash(document.Path))) {
					addUnresolved(&report, "output", options.Out, SourceRef{Kind: "input", Ref: "--out"}, "output must not overwrite a source document")
					break
				}
			}
		}
	}

	if options.GitRange != "" {
		report.ObservedExecution.GitRange = options.GitRange
		commits, changedFiles, gitErr := collectGitEvidence(root, options.GitRange)
		if gitErr != nil {
			addUnresolved(&report, "git-range", options.GitRange, SourceRef{Kind: "input", Ref: "--git-range"}, gitErr.Error())
		} else {
			report.ObservedExecution.Commits = commits
			report.ObservedExecution.ChangedFiles = changedFiles
		}
	}

	for _, reportPath := range options.TestReports {
		verification, reportErr := readVerificationReport(root, reportPath)
		if reportErr != nil {
			addUnresolved(&report, "report", reportPath, SourceRef{Kind: "input", Ref: "--test-report"}, reportErr.Error())
			continue
		}
		if options.Out != "" {
			outputPath, _ := resolveOutputPath(root, options.Out)
			if samePath(outputPath, filepath.Join(root, filepath.FromSlash(verification.Path))) {
				addUnresolved(&report, "output", options.Out, SourceRef{Kind: "input", Ref: "--out"}, "output must not overwrite a verification report")
				continue
			}
		}
		report.ObservedExecution.Verification = append(report.ObservedExecution.Verification, verification)
	}

	finalize(&report)
	if len(report.Unresolved) > 0 {
		return report, buildError(&report)
	}
	return report, nil
}

func emptyReport() Report {
	return Report{
		FormatVersion: FormatVersion,
		ReadOnly:      true,
		DeclaredPriming: DeclaredPriming{
			Documents: []Document{}, Dependencies: []Document{}, Decisions: []Item{},
			OpenQuestions: []Item{}, Blockers: []Item{}, NextSteps: []Item{},
		},
		ObservedExecution: ObservedExecution{Commits: []Commit{}, ChangedFiles: []ChangedFile{}, Verification: []Verification{}},
		Items:             []Item{},
		Unresolved:        []Unresolved{},
	}
}

func buildError(report *Report) error {
	if len(report.Unresolved) == 0 {
		return errors.New("handoff build failed")
	}
	parts := make([]string, 0, len(report.Unresolved))
	for _, unresolved := range report.Unresolved {
		parts = append(parts, fmt.Sprintf("%s %q: %s", unresolved.Kind, unresolved.Reference, unresolved.Reason))
	}
	return fmt.Errorf("handoff build failed: unresolved evidence: %s", strings.Join(parts, "; "))
}

func resolveInputPath(root, value string, allowMemoryBankFallback bool) (string, string, error) {
	absolute, err := absoluteWithinRoot(root, value)
	if err != nil {
		return "", "", err
	}
	if allowMemoryBankFallback {
		if _, statErr := os.Stat(absolute); os.IsNotExist(statErr) && !filepath.IsAbs(value) {
			fallback := filepath.Join(root, "memory-bank", filepath.FromSlash(value))
			if _, fallbackErr := os.Stat(fallback); fallbackErr == nil {
				absolute = fallback
			}
		}
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return "", "", fmt.Errorf("source does not exist: %w", err)
	}
	if !info.Mode().IsRegular() {
		return "", "", errors.New("source is not a regular file")
	}
	relative, err := filepath.Rel(root, absolute)
	if err != nil {
		return "", "", err
	}
	return filepath.ToSlash(relative), absolute, nil
}

func absoluteWithinRoot(root, value string) (string, error) {
	absolute := value
	if !filepath.IsAbs(absolute) {
		absolute = filepath.Join(root, filepath.FromSlash(value))
	}
	abs, err := filepath.Abs(absolute)
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(root, abs)
	if err != nil {
		return "", err
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", errors.New("path escapes repository root")
	}
	return abs, nil
}

func resolveOutputPath(root, value string) (string, error) {
	if filepath.IsAbs(value) {
		absolute, err := filepath.Abs(value)
		if err != nil {
			return "", err
		}
		return absolute, nil
	}
	return filepath.Abs(filepath.Join(root, filepath.FromSlash(value)))
}

func samePath(left, right string) bool {
	leftAbs, leftErr := filepath.Abs(left)
	rightAbs, rightErr := filepath.Abs(right)
	if leftErr != nil || rightErr != nil {
		return false
	}
	leftAbs = evaluatedPath(leftAbs)
	rightAbs = evaluatedPath(rightAbs)
	return filepath.Clean(leftAbs) == filepath.Clean(rightAbs)
}

func evaluatedPath(path string) string {
	if evaluated, err := filepath.EvalSymlinks(path); err == nil {
		return evaluated
	}
	return path
}

type documentItems struct {
	decisions, openQuestions, blockers, nextSteps []Item
}

func collectDocuments(root, startingPath, startingAbsolute string) ([]Document, []Document, documentItems, []Unresolved) {
	type queuedDocument struct {
		path, absolute, role string
	}
	queue := []queuedDocument{{path: startingPath, absolute: startingAbsolute, role: "starting_document"}}
	seen := map[string]bool{}
	all := make([]Document, 0)
	dependencies := make([]Document, 0)
	items := documentItems{}
	unresolved := make([]Unresolved, 0)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if seen[current.path] {
			continue
		}
		seen[current.path] = true
		document := Document{Path: current.path, Role: current.role, Source: SourceRef{Kind: "document", Ref: current.path}}
		all = append(all, document)
		if current.role != "starting_document" {
			dependencies = append(dependencies, document)
		}

		data, err := os.ReadFile(current.absolute)
		if err != nil {
			unresolved = append(unresolved, Unresolved{Kind: "document", Reference: current.path, Source: SourceRef{Kind: "document", Ref: current.path}, Reason: fmt.Sprintf("read failed: %v", err)})
			continue
		}
		text := string(data)
		frontmatter, frontmatterErr := parseFrontmatter(text)
		if frontmatterErr != nil {
			unresolved = append(unresolved, Unresolved{Kind: "document", Reference: current.path, Source: SourceRef{Kind: "document", Ref: current.path}, Reason: fmt.Sprintf("invalid frontmatter: %v", frontmatterErr)})
		}
		references := collectDocumentReferences(frontmatter, text)
		for _, reference := range references {
			if isExternalReference(reference.Reference) {
				continue
			}
			dependencyPath, dependencyAbsolute, resolveErr := resolveDocumentReference(root, current.path, current.absolute, reference.Reference)
			if resolveErr != nil {
				unresolved = append(unresolved, Unresolved{Kind: "document", Reference: reference.Reference, Source: SourceRef{Kind: "document", Ref: current.path, Line: reference.Line}, Reason: resolveErr.Error()})
				continue
			}
			if !seen[dependencyPath] {
				queue = append(queue, queuedDocument{path: dependencyPath, absolute: dependencyAbsolute, role: "explicit_dependency"})
			}
		}

		parsed := extractEvidenceItems(current.path, text)
		items.decisions = append(items.decisions, parsed.decisions...)
		items.openQuestions = append(items.openQuestions, parsed.openQuestions...)
		items.blockers = append(items.blockers, parsed.blockers...)
		items.nextSteps = append(items.nextSteps, parsed.nextSteps...)
	}

	sort.Slice(all, func(i, j int) bool { return all[i].Path < all[j].Path })
	sort.Slice(dependencies, func(i, j int) bool { return dependencies[i].Path < dependencies[j].Path })
	sortItems(items.decisions)
	sortItems(items.openQuestions)
	sortItems(items.blockers)
	sortItems(items.nextSteps)
	return all, dependencies, items, unresolved
}

type documentReference struct {
	Reference string
	Line      int
}

func parseFrontmatter(text string) (map[string]any, error) {
	if !strings.HasPrefix(text, "---\n") {
		return map[string]any{}, nil
	}
	end := strings.Index(text[4:], "\n---")
	if end < 0 {
		return nil, errors.New("opening delimiter has no closing delimiter")
	}
	content := text[4 : 4+end]
	values := make(map[string]any)
	if err := yaml.Unmarshal([]byte(content), &values); err != nil {
		return nil, err
	}
	return values, nil
}

func collectDocumentReferences(frontmatter map[string]any, text string) []documentReference {
	result := make([]documentReference, 0)
	for _, key := range []string{"derived_from", "source_refs", "dependencies", "depends_on"} {
		result = append(result, valuesAsReferences(frontmatter[key])...)
	}
	lines := strings.Split(text, "\n")
	for lineNumber, line := range lines {
		for _, match := range markdownLinkRE.FindAllStringSubmatch(line, -1) {
			if len(match) == 3 && match[1] == "" {
				result = append(result, documentReference{Reference: strings.TrimSpace(strings.Trim(match[2], "<>")), Line: lineNumber + 1})
			}
		}
	}
	unique := make(map[string]documentReference)
	for _, reference := range result {
		reference.Reference = strings.TrimSpace(strings.SplitN(reference.Reference, "#", 2)[0])
		reference.Reference = strings.TrimSpace(strings.SplitN(reference.Reference, "?", 2)[0])
		if reference.Reference == "" || strings.HasPrefix(reference.Reference, "#") {
			continue
		}
		key := reference.Reference + "\x00" + strconv.Itoa(reference.Line)
		unique[key] = reference
	}
	result = result[:0]
	for _, reference := range unique {
		result = append(result, reference)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Reference == result[j].Reference {
			return result[i].Line < result[j].Line
		}
		return result[i].Reference < result[j].Reference
	})
	return result
}

func valuesAsReferences(value any) []documentReference {
	result := make([]documentReference, 0)
	switch typed := value.(type) {
	case string:
		result = append(result, documentReference{Reference: typed})
	case []any:
		for _, element := range typed {
			result = append(result, valuesAsReferences(element)...)
		}
	case map[string]any:
		for _, key := range []string{"path", "ref", "source"} {
			if reference, ok := typed[key].(string); ok {
				result = append(result, documentReference{Reference: reference})
				break
			}
		}
	case map[any]any:
		for _, key := range []string{"path", "ref", "source"} {
			if reference, ok := typed[key].(string); ok {
				result = append(result, documentReference{Reference: reference})
				break
			}
		}
	}
	return result
}

func isExternalReference(reference string) bool {
	lower := strings.ToLower(reference)
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "mailto:") || strings.HasPrefix(reference, "//")
}

func resolveDocumentReference(root, documentPath, documentAbsolute, reference string) (string, string, error) {
	candidate := reference
	if filepath.IsAbs(candidate) {
		candidate = filepath.Clean(candidate)
	} else if strings.HasPrefix(candidate, "memory-bank/") {
		candidate = filepath.Join(root, filepath.FromSlash(candidate))
	} else {
		candidate = filepath.Join(filepath.Dir(documentAbsolute), filepath.FromSlash(candidate))
	}
	absolute, err := filepath.Abs(candidate)
	if err != nil {
		return "", "", err
	}
	relative, err := filepath.Rel(root, absolute)
	if err != nil {
		return "", "", err
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", "", errors.New("reference escapes repository root")
	}
	info, err := os.Stat(absolute)
	if os.IsNotExist(err) {
		return "", "", errors.New("referenced path does not exist")
	}
	if err != nil {
		return "", "", fmt.Errorf("stat referenced path: %w", err)
	}
	if info.IsDir() {
		readme := filepath.Join(absolute, "README.md")
		readmeInfo, readmeErr := os.Stat(readme)
		if readmeErr != nil || !readmeInfo.Mode().IsRegular() {
			return "", "", errors.New("referenced directory has no regular README.md")
		}
		absolute = readme
		relative, _ = filepath.Rel(root, absolute)
	}
	if !info.Mode().IsRegular() && !info.IsDir() {
		return "", "", errors.New("referenced path is not a regular file")
	}
	_ = documentPath
	return filepath.ToSlash(relative), absolute, nil
}

type extractedItems struct {
	decisions, openQuestions, blockers, nextSteps []Item
}

func extractEvidenceItems(path, text string) extractedItems {
	result := extractedItems{}
	section := ""
	inFence := false
	for lineNumber, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			continue
		}
		if !inFence {
			if match := headingRE.FindStringSubmatch(line); len(match) == 2 {
				section = evidenceSection(match[1])
				continue
			}
		}
		if section == "" || inFence || trimmed == "" || strings.HasPrefix(trimmed, "<!--") {
			continue
		}
		if strings.HasPrefix(trimmed, "|") && strings.Trim(trimmed, "| -:") == "" {
			continue
		}
		value := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
		value = strings.TrimSpace(strings.TrimPrefix(value, "* "))
		if value == "" {
			continue
		}
		item := Item{Type: section, Context: "declared_priming", Value: value, Source: SourceRef{Kind: "document", Ref: path, Line: lineNumber + 1}}
		switch section {
		case "decision":
			result.decisions = append(result.decisions, item)
		case "open_question":
			result.openQuestions = append(result.openQuestions, item)
		case "blocker":
			result.blockers = append(result.blockers, item)
		case "next_step":
			result.nextSteps = append(result.nextSteps, item)
		}
	}
	return result
}

func evidenceSection(heading string) string {
	normalized := strings.ToLower(strings.TrimSpace(heading))
	for _, marker := range []struct{ name, value string }{
		{"open question", "open_question"}, {"open questions", "open_question"},
		{"decision", "decision"}, {"decisions", "decision"},
		{"blocker", "blocker"}, {"blockers", "blocker"},
		{"next step", "next_step"}, {"next steps", "next_step"},
	} {
		if strings.Contains(normalized, marker.name) {
			return marker.value
		}
	}
	return ""
}

func collectGitEvidence(root, gitRange string) ([]Commit, []ChangedFile, error) {
	output, err := runGit(root, "log", "--reverse", "--topo-order", "--format=%H%x00%aI%x00%s", gitRange)
	if err != nil {
		return nil, nil, fmt.Errorf("read git range %q: %s", gitRange, strings.TrimSpace(err.Error()))
	}
	commits := make([]Commit, 0)
	changedFiles := make([]ChangedFile, 0)
	fields := bytes.Split(output, []byte{0})
	for index := 0; index+2 < len(fields); index += 3 {
		if len(fields[index]) == 0 {
			continue
		}
		hash := string(fields[index])
		commit := Commit{Hash: hash, AuthorDate: string(fields[index+1]), Subject: strings.TrimSuffix(string(fields[index+2]), "\n"), Source: SourceRef{Kind: "commit", Ref: hash}, ChangedFiles: []ChangedFile{}}
		files, fileErr := collectCommitFiles(root, hash)
		if fileErr != nil {
			return nil, nil, fileErr
		}
		commit.ChangedFiles = files
		commits = append(commits, commit)
		changedFiles = append(changedFiles, files...)
	}
	return commits, changedFiles, nil
}

func collectCommitFiles(root, hash string) ([]ChangedFile, error) {
	output, err := runGit(root, "diff-tree", "--root", "--no-commit-id", "--name-status", "-z", "--find-renames", hash)
	if err != nil {
		return nil, fmt.Errorf("read changed files for commit %s: %s", hash, strings.TrimSpace(err.Error()))
	}
	fields := bytes.Split(output, []byte{0})
	files := make([]ChangedFile, 0)
	for index := 0; index < len(fields); {
		if len(fields[index]) == 0 {
			index++
			continue
		}
		status := string(fields[index])
		index++
		if index >= len(fields) {
			break
		}
		path := string(fields[index])
		index++
		if strings.HasPrefix(status, "R") || strings.HasPrefix(status, "C") {
			if index >= len(fields) {
				break
			}
			path = path + " -> " + string(fields[index])
			index++
		}
		files = append(files, ChangedFile{Path: path, Status: status, Source: SourceRef{Kind: "commit", Ref: hash}})
	}
	sort.Slice(files, func(i, j int) bool {
		if files[i].Path == files[j].Path {
			return files[i].Status < files[j].Status
		}
		return files[i].Path < files[j].Path
	})
	return files, nil
}

func runGit(root string, arguments ...string) ([]byte, error) {
	command := exec.Command("git", append([]string{"-C", root}, arguments...)...)
	output, err := command.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, strings.TrimSpace(string(output)))
	}
	return output, nil
}

func readVerificationReport(root, reportPath string) (Verification, error) {
	relative, absolute, err := resolveInputPath(root, reportPath, false)
	if err != nil {
		return Verification{}, err
	}
	data, err := os.ReadFile(absolute)
	if err != nil {
		return Verification{}, fmt.Errorf("read report: %w", err)
	}
	verification := Verification{Path: relative, Source: SourceRef{Kind: "report", Ref: relative}}
	if strings.EqualFold(filepath.Ext(relative), ".json") {
		var result any
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.UseNumber()
		if err := decoder.Decode(&result); err != nil {
			return Verification{}, fmt.Errorf("invalid JSON report: %w", err)
		}
		var trailing any
		if err := decoder.Decode(&trailing); err == nil {
			return Verification{}, errors.New("invalid JSON report: more than one JSON value")
		} else if !errors.Is(err, io.EOF) {
			return Verification{}, fmt.Errorf("invalid JSON report: %w", err)
		}
		verification.Format = "json"
		verification.Result = result
		return verification, nil
	}
	verification.Format = "text"
	verification.Result = strings.TrimRight(string(data), "\n")
	return verification, nil
}

func finalize(report *Report) {
	sort.Slice(report.DeclaredPriming.Documents, func(i, j int) bool {
		return report.DeclaredPriming.Documents[i].Path < report.DeclaredPriming.Documents[j].Path
	})
	sort.Slice(report.DeclaredPriming.Dependencies, func(i, j int) bool {
		return report.DeclaredPriming.Dependencies[i].Path < report.DeclaredPriming.Dependencies[j].Path
	})
	sortItems(report.DeclaredPriming.Decisions)
	sortItems(report.DeclaredPriming.OpenQuestions)
	sortItems(report.DeclaredPriming.Blockers)
	sortItems(report.DeclaredPriming.NextSteps)
	sort.Slice(report.ObservedExecution.Verification, func(i, j int) bool {
		return report.ObservedExecution.Verification[i].Path < report.ObservedExecution.Verification[j].Path
	})
	sort.Slice(report.Unresolved, func(i, j int) bool {
		if report.Unresolved[i].Source.Ref == report.Unresolved[j].Source.Ref {
			return report.Unresolved[i].Reference < report.Unresolved[j].Reference
		}
		return report.Unresolved[i].Source.Ref < report.Unresolved[j].Source.Ref
	})
	report.Items = report.Items[:0]
	for _, document := range report.DeclaredPriming.Documents {
		report.Items = append(report.Items, Item{Type: "document", Context: "declared_priming", Value: document.Path, Source: document.Source})
	}
	for _, dependency := range report.DeclaredPriming.Dependencies {
		report.Items = append(report.Items, Item{Type: "dependency", Context: "declared_priming", Value: dependency.Path, Source: dependency.Source})
	}
	for _, items := range [][]Item{report.DeclaredPriming.Decisions, report.DeclaredPriming.OpenQuestions, report.DeclaredPriming.Blockers, report.DeclaredPriming.NextSteps} {
		report.Items = append(report.Items, items...)
	}
	for _, commit := range report.ObservedExecution.Commits {
		report.Items = append(report.Items, Item{Type: "commit", Context: "observed_execution", Value: commit.Subject, Source: commit.Source})
		for _, changedFile := range commit.ChangedFiles {
			report.Items = append(report.Items, Item{Type: "changed_file", Context: "observed_execution", Value: changedFile.Path, Source: changedFile.Source})
		}
	}
	for _, verification := range report.ObservedExecution.Verification {
		report.Items = append(report.Items, Item{Type: "verification", Context: "observed_execution", Value: verification.Path, Source: verification.Source})
	}
}

func sortItems(items []Item) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Source.Ref == items[j].Source.Ref {
			if items[i].Source.Line == items[j].Source.Line {
				return items[i].Value < items[j].Value
			}
			return items[i].Source.Line < items[j].Source.Line
		}
		return items[i].Source.Ref < items[j].Source.Ref
	})
}

func addUnresolved(report *Report, kind, reference string, source SourceRef, reason string) {
	report.Unresolved = append(report.Unresolved, Unresolved{Kind: kind, Reference: reference, Source: source, Reason: reason})
	finalize(report)
}

func RenderMarkdown(report Report) string {
	var output strings.Builder
	output.WriteString("# Execution Handoff\n\n")
	fmt.Fprintf(&output, "- Format version: `%d`\n- Read-only: `%t`\n- Starting document: `%s`\n\n", report.FormatVersion, report.ReadOnly, markdownCode(report.StartingDocument))
	output.WriteString("## Declared priming context\n\n")
	writeDocuments(&output, "Documents", report.DeclaredPriming.Documents)
	writeDocuments(&output, "Explicit dependencies", report.DeclaredPriming.Dependencies)
	writeItems(&output, "Decisions", report.DeclaredPriming.Decisions)
	writeItems(&output, "Open questions", report.DeclaredPriming.OpenQuestions)
	writeItems(&output, "Blockers", report.DeclaredPriming.Blockers)
	writeItems(&output, "Next steps", report.DeclaredPriming.NextSteps)

	output.WriteString("## Observed execution context\n\n")
	if report.ObservedExecution.GitRange == "" {
		output.WriteString("Git range: none supplied.\n\n")
	} else {
		fmt.Fprintf(&output, "Git range: `%s`\n\n", markdownCode(report.ObservedExecution.GitRange))
	}
	output.WriteString("### Commits\n\n")
	if len(report.ObservedExecution.Commits) == 0 {
		output.WriteString("- None.\n\n")
	} else {
		for _, commit := range report.ObservedExecution.Commits {
			fmt.Fprintf(&output, "- `%s` %s (source: %s)\n", markdownCode(commit.Hash), commit.Subject, markdownSource(commit.Source))
			if commit.AuthorDate != "" {
				fmt.Fprintf(&output, "  - Author date: `%s`\n", markdownCode(commit.AuthorDate))
			}
			for _, file := range commit.ChangedFiles {
				fmt.Fprintf(&output, "  - Changed `%s` (%s; source: %s)\n", markdownCode(file.Path), markdownCode(file.Status), markdownSource(file.Source))
			}
		}
		output.WriteString("\n")
	}
	output.WriteString("### Changed files\n\n")
	if len(report.ObservedExecution.ChangedFiles) == 0 {
		output.WriteString("- None.\n\n")
	} else {
		for _, file := range report.ObservedExecution.ChangedFiles {
			fmt.Fprintf(&output, "- `%s` (%s; source: %s)\n", markdownCode(file.Path), markdownCode(file.Status), markdownSource(file.Source))
		}
		output.WriteString("\n")
	}
	output.WriteString("### Verification\n\n")
	if len(report.ObservedExecution.Verification) == 0 {
		output.WriteString("- None.\n\n")
	} else {
		for _, verification := range report.ObservedExecution.Verification {
			fmt.Fprintf(&output, "- `%s` (%s; source: %s)\n", markdownCode(verification.Path), markdownCode(verification.Format), markdownSource(verification.Source))
			if verification.Format == "text" {
				fmt.Fprintf(&output, "\n  ```text\n%s\n  ```\n", indentText(fmt.Sprint(verification.Result), "  "))
			} else if verification.Format == "json" {
				if data, err := json.MarshalIndent(verification.Result, "  ", "  "); err == nil {
					fmt.Fprintf(&output, "\n  ```json\n%s\n  ```\n", indentText(string(data), "  "))
				}
			}
		}
		output.WriteString("\n")
	}
	output.WriteString("## Evidence items\n\n")
	if len(report.Items) == 0 {
		output.WriteString("- None.\n\n")
	} else {
		for _, item := range report.Items {
			fmt.Fprintf(&output, "- **%s** (%s): %s (source: %s)\n", markdownCode(item.Type), markdownCode(item.Context), item.Value, markdownSource(item.Source))
		}
		output.WriteString("\n")
	}
	output.WriteString("## Unresolved sources\n\n")
	if len(report.Unresolved) == 0 {
		output.WriteString("- None.\n")
	} else {
		for _, unresolved := range report.Unresolved {
			fmt.Fprintf(&output, "- **%s** `%s`: %s (source: %s)\n", markdownCode(unresolved.Kind), markdownCode(unresolved.Reference), unresolved.Reason, markdownSource(unresolved.Source))
		}
	}
	return output.String()
}

func writeDocuments(output *strings.Builder, title string, documents []Document) {
	fmt.Fprintf(output, "### %s\n\n", title)
	if len(documents) == 0 {
		output.WriteString("- None.\n\n")
		return
	}
	for _, document := range documents {
		fmt.Fprintf(output, "- `%s` (%s; source: %s)\n", markdownCode(document.Path), markdownCode(document.Role), markdownSource(document.Source))
	}
	output.WriteString("\n")
}

func writeItems(output *strings.Builder, title string, items []Item) {
	fmt.Fprintf(output, "### %s\n\n", title)
	if len(items) == 0 {
		output.WriteString("- None.\n\n")
		return
	}
	for _, item := range items {
		fmt.Fprintf(output, "- %s (source: %s)\n", item.Value, markdownSource(item.Source))
	}
	output.WriteString("\n")
}

func markdownSource(source SourceRef) string {
	ref := markdownCode(source.Ref)
	if source.Line > 0 {
		return fmt.Sprintf("%s:%d", ref, source.Line)
	}
	return ref
}

func markdownCode(value string) string {
	return strings.ReplaceAll(value, "`", "'")
}

func indentText(value, prefix string) string {
	lines := strings.Split(value, "\n")
	for index := range lines {
		lines[index] = prefix + lines[index]
	}
	return strings.Join(lines, "\n")
}
