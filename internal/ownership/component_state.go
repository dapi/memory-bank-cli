package ownership

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/dapi/memory-bank-cli/internal/agentinstructions"
	"github.com/dapi/memory-bank-cli/internal/contracts"
	"github.com/dapi/memory-bank-cli/internal/lint"
)

type observation struct {
	Exists      bool   `json:"exists"`
	Digest      string `json:"digest"`
	Mode        string `json:"mode"`
	Permissions string `json:"permissions"`
}
type directoryState struct {
	BeforeExists bool   `json:"before_exists"`
	BeforeMode   string `json:"before_mode"`
	AfterExists  bool   `json:"after_exists"`
	AfterMode    string `json:"after_mode"`
}
type componentTree struct {
	files    map[string][]byte
	observed map[string]observation
}
type componentState struct {
	manifest contracts.Manifest
	catalog  contracts.Catalog
	registry contracts.Registry
	bindings map[string]contracts.Binding
	docs     map[string]contracts.Document
	findings []contracts.Finding
}

func componentHost() bool { return runtime.GOOS == "darwin" || runtime.GOOS == "linux" }

// Observe through pinned, no-follow handles, including exact permission bits.
func observeComponent(repo pinnedRepo, p string) (observation, []byte, error) {
	if !contracts.ValidPath(p) {
		return observation{}, nil, fmt.Errorf("unsafe component path %s", p)
	}
	if err := checkComponentAncestors(repo, p); err != nil {
		return observation{}, nil, err
	}
	_, info, exists, err := inspectDestination(repo, p)
	if err != nil {
		return observation{}, nil, err
	}
	if !exists {
		return observation{}, nil, nil
	}
	readInfo, data, err := secureReadDestination(repo, p)
	if err != nil {
		return observation{}, nil, err
	}
	if !os.SameFile(info, readInfo) {
		return observation{}, nil, fmt.Errorf("%s changed while reading", p)
	}
	if err := checkOriginalComponentFile(readInfo); err != nil {
		return observation{}, nil, fmt.Errorf("%s: %w", p, err)
	}
	return observation{true, digest(data), gitMode(readInfo.Mode().Perm()), fmt.Sprintf("%04o", readInfo.Mode().Perm())}, data, nil
}

// Check each directory's actual spelling and siblings before resolving a target.
// This catches aliases even on case-insensitive or normalization-insensitive hosts.
func checkComponentAncestors(repo pinnedRepo, p string) error {
	parts := strings.Split(p, "/")
	current := repo.root
	for i, part := range parts {
		entries, err := os.ReadDir(current)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		names := []string{part}
		found := false
		for _, e := range entries {
			if e.Name() == part {
				found = true
				continue
			}
			names = append(names, e.Name())
		}
		// Only compare the target's collision key; unrelated nonportable entries do
		// not prevent work in another subtree.
		key := contracts.PortableKey(part)
		for _, name := range names[1:] {
			if contracts.PortableKey(name) == key {
				return fmt.Errorf("portable path alias: %s / %s", part, name)
			}
		}
		if !found {
			return nil
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink in component path %s", p)
		}
		if i < len(parts)-1 && !info.IsDir() {
			return fmt.Errorf("non-directory ancestor of %s", p)
		}
	}
	return nil
}

func readComponentTree(repo pinnedRepo, lock Lock, extra []string) (componentTree, error) {
	t := componentTree{map[string][]byte{}, map[string]observation{}}
	paths := map[string]bool{LockFileName: true, contracts.RegistryPath: true, "AGENTS.md": true}
	for p := range lock.Files {
		paths[p] = true
	}
	for _, p := range extra {
		if p != "" {
			paths[p] = true
		}
	}
	root := filepath.Join(repo.root, "memory-bank")
	err := filepath.WalkDir(root, func(full string, e fs.DirEntry, err error) error {
		if errors.Is(err, os.ErrNotExist) && full == root {
			return nil
		}
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(repo.root, full)
		p := filepath.ToSlash(rel)
		if e.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink in component tree %s", p)
		}
		if e.IsDir() {
			if p == "memory-bank/.repo" {
				return filepath.SkipDir
			}
			return nil
		}
		if !e.Type().IsRegular() {
			return fmt.Errorf("non-regular component file %s", p)
		}
		paths[p] = true
		return nil
	})
	if err != nil {
		return t, err
	}
	if err := contracts.CheckPortable(contracts.Keys(paths)); err != nil {
		return t, err
	}
	for _, p := range contracts.Keys(paths) {
		o, b, err := observeComponent(repo, p)
		if err != nil {
			return t, err
		}
		t.observed[p] = o
		if o.Exists {
			t.files[p] = b
		}
	}
	return t, nil
}

func decodeComponentState(files map[string][]byte, lock Lock, agentFile string, checkBlocks bool) (componentState, error) {
	s := componentState{}
	if lock.SchemaVersion != 2 || lock.Installation == nil {
		return s, errors.New("component installation requires schema-2 lock")
	}
	installation := *lock.Installation
	if digest(files[contracts.ManifestPath]) != installation.ManifestDigest {
		return s, errors.New("installed component manifest digest mismatch")
	}
	m, err := contracts.ReadManifest(files[contracts.ManifestPath], nil)
	if err != nil {
		return s, err
	}
	s.manifest = m
	if err = m.ValidateInstallation(installation); err != nil {
		return s, err
	}
	selected := map[string][]byte{}
	for p, f := range m.Files {
		if !installation.Has(f.Component) {
			continue
		}
		b, exists := files[p]
		record, tracked := lock.Files[p]
		if !exists || !tracked {
			return s, fmt.Errorf("selected payload missing: %s", p)
		}
		if f.Ownership == "managed" {
			if p == "memory-bank/README.md" {
				if record.Ownership != Generated {
					return s, errors.New("root README must be generated")
				}
			} else if record.Ownership != Managed || record.PayloadDigest != digest(b) {
				return s, fmt.Errorf("managed contract drift: %s", p)
			}
		}
		selected[p] = b
	}
	s.catalog, err = contracts.LoadCatalog(m, selected, &installation)
	if err != nil {
		return s, err
	}
	s.docs = map[string]contracts.Document{}
	for _, p := range contracts.Keys(files) {
		if !contracts.DocumentPath(p) {
			continue
		}
		if f, ok := lock.Files[p]; ok && (f.Ownership == Managed || f.Ownership == Generated) {
			continue
		}
		d, e := contracts.ParseDocument(p, files[p])
		if e != nil {
			return s, fmt.Errorf("%s: %w", p, e)
		}
		s.docs[p] = d
	}
	s.registry = contracts.EmptyRegistry()
	if installation.Has("flows") {
		b, ok := files[contracts.RegistryPath]
		if !ok || digest(b) != installation.AdoptionDigest {
			return s, errors.New("missing or corrupt adoption registry")
		}
		s.registry, err = contracts.ReadRegistry(b)
		if err != nil {
			return s, err
		}
	} else if _, ok := files[contracts.RegistryPath]; ok {
		return s, errors.New("registry exists without Flows")
	}
	s.bindings, err = contracts.ValidateRegistry(s.registry, s.catalog, s.docs)
	if err != nil {
		return s, err
	}
	adopted := map[string]bool{}
	for _, id := range contracts.Keys(s.bindings) {
		binding := s.bindings[id]
		adopted[binding.Identity.Path] = true
		s.findings = append(s.findings, contracts.ValidateBundle(s.docs[binding.Identity.Path], s.catalog.Bundles[binding.ContractID], binding.Identity, s.docs)...)
	}
	for _, p := range contracts.Keys(s.docs) {
		d := s.docs[p]
		if adopted[p] {
			continue
		}
		rules := s.catalog.DNA.Rules
		if d.Has("document_type") {
			typ, ok := s.catalog.Types[d.String("document_type")]
			if !ok {
				return s, fmt.Errorf("unknown document type: %s", p)
			}
			rules, err = contracts.MergeRules(rules, typ.Rules)
			if err != nil {
				return s, err
			}
		}
		s.findings = append(s.findings, contracts.ValidateRules(d, rules, contracts.Identity{Path: p, ContextRoot: path.Dir(p)}, s.docs)...)
	}
	contracts.SortFindings(s.findings)
	if checkBlocks {
		if agentinstructions.BuildPlanWithBlock(files["memory-bank/README.md"], contracts.ReadmeBlock(m, installation)).Status != agentinstructions.Current {
			return s, errors.New("generated README block drift")
		}
		if agentFile == "" {
			agentFile = agentinstructions.DefaultTarget
		}
		if agentinstructions.BuildPlanWithBlock(files[agentFile], contracts.AgentBlock(installation)).Status != agentinstructions.Current {
			return s, errors.New("generated AGENTS block drift")
		}
	}
	return s, nil
}

// Navigation is audited in a disposable read-only view of the prospective data.
// It never uses that view to perform actual repository mutations.
func componentNavigation(files map[string][]byte, m contracts.Manifest, s contracts.Installation) (lint.Report, error) {
	view, err := os.MkdirTemp("", "memory-bank-component-audit-")
	if err != nil {
		return lint.Report{}, err
	}
	defer os.RemoveAll(view)
	for p, b := range files {
		if !contracts.ValidPath(p) {
			return lint.Report{}, fmt.Errorf("unsafe audit path %s", p)
		}
		if p == "memory-bank/README.md" {
			v := s
			v.RendererVersion = contracts.CurrentRenderer()
			plan := agentinstructions.BuildPlanWithBlock(b, contracts.ReadmeBlock(m, v))
			if plan.Status == agentinstructions.Ambiguous {
				return lint.Report{}, errors.New("ambiguous README")
			}
			b = plan.Data
		}
		target := filepath.Join(view, filepath.FromSlash(p))
		if err = os.MkdirAll(filepath.Dir(target), 0700); err != nil {
			return lint.Report{}, err
		}
		if err = os.WriteFile(target, b, 0600); err != nil {
			return lint.Report{}, err
		}
	}
	return lint.Run(lint.Options{RepoRoot: view, ScopeRoot: "memory-bank", MaxDepth: 3})
}

// ValidateComponents is shared by doctor and lint. A legacy installation returns
// handled=false; a component marker without its lock is a corruption error.
func ValidateComponents(root, agent string) (handled bool, findings []contracts.Finding, navigation lint.Report, err error) {
	repo, e := pinRepoRoot(root)
	if e != nil {
		return false, nil, navigation, e
	}
	lock, exists, _, e := readLockSnapshot(repo)
	if e != nil {
		return true, nil, navigation, e
	}
	if !exists || lock.SchemaVersion != 2 {
		_, _, marker, e := inspectDestination(repo, contracts.ManifestPath)
		if e != nil {
			return true, nil, navigation, e
		}
		if marker {
			return true, nil, navigation, errors.New("component marker requires schema-2 lock")
		}
		return false, nil, navigation, nil
	}
	tree, e := readComponentTree(repo, lock, []string{agent})
	if e != nil {
		return true, nil, navigation, e
	}
	for p, f := range lock.Files {
		if f.Ownership == Managed && tree.observed[p].Mode != f.PayloadMode {
			return true, nil, navigation, fmt.Errorf("managed mode drift: %s", p)
		}
	}
	state, e := decodeComponentState(tree.files, lock, agent, true)
	if e != nil {
		return true, nil, navigation, e
	}
	nav, e := componentNavigation(tree.files, state.manifest, *lock.Installation)
	return true, state.findings, nav, e
}

func sameObservation(a, b observation) bool { return reflect.DeepEqual(a, b) }
func sameBytes(a, b []byte) bool            { return bytes.Equal(a, b) }

// ValidateComponentSource checks the complete local producer inventory. No
// downstream lock is expected when the explicit template profile is selected.
func ValidateComponentSource(root string) error {
	source, err := pinSourceRoot(root)
	if err != nil {
		return err
	}
	marker := filepath.Join(root, "template", "memory-bank", "components.json")
	if _, err = os.Lstat(marker); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	files, err := readSource(source)
	if err != nil {
		return err
	}
	inventory := map[string][]byte{}
	for p, f := range files {
		inventory[p] = f.data
	}
	m, err := contracts.ReadManifest(inventory[contracts.ManifestPath], inventory)
	if err != nil {
		return err
	}
	_, err = contracts.LoadCatalog(m, inventory, nil)
	return err
}
