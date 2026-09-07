package contracts

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"regexp"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

type Document struct {
	Path           string
	Raw, Body      []byte
	Fields         map[string]any
	HasFrontmatter bool
	closing        int
	newline        string
	keys           map[string][2]*yaml.Node
}

func (d Document) String(key string) string { s, _ := d.Fields[key].(string); return s }
func (d Document) Has(key string) bool      { _, ok := d.Fields[key]; return ok }
func ParseDocument(p string, data []byte) (Document, error) {
	d := Document{Path: p, Raw: data, Body: data, Fields: map[string]any{}, newline: "\n", keys: map[string][2]*yaml.Node{}}
	if !utf8.Valid(data) {
		return d, errors.New("Markdown must be UTF-8")
	}
	start := 0
	if bytes.HasPrefix(data, []byte("---\r\n")) {
		start = 5
		d.newline = "\r\n"
	} else if bytes.HasPrefix(data, []byte("---\n")) {
		start = 4
	} else {
		return d, nil
	}
	d.HasFrontmatter = true
	end := -1
	bodyStart := 0
	for offset := start; offset < len(data); {
		next := bytes.IndexByte(data[offset:], '\n')
		if next < 0 {
			next = len(data)
		} else {
			next += offset + 1
		}
		line := bytes.TrimSuffix(bytes.TrimSuffix(data[offset:next], []byte("\n")), []byte("\r"))
		if bytes.Equal(line, []byte("---")) {
			end = offset
			bodyStart = next
			break
		}
		offset = next
	}
	if end < 0 {
		return d, errors.New("unterminated YAML frontmatter")
	}
	d.closing = end
	d.Body = data[bodyStart:]
	raw := data[start:end]
	var node yaml.Node
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	if err := decoder.Decode(&node); err != nil && err != io.EOF {
		return d, fmt.Errorf("invalid YAML: %w", err)
	}
	if len(node.Content) == 0 {
		return d, nil
	}
	root := node.Content[0]
	if root.Kind != yaml.MappingNode {
		return d, errors.New("frontmatter must be one mapping")
	}
	if err := root.Decode(&d.Fields); err != nil {
		return d, fmt.Errorf("invalid YAML: %w", err)
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); err != io.EOF {
		return d, errors.New("multiple YAML documents are unsupported")
	}
	for i := 0; i < len(root.Content); i += 2 {
		k, v := root.Content[i], root.Content[i+1]
		if k.Kind != yaml.ScalarNode || k.Tag != "!!str" {
			return d, errors.New("frontmatter keys must be strings")
		}
		d.keys[k.Value] = [2]*yaml.Node{k, v}
	}
	return d, nil
}

// Project preserves every unrelated byte and applies the version-1 projection
// writer, rather than round-tripping an owner's YAML through a serializer.
func Project(data []byte, updates map[string]string) ([]byte, error) {
	d, err := ParseDocument("", data)
	if err != nil {
		return nil, err
	}
	order := []string{"document_type", "document_id", "flow_contract"}
	for key := range updates {
		if !Contains(order, key) {
			return nil, fmt.Errorf("unsupported projection %s", key)
		}
	}
	replace := map[int][]byte{}
	var appendLines []byte
	for _, key := range order {
		value, needed := updates[key]
		if !needed {
			continue
		}
		quoted, _ := json.Marshal(value)
		line := append([]byte(key+": "), quoted...)
		if nodes, exists := d.keys[key]; exists {
			k, v := nodes[0], nodes[1]
			if k.Style != 0 || k.Column != 1 || v.Kind != yaml.ScalarNode || v.Tag != "!!str" || v.Anchor != "" || v.Line != k.Line || v.Style&(yaml.TaggedStyle|yaml.LiteralStyle|yaml.FoldedStyle) != 0 {
				return nil, fmt.Errorf("projection %s requires a plain key and single-line untagged string", key)
			}
			physicalLines := bytes.Split(d.Raw, []byte("\n"))
			sourceLine := physicalLines[k.Line] // Node lines are relative to YAML after delimiter.
			var oneLine map[string]any
			if err := yaml.Unmarshal(sourceLine, &oneLine); err != nil || oneLine[key] != d.Fields[key] {
				return nil, fmt.Errorf("projection %s spans physical lines", key)
			}
			if d.String(key) == value {
				continue
			}
			replace[k.Line+1] = line // YAML line one follows the opening delimiter.
		} else {
			if d.Has(key) {
				return nil, fmt.Errorf("projection %s must be a direct field", key)
			}
			appendLines = append(appendLines, append(line, []byte(d.newline)...)...)
		}
	}
	if !d.HasFrontmatter {
		out := append([]byte("---\n"), appendLines...)
		out = append(out, []byte("---\n")...)
		return append(out, data...), nil
	}
	var out []byte
	for offset, lineNo := 0, 1; offset < len(data); lineNo++ {
		if offset == d.closing {
			out = append(out, appendLines...)
		}
		next := bytes.IndexByte(data[offset:], '\n')
		if next < 0 {
			next = len(data)
		} else {
			next += offset + 1
		}
		if replacement, ok := replace[lineNo]; ok {
			out = append(out, replacement...)
			ending := ""
			if next > offset && data[next-1] == '\n' {
				ending = "\n"
				if next-offset > 1 && data[next-2] == '\r' {
					ending = "\r\n"
				}
			}
			out = append(out, ending...)
		} else {
			out = append(out, data[offset:next]...)
		}
		offset = next
	}
	return out, nil
}

func DerivedPaths(d Document) []string {
	raw, exists := d.Fields["derived_from"]
	if !exists {
		return nil
	}
	values, ok := raw.([]any)
	if !ok {
		values = []any{raw}
	}
	out := []string{}
	for _, item := range values {
		value := ""
		switch x := item.(type) {
		case string:
			value = x
		case map[string]any:
			value, _ = x["path"].(string)
		}
		if strings.TrimSpace(value) != "" {
			out = append(out, value)
		}
	}
	return out
}

var headingPattern = regexp.MustCompile(`^ {0,3}(#{1,6})(?:[ \t]+(.*)|[ \t]*)$`)
var closingHeadingPattern = regexp.MustCompile(`[ \t]+#+[ \t]*$`)

func headingParts(line string) (int, string, bool) {
	m := headingPattern.FindStringSubmatch(line)
	if len(m) == 0 {
		return 0, "", false
	}
	title := strings.TrimSpace(closingHeadingPattern.ReplaceAllString(m[2], ""))
	return len(m[1]), title, true
}

// VisibleLines removes fenced blocks and HTML comments before rule matching.
func VisibleLines(data []byte) []string {
	out := []string{}
	fence := byte(0)
	fenceLen := 0
	comment := false
	fenceRun := func(line string) (byte, int, string) {
		trim := strings.TrimSpace(line)
		if len(trim) == 0 {
			return 0, 0, ""
		}
		ch := trim[0]
		n := 0
		if ch == '`' || ch == '~' {
			for n < len(trim) && trim[n] == ch {
				n++
			}
		}
		return ch, n, strings.TrimSpace(trim[n:])
	}
	for _, line := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		if fence != 0 {
			ch, n, rest := fenceRun(line)
			if ch == fence && n >= fenceLen && rest == "" {
				fence = 0
			}
			continue
		}
		if !comment && (strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "\t")) {
			continue
		}
		var visible strings.Builder
		for rest := line; rest != ""; {
			if comment {
				i := strings.Index(rest, "-->")
				if i < 0 {
					break
				}
				rest = rest[i+3:]
				comment = false
				continue
			}
			i := strings.Index(rest, "<!--")
			if i < 0 {
				visible.WriteString(rest)
				break
			}
			visible.WriteString(rest[:i])
			rest = rest[i+4:]
			comment = true
		}
		line = visible.String()
		ch, n, rest := fenceRun(line)
		if n >= 3 && (ch != '`' || !strings.Contains(rest, "`")) {
			fence = ch
			fenceLen = n
			continue
		}
		out = append(out, line)
	}
	return out
}
func Headings(d Document) map[string]bool {
	out := map[string]bool{}
	for _, line := range VisibleLines(d.Body) {
		if _, title, ok := headingParts(line); ok && title != "" {
			out[title] = true
		}
	}
	return out
}
func ContextRoot(p, typ string) (string, error) {
	if !DocumentPath(p) {
		return "", errors.New("document path is outside project document scope")
	}
	section, prefix := "", ""
	switch typ {
	case "feature":
		section, prefix = "features", "FT-"
	case "research":
		section, prefix = "research", "R-"
	case "epic":
		section, prefix = "epics", "EP-"
	}
	if section == "" {
		return path.Dir(p), nil
	}
	parts := strings.Split(p, "/")
	if len(parts) < 4 || parts[1] != section || !strings.HasPrefix(parts[2], prefix) || len(parts[2]) == len(prefix) {
		return "", fmt.Errorf("%s requires a canonical %s package", typ, section)
	}
	return strings.Join(parts[:3], "/"), nil
}
func DocumentPath(p string) bool {
	if !ValidPath(p) || !strings.HasPrefix(p, "memory-bank/") || !strings.HasSuffix(strings.ToLower(p), ".md") {
		return false
	}
	parts := strings.Split(p, "/")
	for _, excluded := range []string{".repo", "dna", "flows", "templates", "document-types", "prompts"} {
		if parts[1] == excluded {
			return false
		}
	}
	return true
}
