// Legacy parser/operators are retained from CLI ac7101c for source f1f04de.
package contracts

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"regexp"
	"strings"
)

var (
	legacyDesignSectionHeading = regexp.MustCompile(`(?i)^\s*(#{1,6})\s+Design Requirement Decision\s*#*\s*$`)
	legacyDesignHeading        = regexp.MustCompile(`^\s*(#{1,6})\s+`)
	legacyDecisionPattern      = regexp.MustCompile("(?im)^\\s*(?:(?:[-+*]|\\d+[.)])\\s+)?(?:\\|\\s*)?`?design\\s+required\\s*:\\s*`?(yes|no)`?(?:\\s*`)?(?:\\s*\\|.*|\\s*[.,;:]?\\s*)$")
)

func legacyDesignDecision(content string) (string, bool) {
	section := legacyDesignSection(content)
	matches := legacyDecisionPattern.FindAllStringSubmatch(section, -1)
	if len(matches) == 0 {
		return "", false
	}
	decision := matches[0][1]
	if decision != "yes" && decision != "no" {
		return "", false
	}
	for _, match := range matches[1:] {
		if match[1] != decision {
			return "", false
		}
	}
	return decision, true
}

func legacyDesignSection(content string) string {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	sectionLines := []string{}
	inSection := false
	inFence := false
	sectionDepth := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		if matches := legacyDesignSectionHeading.FindStringSubmatch(line); len(matches) > 0 {
			inSection = true
			sectionDepth = len(matches[1])
			continue
		}
		if inSection {
			if matches := legacyDesignHeading.FindStringSubmatch(line); len(matches) > 0 && len(matches[1]) <= sectionDepth {
				return strings.Join(sectionLines, "\n")
			}
			sectionLines = append(sectionLines, line)
		}
	}
	return strings.Join(sectionLines, "\n")
}

func parseLegacyFrontmatter(data []byte) (map[string]any, bool, error) {
	// YAML permits CRLF line endings. Normalize them before recognizing the
	// Markdown delimiters so governed documents work consistently across
	// platforms.
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	if !bytes.HasPrefix(data, []byte("---\n")) {
		return nil, false, nil
	}
	remainder := data[4:]
	end := -1
	for offset := 0; offset < len(remainder); {
		candidate := bytes.Index(remainder[offset:], []byte("\n---"))
		if candidate < 0 {
			break
		}
		candidate += offset
		// A delimiter must occupy its entire line. Without this check a value
		// such as "---not-a-delimiter" silently closes the frontmatter.
		afterDelimiter := candidate + len("\n---")
		if afterDelimiter == len(remainder) || remainder[afterDelimiter] == '\n' {
			end = candidate
			break
		}
		offset = afterDelimiter
	}
	if end < 0 {
		return nil, true, fmt.Errorf("unterminated YAML frontmatter")
	}
	frontmatter := map[string]any{}
	decoder := yaml.NewDecoder(bytes.NewReader(remainder[:end]))
	if err := decoder.Decode(&frontmatter); err != nil {
		return nil, true, fmt.Errorf("invalid YAML frontmatter: %w", err)
	}
	return frontmatter, true, nil
}
