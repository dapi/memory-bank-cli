package ownership

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/dapi/memory-bank-cli/internal/agentinstructions"
	"github.com/dapi/memory-bank-cli/internal/contracts"
)

type componentPlan struct {
	directories map[string]directoryState
	tree        componentTree
	source      map[string]payload
	mutations   []mutation
	report      Report
	next        Lock
	manifest    contracts.Manifest
}

func prepareComponents(options Options, old Lock, hasLock bool, repo pinnedRepo, lockDigest string, source map[string]payload) (componentPlan, error) {
	p := componentPlan{source: source, report: Report{FormatVersion: ReportFormatVersion, DryRun: options.DryRun, Decisions: []Decision{}}}
	if !componentHost() {
		return p, errors.New("component mutations require Linux or macOS")
	}
	if options.SkipAgentInstructions || (options.AgentFile != "" && options.AgentFile != "AGENTS.md") {
		return p, errors.New("component installation requires the canonical AGENTS.md block")
	}
	if err := checkComponentRecovery(repo, !options.DryRun); err != nil {
		return p, err
	}
	inventory := map[string][]byte{}
	for name, f := range source {
		inventory[name] = f.data
	}
	m, err := contracts.ReadManifest(inventory[contracts.ManifestPath], inventory)
	if err != nil {
		return p, err
	}
	p.manifest = m
	if _, err = contracts.LoadCatalog(m, inventory, nil); err != nil {
		return p, err
	}
	fullSelection, e := m.Select("legacy", nil, nil)
	if e != nil {
		return p, e
	}
	sourceNavigation, e := componentNavigation(inventory, m, fullSelection)
	if e != nil {
		return p, e
	}
	if sourceNavigation.ExitCode != 0 {
		return p, fmt.Errorf("source navigation invalid: %+v", sourceNavigation.Errors)
	}
	legacyMode := hasLock && old.SchemaVersion != 2
	if legacyMode {
		if !options.MigrateComponents {
			return p, errors.New("legacy installation requires --migrate-components and reviewed preview")
		}
	} else if options.MigrateComponents || options.MigrationPlanDigest != "" || len(options.MigrationResolution) > 0 {
		return p, errors.New("migration flags apply only to legacy installation")
	}
	selection, err := m.Select(options.Preset, options.Adapters, old.Installation)
	if err != nil {
		return p, err
	}
	p.tree, err = readComponentTree(repo, old, contracts.Keys(source))
	if err != nil {
		return p, err
	}
	var legacy legacyMigration
	if legacyMode {
		legacy, err = prepareLegacyMigration(options, old, p.tree, m, source, lockDigest)
		if err != nil {
			return p, err
		}
		selection.LegacySourceRef = old.Template.SourceRef
		previousSource, e := pinSourceRoot(options.SourceRoot)
		if e != nil {
			return p, e
		}
		historical, e := readGitSource(previousSource, old.Template.SourceRef)
		if e != nil {
			return p, e
		}
		for name := range historical {
			if !strings.HasPrefix(name, "memory-bank/") {
				f, ok := m.Files[name]
				if !ok || !selection.Has(f.Component) {
					return p, fmt.Errorf("migration would remove legacy adapter %s", name)
				}
			}
		}
	}
	if hasLock && !legacyMode {
		state, e := decodeComponentState(p.tree.files, old, "AGENTS.md", true)
		if e != nil {
			return p, e
		}
		for id, ref := range state.manifest.Contracts {
			if nextRef, present := m.Contracts[id]; present && nextRef.Digest != ref.Digest {
				return p, fmt.Errorf("published bundle changed behind existing ID: %s", id)
			}
		}
		if ref := old.Installation.LegacySourceRef; ref != "" && !reflect.DeepEqual(state.manifest.LegacySources[ref], m.LegacySources[ref]) {
			return p, errors.New("pinned legacy creation mapping changed")
		}
		if len(state.findings) > 0 {
			return p, fmt.Errorf("installed document validation failed: %v", state.findings)
		}
		for name, f := range old.Files {
			if f.Ownership == Managed && p.tree.observed[name].Mode != f.PayloadMode {
				return p, fmt.Errorf("managed mode drift: %s", name)
			}
		}
	}
	selection.RendererVersion = contracts.CurrentRenderer()
	selection.ManifestDigest = digest(inventory[contracts.ManifestPath])
	selected := map[string]payload{}
	for name, f := range source {
		declaration := m.Files[name]
		if selection.Has(declaration.Component) {
			f.class = Class(declaration.Ownership)
			selected[name] = f
		}
	}
	original := p.tree.files["memory-bank/README.md"]
	if original == nil {
		original = selected["memory-bank/README.md"].data
	}
	readme := agentinstructions.BuildPlanWithBlock(original, contracts.ReadmeBlock(m, selection))
	if readme.Status == agentinstructions.Ambiguous {
		return p, errors.New("ambiguous root README markers")
	}
	rawReadme := selected["memory-bank/README.md"]
	composed := rawReadme
	composed.class = Generated
	composed.data = readme.Data
	composed.digest = digest(readme.Data)
	if o := p.tree.observed["memory-bank/README.md"]; o.Exists {
		composed.mode = o.Mode
	}
	selected["memory-bank/README.md"] = composed
	// Existing managed assets must never become adapted through implicit drift.
	if !hasLock {
		for name, f := range selected {
			if f.class == Managed {
				if o := p.tree.observed[name]; o.Exists && (o.Digest != f.digest || o.Mode != f.mode) {
					return p, fmt.Errorf("unmanaged file blocks component asset %s", name)
				}
			}
		}
	}
	planningOld := old
	if legacyMode {
		planningOld, options, err = resolveLegacyOwnership(repo, selected, old, p.tree, legacy.resolution, options)
		if err != nil {
			return p, err
		}
	}
	p.mutations, p.report.Decisions, p.next, err = buildPlan(repo, selected, planningOld, hasLock, options.UserOwnedResolutions, options.AdaptedResolutions, options.DetachUserOwnedRemovals)
	if err != nil {
		return p, err
	}
	for _, d := range p.report.Decisions {
		if d.Action == Conflict {
			p.report.ConflictCount++
		}
	}
	if p.report.ConflictCount > 0 {
		return p, nil
	}
	for name, f := range selected {
		if f.class == UserOwned {
			record := p.next.Files[name]
			record.Ownership = UserOwned
			p.next.Files[name] = record
		}
	}
	record := p.next.Files["memory-bank/README.md"]
	record.Ownership = Generated
	record.BaseDigest = rawReadme.digest
	record.BaseMode = rawReadme.mode
	record.PayloadDigest = composed.digest
	record.PayloadMode = composed.mode
	p.add("memory-bank/README.md", readme.Data, Generated, "compose component index")
	p.next.Files["memory-bank/README.md"] = record
	agent := agentinstructions.BuildPlanWithBlock(p.tree.files["AGENTS.md"], contracts.AgentBlock(selection))
	if agent.Status == agentinstructions.Ambiguous {
		return p, errors.New("ambiguous AGENTS markers")
	}
	if agent.Status != agentinstructions.Current {
		p.add("AGENTS.md", agent.Data, Managed, "compose component routing")
	}
	if selection.Has("flows") && (old.Installation == nil || !old.Installation.Has("flows")) {
		if p.tree.observed[contracts.RegistryPath].Exists {
			return p, errors.New("untracked adoption registry blocks initialization")
		}
		registry := contracts.EmptyRegistry()
		if legacyMode {
			registry = legacy.registry
		}
		registryBytes, e := contracts.RegistryBytes(registry)
		if e != nil {
			return p, e
		}
		selection.AdoptionDigest = digest(registryBytes)
		p.add(contracts.RegistryPath, registryBytes, Generated, "initialize explicit adoption registry")
	}
	if legacyMode {
		for _, name := range contracts.Keys(legacy.documents) {
			p.add(name, legacy.documents[name], UserOwned, "bind existing legacy document identity")
			if _, tracked := p.next.Files[name]; tracked {
				p.next.Files[name] = File{Ownership: UserOwned}
			}
		}
	}
	p.next.SchemaVersion = 2
	p.next.Installation = &selection
	p.next.Template = Template{options.TemplateVersion, options.SourceRef}
	// A true no-op keeps timestamp and lock bytes stable.
	p.next.LastUpdate = old.LastUpdate
	changed := !hasLock || !reflect.DeepEqual(old, p.next) || len(p.mutations) > 0
	if changed {
		now := time.Now
		if options.Now != nil {
			now = options.Now
		}
		p.next.LastUpdate = UpdateRecord{options.TemplateVersion, now().UTC()}
	}
	future := p.future()
	state, err := decodeComponentState(future, p.next, "AGENTS.md", true)
	if err != nil {
		return p, err
	}
	if legacyMode {
		if !reflect.DeepEqual(state.findings, legacy.findings) && !(len(state.findings) == 0 && len(legacy.findings) == 0) {
			return p, fmt.Errorf("legacy findings changed during migration: before=%v after=%v", legacy.findings, state.findings)
		}
	} else if len(state.findings) > 0 {
		return p, fmt.Errorf("prospective document validation failed: %v", state.findings)
	}
	nav, err := componentNavigation(future, m, selection)
	if err != nil {
		return p, err
	}
	if nav.ExitCode != 0 {
		return p, fmt.Errorf("prospective navigation invalid: %+v", nav.Errors)
	}
	if changed {
		b, e := marshalLock(p.next)
		if e != nil {
			return p, e
		}
		p.add(LockFileName, b, Generated, "record component transaction")
	}
	if legacyMode {
		if err = p.describeMigration(repo, old, lockDigest, legacy); err != nil {
			return p, err
		}
		if !options.DryRun && options.MigrationPlanDigest != p.report.MigrationPlanDigest {
			return p, errors.New("migration preview digest missing or stale; regenerate and review --dry-run --json")
		}
	}
	preconditionDigest, e := p.capturePreconditions(repo)
	if e != nil {
		return p, e
	}
	if options.expectedComponentPreconditions != "" && options.expectedComponentPreconditions != preconditionDigest {
		return p, errors.New("component resolution inputs changed before apply")
	}
	return p, nil
}

func (p *componentPlan) add(name string, data []byte, class Class, reason string) {
	observed := p.tree.observed[name]
	action := Create
	if observed.Exists {
		action = UpdateFile
	}
	if observed.Exists && observed.Digest == digest(data) {
		return
	}
	d := Decision{Path: name, Action: action, Ownership: class, Reason: reason}
	mode := fileMode("100644")
	if observed.Exists {
		_, _ = fmt.Sscanf(observed.Permissions, "%o", &mode)
	}
	for i, item := range p.mutations {
		if item.decision.Path == name {
			p.mutations = append(p.mutations[:i], p.mutations[i+1:]...)
			break
		}
	}
	for i, decision := range p.report.Decisions {
		if decision.Path == name {
			p.report.Decisions = append(p.report.Decisions[:i], p.report.Decisions[i+1:]...)
			break
		}
	}
	p.mutations = append(p.mutations, mutation{decision: d, data: data, mode: mode, modeSet: true, expectedExists: observed.Exists, expectedDigest: observed.Digest, expectedMode: observed.Mode})
	p.report.Decisions = append(p.report.Decisions, d)
}
func (p componentPlan) future() map[string][]byte {
	files := map[string][]byte{}
	for name, b := range p.tree.files {
		files[name] = b
	}
	for _, item := range p.mutations {
		if item.decision.Action == Delete {
			delete(files, item.decision.Path)
		} else {
			files[item.decision.Path] = item.data
		}
	}
	return files
}
func runComponents(options Options, old Lock, hasLock bool, repo pinnedRepo, lockDigest string, source map[string]payload) (Report, error) {
	plan, err := prepareComponents(options, old, hasLock, repo, lockDigest, source)
	if err != nil {
		return Report{}, err
	}
	return applyComponentPlan(options, repo, plan)
}

func applyComponentPlan(options Options, repo pinnedRepo, plan componentPlan) (Report, error) {
	var err error
	if plan.report.ConflictCount > 0 || options.DryRun || len(plan.mutations) == 0 {
		return plan.report, nil
	}
	options.componentTransaction = true
	options.componentObservations = plan.tree.observed
	options.componentDirectories = plan.directories
	for i := range plan.mutations {
		plan.mutations[i].expectedPermissions = plan.tree.observed[plan.mutations[i].decision.Path].Permissions
	}
	// Until staging starts, every observed input must retain its full before state.
	for name, expected := range plan.tree.observed {
		actual, _, e := observeComponent(repo, name)
		if e != nil {
			return Report{}, e
		}
		if !sameObservation(actual, expected) {
			return Report{}, fmt.Errorf("component input changed: %s", name)
		}
	}
	// Bind unmodified read inputs to lock-last, in addition to the writer's per-file checks.
	mutated := map[string]bool{}
	for _, item := range plan.mutations {
		mutated[item.decision.Path] = true
	}
	last := len(plan.mutations) - 1
	for _, name := range contracts.Keys(plan.tree.observed) {
		o := plan.tree.observed[name]
		if !mutated[name] {
			plan.mutations[last].preconditions = append(plan.mutations[last].preconditions, destinationPrecondition{path: name, checkExistence: true, exists: o.Exists, digest: o.Digest, mode: o.Mode, permissions: o.Permissions})
		}
	}
	if err = applyAtomicallyPinned(options, plan.mutations, repo); err != nil {
		var committed *committedError
		plan.report.Applied = errors.As(err, &committed)
		return plan.report, err
	}
	plan.report.Applied = true
	return plan.report, nil
}
func componentFlags(options Options) bool {
	return options.Preset != "" || len(options.Adapters) > 0 || options.MigrateComponents || options.MigrationPlanDigest != "" || len(options.MigrationResolution) > 0
}
func hasComponentSource(source map[string]payload) bool {
	_, ok := source[contracts.ManifestPath]
	return ok
}
