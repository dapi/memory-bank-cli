// Package ownership implements the versioned Memory Bank ownership and update contract.
package ownership

import "time"

const (
	LockFileName          = "memory-bank/.lock"
	CurrentSchemaVersion  = 1
	ReportFormatVersion   = 1
	ResolutionPlanVersion = 1
)

type Class string

const (
	Managed   Class = "managed"
	Adapted   Class = "adapted"
	UserOwned Class = "user-owned"
	Generated Class = "generated"
)

type Template struct {
	Version   string `json:"version"`
	SourceRef string `json:"source_ref"`
}

type UpdateRecord struct {
	Version string    `json:"version"`
	At      time.Time `json:"at"`
}

type File struct {
	Ownership     Class  `json:"ownership"`
	BaseDigest    string `json:"base_digest,omitempty"`
	PayloadDigest string `json:"payload_digest,omitempty"`
	BaseMode      string `json:"base_mode,omitempty"`
	PayloadMode   string `json:"payload_mode,omitempty"`
}

type Lock struct {
	SchemaVersion int             `json:"schema_version"`
	Template      Template        `json:"template"`
	LastUpdate    UpdateRecord    `json:"last_update"`
	Files         map[string]File `json:"files"`
}

type Action string

const (
	Create     Action = "create"
	UpdateFile Action = "update"
	Preserve   Action = "preserve"
	Conflict   Action = "conflict"
	Delete     Action = "delete"
)

type Decision struct {
	Path         string `json:"path"`
	Ownership    Class  `json:"ownership"`
	Action       Action `json:"action"`
	Reason       string `json:"reason"`
	Diff         string `json:"diff,omitempty"`
	CanOverwrite bool   `json:"-"`
}

type Report struct {
	FormatVersion int        `json:"format_version"`
	DryRun        bool       `json:"dry_run"`
	Applied       bool       `json:"applied"`
	Decisions     []Decision `json:"decisions"`
	ConflictCount int        `json:"conflict_count"`
	DriftCount    int        `json:"drift_count"`
}

// ResolutionPlan is a reviewable, non-mutating snapshot of a complete pull.
// SelectedAction is the only reviewer-authored field; apply regenerates and
// compares every other field before allowing the ownership transaction.
type ResolutionPlan struct {
	FormatVersion int                   `json:"format_version"`
	BaseTemplate  Template              `json:"base_template"`
	Template      Template              `json:"template"`
	LockDigest    string                `json:"lock_digest"`
	Entries       []ResolutionPlanEntry `json:"entries"`
}

type ResolutionPlanEntry struct {
	Path                   string          `json:"path"`
	Ownership              Class           `json:"ownership"`
	BaseDigest             string          `json:"base_digest,omitempty"`
	BaseMode               string          `json:"base_mode,omitempty"`
	BaseSourceRef          string          `json:"base_source_ref,omitempty"`
	BasePath               string          `json:"base_path,omitempty"`
	LocalDigest            string          `json:"local_digest,omitempty"`
	LocalMode              string          `json:"local_mode,omitempty"`
	UpstreamDigest         string          `json:"upstream_digest,omitempty"`
	UpstreamMode           string          `json:"upstream_mode,omitempty"`
	UpstreamSourceRef      string          `json:"upstream_source_ref,omitempty"`
	UpstreamPath           string          `json:"upstream_path,omitempty"`
	ProposedAction         Action          `json:"proposed_action"`
	Reason                 string          `json:"reason"`
	RequiresHumanDecision  bool            `json:"requires_human_decision"`
	AllowedActions         []string        `json:"allowed_actions,omitempty"`
	SelectedAction         string          `json:"selected_action,omitempty"`
	Merge                  *MergeCandidate `json:"merge,omitempty"`
	MergeUnavailableReason string          `json:"merge_unavailable_reason,omitempty"`
}

type MergeCandidate struct {
	Algorithm     string `json:"algorithm"`
	ContentBase64 string `json:"content_base64"`
	Digest        string `json:"digest"`
	Mode          string `json:"mode"`
}

type AdaptedResolution struct {
	Action string
	Data   []byte
	Mode   string
}

type Options struct {
	RepoRoot        string
	SourceRoot      string
	TemplateVersion string
	SourceRef       string
	DryRun          bool
	// UserOwnedResolutions maps user-owned managed-file collisions to their
	// explicit resolution: false keeps local content, true replaces it with the
	// incoming source payload.
	UserOwnedResolutions    map[string]bool
	AdaptedResolutions      map[string]AdaptedResolution
	ExpectedLockDigest      string
	ExpectedPaths           []destinationPrecondition
	DetachUserOwnedRemovals bool
	SkipAgentInstructions   bool
	AgentFile               string
	Now                     func() time.Time
	// verifySource is replaced by unit tests that use synthetic source trees.
	// CLI callers always use the Git-backed provenance verifier.
	verifySource func(string, string) error
	// BeforeMutation is used by tests after staging to simulate an interrupted update.
	BeforeMutation func(Decision) error
}
