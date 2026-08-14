package ownership

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// ErrResolutionPlanApplicationUnavailable is returned until FT-054 has a
// selected, independently verifiable human-authorization mechanism and its
// protected historical-base state. A plan writer is not an authorization
// authority, so accepting selected actions before those controls exist would
// turn the plan itself into an authority to modify downstream content.
var ErrResolutionPlanApplicationUnavailable = errors.New("resolution plan application is unavailable until human authorization and protected historical-base verification are implemented")

// PlanPull creates a versioned, read-only plan for an existing ownership lock.
// It performs the same source provenance checks as Update but never calls the
// transaction writer.
func PlanPull(options Options) (ResolutionPlan, error) {
	repo, err := pinRepoRoot(options.RepoRoot)
	if err != nil {
		return ResolutionPlan{}, err
	}
	lock, exists, lockDigest, err := readLockSnapshot(repo)
	if err != nil {
		return ResolutionPlan{}, err
	}
	if !exists {
		return ResolutionPlan{}, ErrLockNotFound
	}
	if options.RepoRoot == "" || options.SourceRoot == "" || options.TemplateVersion == "" || options.SourceRef == "" {
		return ResolutionPlan{}, fmt.Errorf("repo root, source root, template version, and immutable source ref are required")
	}
	if !immutableRefPattern.MatchString(options.SourceRef) {
		return ResolutionPlan{}, fmt.Errorf("source ref must be a full 40- or 64-character hexadecimal commit ID")
	}
	pinnedSource, err := pinSourceRoot(options.SourceRoot)
	if err != nil {
		return ResolutionPlan{}, err
	}
	if err := rejectOverlappingRoots(repo, pinnedSource); err != nil {
		return ResolutionPlan{}, err
	}
	verifySource := verifySourceCheckout
	if options.verifySource != nil {
		verifySource = options.verifySource
	}
	if err := verifySource(pinnedSource.root, options.SourceRef); err != nil {
		return ResolutionPlan{}, err
	}
	var source map[string]payload
	payloadRoot := ""
	if options.verifySource == nil {
		payloadRoot, err = selectGitSourcePayloadRoot(pinnedSource.root, options.SourceRef)
		if err != nil {
			return ResolutionPlan{}, err
		}
		source, err = readGitSource(pinnedSource, options.SourceRef)
	} else {
		source, err = readSource(pinnedSource)
	}
	if err != nil {
		return ResolutionPlan{}, err
	}
	if err := verifySource(pinnedSource.root, options.SourceRef); err != nil {
		return ResolutionPlan{}, fmt.Errorf("source checkout changed while reading template: %w", err)
	}
	identities := make(map[string]destinationIdentity)
	_, decisions, _, err := buildPlan(repo, source, lock, true, nil, true, identities)
	if err != nil {
		return ResolutionPlan{}, err
	}
	if err := verifyPlanIdentities(repo, identities); err != nil {
		return ResolutionPlan{}, err
	}
	entries := make([]ResolutionPlanEntry, 0, len(decisions))
	for _, decision := range decisions {
		// Agent instruction maintenance has no ownership-lock/source payload
		// identity, so it stays on the existing pull path rather than becoming a
		// resolution-plan entry.
		incoming, sourceExists := source[decision.Path]
		prior := lock.Files[decision.Path]
		identity := identities[decision.Path]
		// The lock records a base digest and mode but does not retain a per-entry
		// source identity. In particular, lock.Template.SourceRef can advance for
		// another path while this entry keeps an older base. Do not present that
		// global ref as a reviewable base binding until protected historical-base
		// state provides a verified per-entry identity.
		entry := ResolutionPlanEntry{Path: decision.Path, LocalPath: decision.Path, Ownership: decision.Ownership, ProposedAction: decision.Action, Reason: decision.Reason}
		if identity.exists {
			entry.LocalDigest = identity.digest
			entry.LocalMode = identity.mode
		}
		if sourceExists {
			entry.UpstreamDigest, entry.UpstreamMode = incoming.digest, incoming.mode
			entry.UpstreamSourceRef, entry.UpstreamPath = options.SourceRef, sourceTreePath(payloadRoot, decision.Path)
		}
		if decision.Action == Conflict && requiresResolutionDecision(decision, prior) {
			entry.RequiresHumanDecision = true
			// A reviewed merge requires a verified historical-base snapshot and a
			// deterministic re-computation of its result. Neither exists in the
			// current lock format, so only non-merge actions can be proposed.
			entry.AllowedActions = []string{"keep-local", "take-upstream"}
		}
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return ResolutionPlan{FormatVersion: ResolutionPlanVersion, Template: Template{Version: options.TemplateVersion, SourceRef: options.SourceRef}, LockDigest: lockDigest, Entries: entries}, nil
}

func verifyPlanIdentities(repo pinnedRepo, identities map[string]destinationIdentity) error {
	paths := make([]string, 0, len(identities))
	for path := range identities {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		expected := identities[path]
		if expected.topology != nil {
			if err := verifyTopologySnapshot(repo, path, expected.topology); err != nil {
				return fmt.Errorf("destination changed while constructing resolution plan: %s: %w", path, err)
			}
			continue
		}
		observed, err := captureDestinationIdentity(repo, path)
		if err != nil {
			return fmt.Errorf("destination changed while constructing resolution plan: %s: %w", path, err)
		}
		if observed.exists != expected.exists || observed.digest != expected.digest || observed.mode != expected.mode {
			return fmt.Errorf("destination changed while constructing resolution plan: %s", path)
		}
	}
	return nil
}

func requiresResolutionDecision(decision Decision, prior File) bool {
	if decision.Action != Conflict {
		return false
	}
	// Adapted conflicts always require a human decision. A managed file only
	// becomes selectable when it is an already tracked managed path with local
	// drift; unmanaged collisions continue through the conservative default
	// pull path.
	return prior.Ownership == Adapted || prior.Ownership == Managed && decision.Ownership == Managed
}

// ApplyResolutionPlan is intentionally unavailable until the feature's
// authorization and protected historical-base contracts are implemented.
// Keeping this guard before reading or acting on plan fields prevents an AI or
// any other plan writer from treating selected_action as human approval.
func ApplyResolutionPlan(options Options, plan ResolutionPlan) (Report, error) {
	return Report{}, ErrResolutionPlanApplicationUnavailable
}

func sourceTreePath(payloadRoot, downstream string) string {
	if payloadRoot == "" {
		return downstream
	}
	// This is the inverse of downstreamPayloadPath. Legacy source roots map
	// their relative payload below memory-bank/, while canonical template/
	// sources preserve every downstream component.
	if payloadRoot != targetSourcePayloadRoot && strings.HasPrefix(downstream, downstreamPayloadRoot+"/") {
		downstream = strings.TrimPrefix(downstream, downstreamPayloadRoot+"/")
	}
	return payloadRoot + "/" + downstream
}
