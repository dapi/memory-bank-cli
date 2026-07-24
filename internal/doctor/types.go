// Package doctor diagnoses Memory Bank adoption and governance without mutation.
package doctor

import (
	"github.com/dapi/memory-bank-cli/internal/lint"
	"github.com/dapi/memory-bank-cli/internal/ownership"
)

// ReportFormatVersion identifies the aggregate doctor JSON schema. Version 3
// adds the optional structured repair plan emitted by doctor --fix.
const ReportFormatVersion = 3

type Profile string

const (
	ProfileAuto       Profile = "auto"
	ProfileTemplate   Profile = "template"
	ProfileDownstream Profile = "downstream"
)

type Severity string

const (
	Error   Severity = "error"
	Warning Severity = "warning"
	Info    Severity = "info"
)

type Finding struct {
	Code        string   `json:"code"`
	Severity    Severity `json:"severity"`
	Group       string   `json:"group"`
	Path        string   `json:"path,omitempty"`
	Subject     string   `json:"subject,omitempty"`
	Message     string   `json:"message"`
	Remediation string   `json:"remediation"`
}

type TemplateIdentity struct {
	SchemaVersion int    `json:"schema_version"`
	Version       string `json:"version"`
	SourceRef     string `json:"source_ref"`
}

type Summary struct {
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
	Info     int `json:"info"`
}

type Report struct {
	FormatVersion    int              `json:"format_version"`
	Profile          Profile          `json:"profile"`
	RepoRoot         string           `json:"repo_root"`
	TemplateIdentity TemplateIdentity `json:"template_identity"`
	Summary          Summary          `json:"summary"`
	Findings         []Finding        `json:"findings"`
	Navigation       lint.Report      `json:"navigation"`
	Repair           *Repair          `json:"repair,omitempty"`
}

// Repair records the opt-in ownership operation performed for a finding. The
// plan is the same validated adoption plan produced by memory-bank-cli init.
type Repair struct {
	Finding string           `json:"finding"`
	Plan    ownership.Report `json:"plan"`
}

type Options struct {
	RepoRoot  string
	ScopeRoot string
	AgentFile string
	Profile   Profile
	MaxDepth  int
}
