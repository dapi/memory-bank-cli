package ownership

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/dapi/memory-bank-cli/internal/contracts"
)

func planComponentResolution(options Options, repo pinnedRepo, old Lock, lockDigest string, source map[string]payload) (ResolutionPlan, error) {
	options.DryRun = true
	// Timestamps are absent from resolution entries and normalized in migration.
	options.Now = func() time.Time { return time.Unix(1, 0) }
	p, err := prepareComponents(options, old, true, repo, lockDigest, source)
	if err != nil {
		return ResolutionPlan{}, err
	}
	selection := p.next.Installation
	if selection == nil {
		s, e := p.manifest.Select(options.Preset, options.Adapters, old.Installation)
		if e != nil {
			return ResolutionPlan{}, e
		}
		s.ManifestDigest = digest(source[contracts.ManifestPath].data)
		s.RendererVersion = contracts.CurrentRenderer()
		selection = &s
	}
	preconditionDigest, e := p.capturePreconditions(repo)
	if e != nil {
		return ResolutionPlan{}, e
	}
	plan := ResolutionPlan{PreconditionDigest: preconditionDigest, FormatVersion: 2, BaseTemplate: old.Template, Template: Template{options.TemplateVersion, options.SourceRef}, LockDigest: lockDigest, Entries: []ResolutionPlanEntry{}, Installation: selection, MigrationPlanDigest: p.report.MigrationPlanDigest}
	mutationByPath := map[string]mutation{}
	for _, item := range p.mutations {
		mutationByPath[item.decision.Path] = item
	}
	for _, d := range p.report.Decisions {
		if d.Path == LockFileName {
			continue
		}
		entry := ResolutionPlanEntry{Path: d.Path, Ownership: d.Ownership, ProposedAction: d.Action, Reason: d.Reason}
		if f, ok := old.Files[d.Path]; ok {
			entry.BaseDigest = f.BaseDigest
			entry.BaseMode = f.BaseMode
			entry.BaseSourceRef = old.Template.SourceRef
			entry.BasePath = d.Path
		}
		o := p.tree.observed[d.Path]
		entry.LocalDigest = o.Digest
		entry.LocalMode = o.Mode
		if incoming, ok := source[d.Path]; ok {
			entry.UpstreamDigest = incoming.digest
			entry.UpstreamMode = incoming.mode
			entry.UpstreamSourceRef = options.SourceRef
			entry.UpstreamPath = d.Path
		}
		if item, ok := mutationByPath[d.Path]; ok && item.decision.Action != Delete && d.Path != contracts.RegistryPath {
			entry.UpstreamDigest = digest(item.data)
			mode := item.mode
			if mode == 0 && !item.modeSet {
				mode = 0644
			}
			entry.UpstreamMode = gitMode(mode)
		}
		if d.Action == Conflict && d.CanOverwrite {
			entry.RequiresHumanDecision = true
			entry.AllowedActions = []string{"keep-local", "take-upstream"}
		}
		plan.Entries = append(plan.Entries, entry)
	}
	sort.Slice(plan.Entries, func(i, j int) bool { return plan.Entries[i].Path < plan.Entries[j].Path })
	return plan, nil
}
func applyComponentResolution(options Options, plan ResolutionPlan) (Report, error) {
	if plan.Installation == nil {
		return Report{}, errors.New("component resolution selection missing")
	}
	if plan.Template != (Template{options.TemplateVersion, options.SourceRef}) {
		return Report{}, errors.New("component resolution source mismatch")
	}
	if options.Preset != "" && options.Preset != plan.Installation.Preset {
		return Report{}, errors.New("component resolution preset mismatch")
	}
	for _, adapter := range options.Adapters {
		if !contracts.Contains(plan.Installation.Adapters, adapter) {
			return Report{}, errors.New("component resolution adapter mismatch")
		}
	}
	options.Preset = plan.Installation.Preset
	options.Adapters = append([]string{}, plan.Installation.Adapters...)
	old, exists, e := ReadLock(options.RepoRoot)
	if e != nil {
		return Report{}, e
	}
	if exists && old.Installation != nil && old.Installation.Preset == plan.Installation.Preset && reflect.DeepEqual(old.Installation.Components, plan.Installation.Components) && reflect.DeepEqual(old.Installation.Adapters, plan.Installation.Adapters) {
		options.Preset = ""
		options.Adapters = nil
	}
	current, err := PlanPull(options)
	if err != nil {
		return Report{}, err
	}
	reviewed := plan
	reviewed.Entries = append([]ResolutionPlanEntry{}, plan.Entries...)
	choices := map[string]bool{}
	for i := range reviewed.Entries {
		entry := &reviewed.Entries[i]
		choice := entry.SelectedAction
		entry.SelectedAction = ""
		if entry.RequiresHumanDecision {
			if !containsResolutionAction(entry.AllowedActions, choice) {
				return Report{}, fmt.Errorf("unresolved component path %s", entry.Path)
			}
			choices[entry.Path] = choice == "take-upstream"
		} else if choice != "" {
			return Report{}, errors.New("deterministic component path cannot have a resolution")
		}
	}
	if !reflect.DeepEqual(reviewed, current) {
		return Report{}, errors.New("component resolution plan is stale or altered")
	}
	if plan.MigrationPlanDigest != "" {
		if !options.MigrateComponents || options.MigrationPlanDigest != plan.MigrationPlanDigest {
			return Report{}, errors.New("saved plan is not consent: migration flags and digest are required")
		}
	}
	options.UserOwnedResolutions = choices
	options.ExpectedLockDigest = plan.LockDigest
	options.expectedComponentPreconditions = plan.PreconditionDigest
	return Update(options)
}

func (p *componentPlan) capturePreconditions(repo pinnedRepo) (string, error) {
	directories, err := componentPlanDirectories(repo, p.tree.observed, p.mutations)
	if err != nil {
		return "", err
	}
	p.directories = directories
	b, err := contracts.Canonical(struct {
		Observed    map[string]observation    `json:"observed"`
		Directories map[string]directoryState `json:"directories"`
	}{p.tree.observed, directories})
	if err != nil {
		return "", err
	}
	return digest(b), nil
}
