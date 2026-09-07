package contracts

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type textReplacement struct {
	start, end int
	value      []byte
}

// RelocateBaseDocument preserves the resolved target of supported Markdown and
// derived_from references while instantiating a base template at another path.
// It changes only reference tokens, never the source file or unrelated text.
func RelocateBaseDocument(data []byte, from, to string) ([]byte, error) {
	if !ValidPath(from) || !ValidPath(to) {
		return nil, errors.New("unsafe template relocation path")
	}
	if path.Dir(from) == path.Dir(to) {
		return append([]byte{}, data...), nil
	}
	d, err := ParseDocument(from, data)
	if err != nil {
		return nil, err
	}
	relocate := func(ref string) (string, error) {
		if strings.ContainsAny(ref, "\\\r\n") || referenceEntity.MatchString(ref) {
			return "", errors.New("escaped references are unsupported")
		}
		if ref == "" || strings.HasPrefix(ref, "/") || strings.HasPrefix(ref, "#") || strings.HasPrefix(ref, "https://") || strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "mailto:") {
			return ref, nil
		}
		if strings.ContainsAny(ref, "\\\r\n") {
			return "", errors.New("unsupported relative reference syntax")
		}
		bare, suffix := ref, ""
		if i := strings.IndexAny(ref, "?#"); i >= 0 {
			bare, suffix = ref[:i], ref[i:]
		}
		decoded, e := url.PathUnescape(bare)
		if e != nil || strings.HasPrefix(decoded, "/") || strings.ContainsAny(decoded, "\\\r\n") {
			return "", errors.New("unsupported encoded relative destination")
		}
		target := path.Clean(path.Join(path.Dir(from), decoded))
		if !ValidPath(target) {
			return "", fmt.Errorf("template reference escapes repository: %s", ref)
		}
		rel, e := filepath.Rel(filepath.FromSlash(path.Dir(to)), filepath.FromSlash(target))
		if e != nil {
			return "", e
		}
		rel = filepath.ToSlash(rel)
		if strings.HasSuffix(bare, "/") && !strings.HasSuffix(rel, "/") {
			rel += "/"
		}
		parts := strings.Split(rel, "/")
		for i := range parts {
			parts[i] = url.PathEscape(parts[i])
		}
		return strings.Join(parts, "/") + suffix, nil
	}
	changes := []textReplacement{}
	// Node positions select the exact scalar token; aliases, folded scalars and
	// unfamiliar derived_from shapes fail instead of rewriting a YAML subtree.
	if nodes, ok := d.keys["derived_from"]; ok {
		var visit func(*yaml.Node) error
		visit = func(node *yaml.Node) error {
			if node.Anchor != "" || node.Style&yaml.TaggedStyle != 0 {
				return errors.New("dependency anchors and tags are unsupported")
			}
			switch node.Kind {
			case yaml.SequenceNode:
				for _, child := range node.Content {
					if child.Kind != yaml.ScalarNode && child.Kind != yaml.MappingNode {
						return errors.New("unsupported nested dependency sequence")
					}
					if e := visit(child); e != nil {
						return e
					}
				}
			case yaml.MappingNode:
				found := false
				for i := 0; i < len(node.Content); i += 2 {
					if node.Content[i].Value != "path" && node.Content[i].Value != "fit" {
						return errors.New("unsupported dependency field")
					}
					if node.Content[i].Value == "path" {
						if node.Content[i+1].Kind != yaml.ScalarNode {
							return errors.New("dependency path must be a string")
						}
						if found {
							return errors.New("duplicate dependency path")
						}
						found = true
						if e := visit(node.Content[i+1]); e != nil {
							return e
						}
					}
				}
				if !found {
					return errors.New("dependency mapping lacks path")
				}
			case yaml.ScalarNode:
				if node.Tag != "!!str" || node.Anchor != "" || node.Style&(yaml.TaggedStyle|yaml.LiteralStyle|yaml.FoldedStyle) != 0 {
					return errors.New("unsupported dependency scalar")
				}
				value, e := relocate(node.Value)
				if e != nil {
					return e
				}
				if value == node.Value {
					return nil
				}
				start, end, e := yamlScalarRange(data, node)
				if e != nil {
					return e
				}
				b, _ := json.Marshal(value)
				changes = append(changes, textReplacement{start, end, b})
			default:
				return errors.New("unsupported dependency alias or node")
			}
			return nil
		}
		if err = visit(nodes[1]); err != nil {
			return nil, err
		}
	}
	masked := maskMarkdownCode(d.Body)
	bodyOffset := len(data) - len(d.Body)
	for _, pattern := range []*regexp.Regexp{draftInlineReference, draftReferenceDefinition} {
		matches := pattern.FindAllSubmatchIndex(masked, -1)
		for _, match := range matches {
			start, end := match[2], match[3]
			if start < 0 {
				start, end = match[4], match[5]
			}
			ref := string(d.Body[start:end])
			value, e := relocate(ref)
			if e != nil {
				return nil, e
			}
			if value != ref {
				changes = append(changes, textReplacement{bodyOffset + start, bodyOffset + end, []byte(value)})
			}
			for i := match[0]; i < match[1]; i++ {
				masked[i] = ' '
			}
		}
	}
	if draftReferenceSyntax.Match(masked) || draftHTML.Match(draftAutolink.ReplaceAll(masked, nil)) {
		return nil, errors.New("unsupported template reference syntax")
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].start < changes[j].start })
	result := []byte{}
	offset := 0
	for _, change := range changes {
		if change.start < offset {
			return nil, errors.New("overlapping template references")
		}
		result = append(result, data[offset:change.start]...)
		result = append(result, change.value...)
		offset = change.end
	}
	return append(result, data[offset:]...), nil
}
func yamlScalarRange(data []byte, node *yaml.Node) (int, int, error) {
	lines := bytes.SplitAfter(data, []byte("\n"))
	if node.Line < 1 || node.Line >= len(lines) {
		return 0, 0, errors.New("dependency position outside frontmatter")
	}
	offset := 0
	for _, line := range lines[:node.Line] {
		offset += len(line)
	}
	runes := []rune(string(lines[node.Line]))
	if node.Column < 1 || node.Column > len(runes) {
		return 0, 0, errors.New("invalid dependency column")
	}
	offset += len(string(runes[:node.Column-1]))
	end := offset
	if node.Style == 0 {
		end += len(node.Value)
		if end > len(data) || string(data[offset:end]) != node.Value {
			return 0, 0, errors.New("multiline dependency is unsupported")
		}
	} else {
		quote := data[offset]
		if quote != '\'' && quote != '"' {
			return 0, 0, errors.New("unsupported dependency quoting")
		}
		end++
		for end < len(data) {
			if data[end] == '\n' || data[end] == '\r' {
				return 0, 0, errors.New("multiline dependency is unsupported")
			}
			if quote == '"' && data[end] == '\\' {
				end += 2
				continue
			}
			if data[end] == quote {
				if quote == '\'' && end+1 < len(data) && data[end+1] == quote {
					end += 2
					continue
				}
				end++
				break
			}
			end++
		}
	}
	if end > len(data) {
		return 0, 0, errors.New("unterminated dependency")
	}
	var value string
	if e := yaml.Unmarshal(data[offset:end], &value); e != nil || value != node.Value {
		return 0, 0, errors.New("dependency token mismatch")
	}
	return offset, end, nil
}

// Masking keeps byte offsets intact, including CRLF. Examples inside code and
// comments are not navigation references and must remain verbatim.
func maskMarkdownCode(data []byte) []byte {
	out := append([]byte{}, data...)
	fence := byte(0)
	fenceLength := 0
	comment := false
	offset := 0
	blank := func(start, end int) {
		for i := start; i < end; i++ {
			if out[i] != '\n' && out[i] != '\r' {
				out[i] = ' '
			}
		}
	}
	for _, line := range bytes.SplitAfter(data, []byte("\n")) {
		trim := strings.TrimSpace(string(line))
		run := 0
		var ch byte
		if len(trim) > 0 {
			ch = trim[0]
			if ch == '`' || ch == '~' {
				for run < len(trim) && trim[run] == ch {
					run++
				}
			}
		}
		if fence != 0 {
			blank(offset, offset+len(line))
			if ch == fence && run >= fenceLength && strings.TrimSpace(trim[run:]) == "" {
				fence = 0
			}
			offset += len(line)
			continue
		}
		if !comment && run >= 3 && (ch != '`' || !strings.Contains(trim[run:], "`")) {
			fence = ch
			fenceLength = run
			blank(offset, offset+len(line))
			offset += len(line)
			continue
		}
		if !comment && (bytes.HasPrefix(line, []byte("    ")) || bytes.HasPrefix(line, []byte("\t"))) {
			blank(offset, offset+len(line))
			offset += len(line)
			continue
		}
		for i := 0; i < len(line); {
			if comment {
				end := bytes.Index(line[i:], []byte("-->"))
				if end < 0 {
					blank(offset+i, offset+len(line))
					break
				}
				blank(offset+i, offset+i+end+3)
				i += end + 3
				comment = false
				continue
			}
			start := bytes.Index(line[i:], []byte("<!--"))
			if start < 0 {
				break
			}
			i += start
			comment = true
		}
		offset += len(line)
	}
	for i := 0; i < len(out); i++ {
		if out[i] != '`' {
			continue
		}
		n := 1
		for i+n < len(out) && out[i+n] == '`' {
			n++
		}
		end := bytes.Index(out[i+n:], bytes.Repeat([]byte{'`'}, n))
		if end >= 0 {
			last := i + n + end + n
			blank(i, last)
			i = last - 1
		} else {
			i += n - 1
		}
	}
	return out
}
