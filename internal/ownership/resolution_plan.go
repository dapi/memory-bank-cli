package ownership

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
)

// PlanPull creates a complete, read-only resolution plan for an existing lock.
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
		return ResolutionPlan{}, errors.New("repo root, source root, template version, and immutable source ref are required")
	}
	if !immutableRefPattern.MatchString(options.SourceRef) {
		return ResolutionPlan{}, errors.New("source ref must be a full 40- or 64-character hexadecimal commit ID")
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
		if err == nil {
			source, err = readGitSource(pinnedSource, options.SourceRef)
		}
	} else {
		source, err = readSource(pinnedSource)
	}
	if err != nil {
		return ResolutionPlan{}, err
	}
	if err := verifySource(pinnedSource.root, options.SourceRef); err != nil {
		return ResolutionPlan{}, fmt.Errorf("source checkout changed while reading template: %w", err)
	}
	_, decisions, _, err := buildPlan(repo, source, lock, true, nil, nil, false)
	if err != nil {
		return ResolutionPlan{}, err
	}

	var historical map[string]payload
	historicalRoot := ""
	historicalReason := ""
	if options.verifySource == nil {
		historicalRoot, err = selectGitSourcePayloadRoot(pinnedSource.root, lock.Template.SourceRef)
		if err == nil {
			historical, err = readGitSource(pinnedSource, lock.Template.SourceRef)
		}
		if err != nil {
			historicalReason = "historical source ref or payload is unavailable"
		}
	} else {
		historicalReason = "historical Git verification is unavailable for synthetic source"
	}

	entries := make([]ResolutionPlanEntry, 0, len(decisions))
	for _, decision := range decisions {
		incoming, sourceExists := source[decision.Path]
		prior, tracked := lock.Files[decision.Path]
		entry := ResolutionPlanEntry{
			Path: decision.Path, Ownership: decision.Ownership,
			ProposedAction: decision.Action, Reason: decision.Reason,
		}
		if tracked {
			entry.BaseDigest, entry.BaseMode = prior.BaseDigest, prior.BaseMode
			entry.BaseSourceRef = lock.Template.SourceRef
			entry.BasePath = sourceTreePath(historicalRoot, decision.Path)
		}
		localInfo, localData, localExists, readErr := readPlanDestination(repo, decision.Path)
		if readErr != nil {
			return ResolutionPlan{}, readErr
		}
		if localExists {
			entry.LocalDigest = digest(localData)
			entry.LocalMode = observedMode(localInfo.Mode().Perm())
		}
		if sourceExists {
			entry.UpstreamDigest, entry.UpstreamMode = incoming.digest, incoming.mode
			entry.UpstreamSourceRef = options.SourceRef
			entry.UpstreamPath = sourceTreePath(payloadRoot, decision.Path)
		}
		if decision.CanOverwrite && decision.Action == Conflict {
			entry.RequiresHumanDecision = true
			entry.AllowedActions = []string{"keep-local", "take-upstream"}
		} else if decision.Action == Conflict && decision.Ownership == Adapted {
			entry.RequiresHumanDecision = true
			switch {
			case !localExists && sourceExists:
				entry.AllowedActions = []string{"take-upstream"}
				entry.MergeUnavailableReason = "lock v1 cannot retain a reviewed local deletion while upstream is present"
			case localExists && !sourceExists:
				entry.AllowedActions = []string{"keep-local", "take-upstream"}
				entry.MergeUnavailableReason = "mechanical merge requires present local and upstream files"
			case !localExists && !sourceExists:
				entry.AllowedActions = []string{"keep-local", "take-upstream"}
				entry.MergeUnavailableReason = "mechanical merge requires present local and upstream files"
			default:
				entry.AllowedActions = []string{"keep-local", "take-upstream"}
			}
			if localExists && sourceExists && historicalReason != "" {
				entry.MergeUnavailableReason = historicalReason
			} else if localExists && sourceExists {
				base, ok := historical[decision.Path]
				if !ok {
					entry.MergeUnavailableReason = "historical source does not contain the recorded path"
				} else if base.digest != prior.BaseDigest || !modeMatches(base.mode, prior.BaseMode) {
					entry.MergeUnavailableReason = "historical source payload does not match the ownership lock base"
				} else {
					localMode := entry.LocalMode
					if localMode == "" {
						localMode = prior.BaseMode
					}
					merged, mode, mergeErr := mechanicalMerge(base.data, localData, incoming.data, base.mode, localMode, incoming.mode)
					if mergeErr != nil {
						entry.MergeUnavailableReason = mergeErr.Error()
					} else {
						entry.AllowedActions = append(entry.AllowedActions, "apply-reviewed-merge")
						entry.Merge = &MergeCandidate{
							Algorithm: mechanicalMergeAlgorithm, ContentBase64: base64.StdEncoding.EncodeToString(merged),
							Digest: digest(merged), Mode: mode,
						}
					}
				}
			}
		}
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	plan := ResolutionPlan{
		FormatVersion: ResolutionPlanVersion,
		BaseTemplate:  lock.Template,
		Template:      Template{Version: options.TemplateVersion, SourceRef: options.SourceRef},
		LockDigest:    lockDigest,
		Entries:       entries,
	}
	if err := verifyPlanLocalIdentities(repo, plan); err != nil {
		return ResolutionPlan{}, err
	}
	return plan, nil
}

// ApplyResolutionPlan regenerates all deterministic plan fields, validates the
// selected resolution overlays, and delegates one atomic mutation to Update.
func ApplyResolutionPlan(options Options, plan ResolutionPlan) (Report, error) {
	if plan.FormatVersion != ResolutionPlanVersion {
		return Report{}, fmt.Errorf("unsupported resolution plan format %d", plan.FormatVersion)
	}
	if plan.Template.Version != options.TemplateVersion || plan.Template.SourceRef != options.SourceRef {
		return Report{}, errors.New("resolution plan template identity does not match current pull source")
	}
	current, err := PlanPull(options)
	if err != nil {
		return Report{}, err
	}
	if plan.BaseTemplate != current.BaseTemplate {
		return Report{}, errors.New("resolution plan base template identity changed")
	}

	reviewed := plan
	resolutions := make(map[string]AdaptedResolution)
	userOwnedResolutions := make(map[string]bool)
	for index := range reviewed.Entries {
		entry := &reviewed.Entries[index]
		selected := entry.SelectedAction
		entry.SelectedAction = ""
		if !entry.RequiresHumanDecision {
			if selected != "" {
				return Report{}, fmt.Errorf("resolution supplied for deterministic path %s", entry.Path)
			}
			continue
		}
		if selected == "" {
			return Report{}, fmt.Errorf("resolution plan leaves adapted path unresolved: %s", entry.Path)
		}
		if !containsResolutionAction(entry.AllowedActions, selected) {
			return Report{}, fmt.Errorf("resolution plan action %q is unavailable for %s", selected, entry.Path)
		}
		if entry.Ownership == UserOwned {
			userOwnedResolutions[entry.Path] = selected == "take-upstream"
			continue
		}
		resolution := AdaptedResolution{Action: selected}
		if selected == "apply-reviewed-merge" {
			if entry.Merge == nil || entry.Merge.Algorithm != mechanicalMergeAlgorithm {
				return Report{}, fmt.Errorf("resolution plan has no valid reviewed merge for %s", entry.Path)
			}
			data, decodeErr := base64.StdEncoding.DecodeString(entry.Merge.ContentBase64)
			if decodeErr != nil || digest(data) != entry.Merge.Digest || entry.Merge.Mode != "100644" && entry.Merge.Mode != "100755" {
				return Report{}, fmt.Errorf("resolution plan has an invalid reviewed merge payload for %s", entry.Path)
			}
			resolution.Data, resolution.Mode = data, entry.Merge.Mode
		}
		resolutions[entry.Path] = resolution
	}
	if !reflect.DeepEqual(reviewed, current) {
		return Report{}, errors.New("resolution plan is stale or altered; regenerate and review it before applying")
	}

	repo, err := pinRepoRoot(options.RepoRoot)
	if err != nil {
		return Report{}, err
	}
	agentTarget := options.AgentFile
	if agentTarget == "" {
		agentTarget = "AGENTS.md"
	}
	templateOwnsAgent := false
	for _, entry := range current.Entries {
		if entry.Path == agentTarget {
			templateOwnsAgent = true
			break
		}
	}
	if !templateOwnsAgent {
		_, agentDecision, agentErr := buildAgentPlan(repo, options.AgentFile)
		if agentErr != nil {
			return Report{}, agentErr
		}
		if agentDecision.Action != Preserve {
			return Report{}, errors.New("agent instruction state is not current; resolve it with ordinary pull before applying a resolution plan")
		}
	}
	options.AdaptedResolutions = resolutions
	options.UserOwnedResolutions = userOwnedResolutions
	options.ExpectedLockDigest = plan.LockDigest
	options.ExpectedPaths = planPreconditions(plan)
	options.DetachUserOwnedRemovals = true
	options.SkipAgentInstructions = true
	return Update(options)
}

func readPlanDestination(repo pinnedRepo, path string) (os.FileInfo, []byte, bool, error) {
	_, info, exists, err := inspectDestination(repo, path)
	if err != nil || !exists {
		return nil, nil, exists, err
	}
	readInfo, data, err := secureReadDestination(repo, path)
	if err != nil {
		return nil, nil, false, err
	}
	if !os.SameFile(info, readInfo) {
		return nil, nil, false, fmt.Errorf("destination changed while constructing resolution plan: %s", path)
	}
	return readInfo, data, true, nil
}

func verifyPlanLocalIdentities(repo pinnedRepo, plan ResolutionPlan) error {
	for _, entry := range plan.Entries {
		info, data, exists, err := readPlanDestination(repo, entry.Path)
		if err != nil {
			return err
		}
		if exists != (entry.LocalDigest != "") {
			return fmt.Errorf("destination changed while constructing resolution plan: %s", entry.Path)
		}
		if exists && (digest(data) != entry.LocalDigest || !modeMatches(observedMode(info.Mode().Perm()), entry.LocalMode)) {
			return fmt.Errorf("destination changed while constructing resolution plan: %s", entry.Path)
		}
	}
	return nil
}

func planPreconditions(plan ResolutionPlan) []destinationPrecondition {
	result := make([]destinationPrecondition, 0, len(plan.Entries))
	for _, entry := range plan.Entries {
		result = append(result, destinationPrecondition{
			path: entry.Path, checkExistence: true, exists: entry.LocalDigest != "", digest: entry.LocalDigest, mode: entry.LocalMode,
		})
	}
	return result
}

func sourceTreePath(payloadRoot, downstream string) string {
	if payloadRoot == "" {
		return ""
	}
	if payloadRoot == targetSourcePayloadRoot {
		return payloadRoot + "/" + downstream
	}
	return payloadRoot + "/" + strings.TrimPrefix(downstream, downstreamPayloadRoot+"/")
}

func containsResolutionAction(actions []string, target string) bool {
	for _, action := range actions {
		if action == target {
			return true
		}
	}
	return false
}
