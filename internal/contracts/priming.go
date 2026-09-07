package contracts

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var primingIdentifier = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
var primingPlaceholder = regexp.MustCompile(`<[A-Z][A-Z0-9-]*>`)
var primingExternal = regexp.MustCompile(`(?i)^[a-z][a-z0-9+.-]*://`)

// ValidatePriming mirrors the payload's version-1 priming schema without
// executing the manifests or consulting the human prompt catalog as workflow.
func ValidatePriming(files map[string][]byte) error {
	processes := map[string]bool{}
	for _, p := range Keys(files) {
		template := p == "memory-bank/flows/templates/process/priming.yaml"
		if !template && !(path.Dir(p) == "memory-bank/flows/priming" && strings.HasSuffix(p, ".yaml")) {
			continue
		}
		var tree yaml.Node
		if err := yaml.Unmarshal(files[p], &tree); err != nil {
			return err
		}
		var ordinary func(*yaml.Node) bool
		ordinary = func(n *yaml.Node) bool {
			if n.Kind == yaml.AliasNode || n.Anchor != "" || n.Style&yaml.TaggedStyle != 0 {
				return false
			}
			for _, child := range n.Content {
				if !ordinary(child) {
					return false
				}
			}
			return true
		}
		if !ordinary(&tree) {
			return errors.New("priming aliases, anchors and tags are unsupported")
		}
		var raw map[string]any
		decoder := yaml.NewDecoder(bytes.NewReader(files[p]))
		decoder.KnownFields(true)
		if err := decoder.Decode(&raw); err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
		if err := decoder.Decode(&struct{}{}); err != io.EOF {
			return fmt.Errorf("%s: trailing YAML", p)
		}
		if len(raw) != 3 || raw["version"] != 1 {
			return fmt.Errorf("%s: invalid priming schema", p)
		}
		for _, key := range []string{"version", "process", "stages"} {
			if _, ok := raw[key]; !ok {
				return fmt.Errorf("%s: priming field missing", p)
			}
		}
		process, ok := raw["process"].(string)
		if !ok || (!template && !primingIdentifier.MatchString(process)) {
			return errors.New("invalid priming process")
		}
		if !template {
			if processes[process] {
				return errors.New("duplicate priming process")
			}
			processes[process] = true
		}
		stages, ok := raw["stages"].(map[string]any)
		if !ok || len(stages) == 0 {
			return errors.New("priming stages must be nonempty mapping")
		}
		for stage, value := range stages {
			if !primingIdentifier.MatchString(stage) {
				return errors.New("invalid priming stage")
			}
			values, ok := value.([]any)
			if !ok || len(values) == 0 {
				return errors.New("priming inputs must be nonempty list")
			}
			seen := map[string]bool{}
			for _, value := range values {
				input, ok := value.(string)
				if !ok || input == "" || seen[input] {
					return errors.New("invalid or duplicate priming input")
				}
				seen[input] = true
				if primingExternal.MatchString(input) {
					continue
				}
				if !strings.HasPrefix(input, "memory-bank/") || strings.Contains(input, "\\") || strings.Contains(strings.ToUpper(input), "TODO") || strings.Contains(input, "**") || strings.ContainsAny(input, "?[]{}") || path.Clean(input) != input || strings.Contains(input, "/../") {
					return fmt.Errorf("unsafe priming input %s", input)
				}
				if template {
					continue
				}
				remaining := primingPlaceholder.ReplaceAllString(input, "")
				if strings.ContainsAny(remaining, "<>") {
					return errors.New("invalid priming placeholder")
				}
				found := false
				if strings.Contains(input, "<") {
					prefix := strings.SplitN(input, "<", 2)[0]
					for candidate := range files {
						if strings.HasPrefix(candidate, prefix) {
							found = true
							break
						}
					}
				} else if strings.Contains(input, "*") {
					for candidate := range files {
						match, err := path.Match(input, candidate)
						if err != nil {
							return err
						}
						if match {
							found = true
							break
						}
					}
				} else {
					_, found = files[input]
				}
				if !found {
					return fmt.Errorf("%s: unresolved priming input %s", p, input)
				}
			}
		}
	}
	return nil
}
