package ownership

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

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
	_, decisions, _, err := buildPlan(repo, source, lock, true, nil, nil, true)
	if err != nil {
		return ResolutionPlan{}, err
	}
	entries := make([]ResolutionPlanEntry, 0, len(decisions))
	for _, decision := range decisions {
		// Agent instruction maintenance has no ownership-lock/source payload
		// identity, so it stays on the existing pull path rather than becoming a
		// resolution-plan entry.
		incoming, sourceExists := source[decision.Path]
		prior := lock.Files[decision.Path]
		localDigest, exists, err := digestDestinationFile(repo, decision.Path)
		if err != nil {
			// buildPlan has already captured and validated clean topology changes.
			// A second regular-file lookup cannot traverse the old topology; leave
			// its local identity empty rather than rejecting a deterministic plan.
			localDigest, exists = "", false
		}
		entry := ResolutionPlanEntry{Path: decision.Path, LocalPath: decision.Path, Ownership: decision.Ownership, BaseDigest: prior.BaseDigest, BaseMode: prior.BaseMode, BaseSourceRef: lock.Template.SourceRef, BasePath: decision.Path, ProposedAction: decision.Action, Reason: decision.Reason}
		if exists {
			entry.LocalDigest = localDigest
			_, info, stillExists, inspectErr := inspectDestination(repo, decision.Path)
			if inspectErr != nil || !stillExists || info == nil {
				if inspectErr == nil {
					inspectErr = fmt.Errorf("destination changed while reading plan")
				}
				return ResolutionPlan{}, inspectErr
			}
			entry.LocalMode = observedMode(info.Mode().Perm())
		}
		if sourceExists {
			entry.UpstreamDigest, entry.UpstreamMode = incoming.digest, incoming.mode
			entry.UpstreamSourceRef, entry.UpstreamPath = options.SourceRef, sourceTreePath(payloadRoot, decision.Path)
		}
		if decision.Action == Conflict && decision.Ownership == Adapted {
			entry.RequiresHumanDecision = true
			entry.AllowedActions = []string{"keep-local", "take-upstream", "apply-reviewed-merge"}
		}
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return ResolutionPlan{FormatVersion: ResolutionPlanVersion, Template: Template{Version: options.TemplateVersion, SourceRef: options.SourceRef}, LockDigest: lockDigest, Entries: entries}, nil
}

// ApplyResolutionPlan re-generates the plan from current state, rejects any
// stale or altered review carrier, and only then delegates mutation to Update.
func ApplyResolutionPlan(options Options, plan ResolutionPlan) (Report, error) {
	if plan.FormatVersion != ResolutionPlanVersion {
		return Report{}, fmt.Errorf("unsupported resolution plan format %d", plan.FormatVersion)
	}
	if plan.Template.Version != options.TemplateVersion || plan.Template.SourceRef != options.SourceRef {
		return Report{}, fmt.Errorf("resolution plan template identity does not match apply inputs")
	}
	repo, err := pinRepoRoot(options.RepoRoot)
	if err != nil {
		return Report{}, err
	}
	_, agentDecision, err := buildAgentPlan(repo, options.AgentFile)
	if err != nil {
		return Report{}, err
	}
	if agentDecision.Action != Preserve {
		return Report{}, fmt.Errorf("agent instruction state is not current; resolve it with ordinary pull before applying a resolution plan")
	}
	current, err := PlanPull(options)
	if err != nil {
		return Report{}, err
	}
	reviewed := plan
	resolutions := make(map[string]AdaptedResolution)
	for index := range reviewed.Entries {
		entry := &reviewed.Entries[index]
		if !entry.RequiresHumanDecision {
			if entry.SelectedAction != "" || entry.ReviewedContent != "" || entry.ReviewedDigest != "" || entry.ReviewedMode != "" {
				return Report{}, fmt.Errorf("resolution supplied for non-human-decision path %s", entry.Path)
			}
			continue
		}
		if entry.SelectedAction == "" {
			return Report{}, fmt.Errorf("resolution plan leaves adapted path unresolved: %s", entry.Path)
		}
		if !containsAction(entry.AllowedActions, entry.SelectedAction) {
			return Report{}, fmt.Errorf("resolution plan has invalid action %q for %s", entry.SelectedAction, entry.Path)
		}
		resolution := AdaptedResolution{Action: entry.SelectedAction}
		if entry.SelectedAction == "apply-reviewed-merge" {
			data, decodeErr := base64.StdEncoding.DecodeString(entry.ReviewedContent)
			if decodeErr != nil || entry.ReviewedDigest == "" || digest(data) != entry.ReviewedDigest || (entry.ReviewedMode != "100644" && entry.ReviewedMode != "100755") {
				return Report{}, fmt.Errorf("resolution plan has invalid reviewed merge for %s", entry.Path)
			}
			resolution.Data, resolution.Mode = data, entry.ReviewedMode
		}
		resolutions[entry.Path] = resolution
		entry.SelectedAction, entry.ReviewedContent, entry.ReviewedDigest, entry.ReviewedMode = "", "", "", ""
	}
	if !reflect.DeepEqual(reviewed, current) {
		return Report{}, fmt.Errorf("resolution plan is stale or tampered; regenerate and review it before applying")
	}
	options.AdaptedResolutions = resolutions
	options.ExpectedLockDigest = plan.LockDigest
	options.ExpectedPaths = planPreconditions(plan)
	options.DetachUserOwnedRemovals = true
	options.SkipAgentInstructions = true
	return Update(options)
}

func planPreconditions(plan ResolutionPlan) []destinationPrecondition {
	result := make([]destinationPrecondition, 0, len(plan.Entries))
	for _, entry := range plan.Entries {
		result = append(result, destinationPrecondition{path: entry.Path, exists: entry.LocalDigest != "", digest: entry.LocalDigest, mode: entry.LocalMode})
	}
	return result
}

func sourceTreePath(payloadRoot, downstream string) string {
	if payloadRoot == "" {
		return downstream
	}
	if payloadRoot == targetSourcePayloadRoot && strings.HasPrefix(downstream, downstreamPayloadRoot+"/") {
		return payloadRoot + "/" + strings.TrimPrefix(downstream, downstreamPayloadRoot+"/")
	}
	return payloadRoot + "/" + downstream
}

func containsAction(actions []string, target string) bool {
	for _, action := range actions {
		if action == target {
			return true
		}
	}
	return false
}
