package ownership

import (
	"path"
	"path/filepath"
	"strings"
)

const (
	// CanonicalTemplateRoot is the complete tracked source payload tree.
	CanonicalTemplateRoot = "template"
	// DownstreamPayloadRoot is retained for legacy source-root translation and
	// for the ownership lock stored below memory-bank/.
	DownstreamPayloadRoot = "memory-bank"
)

// CanonicalDownstreamPath maps a path relative to template/ into a downstream
// repository. The mapping is deliberately name-agnostic: stripping the
// canonical root preserves every relative component, including memory-bank/.
func CanonicalDownstreamPath(sourceRelative string) string {
	return strings.TrimPrefix(filepath.ToSlash(sourceRelative), "/")
}

// CanonicalTemplatePath is the inverse mapping used when publishing a locked
// downstream path into the canonical template tree.
func CanonicalTemplatePath(downstreamPath string) string {
	return path.Join(CanonicalTemplateRoot, filepath.ToSlash(downstreamPath))
}
