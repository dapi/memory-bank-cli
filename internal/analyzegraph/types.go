// Package analyzegraph analyses the typed evidence in an execution handoff.
package analyzegraph

const ReportFormatVersion = 1

// Options controls a read-only handoff analysis.
type Options struct {
	HandoffPath string
}

// Report is the stable machine-readable and human-readable analysis result.
// Findings, recommendations, and evidence are deliberately separate so that
// consumers do not have to turn an advisory result back into raw graph data.
type Report struct {
	FormatVersion        int              `json:"format_version"`
	HandoffPath          string           `json:"handoff_path"`
	HandoffKind          string           `json:"handoff_kind"`
	HandoffSchemaVersion int              `json:"handoff_schema_version"`
	Task                 Task             `json:"task"`
	Findings             []Finding        `json:"findings"`
	Recommendations      []Recommendation `json:"recommendations"`
	Evidence             []Evidence       `json:"evidence"`
}

type Task struct {
	ID     string `json:"id,omitempty"`
	Source string `json:"source,omitempty"`
}

type NodeRef struct {
	ID      string   `json:"id"`
	Type    string   `json:"type,omitempty"`
	Label   string   `json:"label,omitempty"`
	Source  string   `json:"source,omitempty"`
	Sources []string `json:"sources,omitempty"`
}

type Evidence struct {
	Kind    string   `json:"kind"`
	Node    string   `json:"node,omitempty"`
	From    string   `json:"from,omitempty"`
	To      string   `json:"to,omitempty"`
	Type    string   `json:"type"`
	Source  string   `json:"source,omitempty"`
	Sources []string `json:"sources,omitempty"`
	Count   int      `json:"count,omitempty"`
	Weak    bool     `json:"weak,omitempty"`
}

type Finding struct {
	Kind          string     `json:"kind"`
	Severity      string     `json:"severity"`
	Subject       string     `json:"subject"`
	Message       string     `json:"message"`
	Nodes         []string   `json:"nodes"`
	RelationTypes []string   `json:"relation_types"`
	Sources       []string   `json:"sources"`
	Evidence      []Evidence `json:"evidence"`
}

type Recommendation struct {
	Kind              string     `json:"kind"`
	Summary           string     `json:"summary"`
	ContributingNodes []NodeRef  `json:"contributing_nodes"`
	RelationTypes     []string   `json:"relation_types"`
	Sources           []string   `json:"sources"`
	Evidence          []Evidence `json:"evidence"`
}
