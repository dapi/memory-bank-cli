package ownership

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
)

// Source classification is specified in docs/source-format-bridge.md.
const SourceDeclarationFile = "memory-bank-source.json"

// SupportedLegacySourceRefs returns the published immutable manifestless sources.
func SupportedLegacySourceRefs() []string {
	return []string{"f1f04de843aef45a2425d4a7351d577bbf89e940"}
}

// SupportedCapabilities is the versioned capability handshake with installers.
func SupportedCapabilities() []string { return []string{"source-format/v1", "legacy/v1"} }

type sourceDeclaration struct {
	SchemaVersion int      `json:"schema_version"`
	PayloadFormat string   `json:"payload_format"`
	Capabilities  []string `json:"capabilities"`
}

func verifySourceFormat(root, ref, payloadRoot string) error {
	marker := path.Join(payloadRoot, "components.json")
	if payloadRoot == CanonicalTemplateRoot {
		marker = path.Join(payloadRoot, "memory-bank/components.json")
	}
	entry, err := gitOutput(root, "ls-tree", ref, "--", marker)
	if err != nil {
		return fmt.Errorf("inspect component marker: %w", err)
	}
	if entry != "" {
		return errors.New("unsupported component source: this bridge supports only legacy/v1; upgrade the CLI before installing components")
	}
	data, exists, err := readSourceDeclaration(root, ref)
	if err != nil {
		return err
	}
	if !exists {
		for _, allowed := range SupportedLegacySourceRefs() {
			if strings.EqualFold(ref, allowed) {
				return nil
			}
		}
		return fmt.Errorf("unsupported manifestless source %s: use a published supported legacy commit or a declared source", ref)
	}
	var fields map[string]json.RawMessage
	if err := decodeSourceJSON(data, &fields); err != nil {
		return fmt.Errorf("invalid %s: %w", SourceDeclarationFile, err)
	}
	for key := range fields {
		if key != "schema_version" && key != "payload_format" && key != "capabilities" {
			return fmt.Errorf("invalid %s: unknown field %q", SourceDeclarationFile, key)
		}
	}
	// The first pass validates exact key spelling (encoding/json accepts case
	// aliases), duplicate keys and trailing bytes. Only typed decoding remains.
	var declaration sourceDeclaration
	if err := json.Unmarshal(data, &declaration); err != nil {
		return fmt.Errorf("invalid %s: %w", SourceDeclarationFile, err)
	}
	if declaration.SchemaVersion != 1 || declaration.PayloadFormat != "legacy/v1" {
		return fmt.Errorf("unsupported source format: schema=%d payload_format=%q", declaration.SchemaVersion, declaration.PayloadFormat)
	}
	seen := map[string]bool{}
	for _, capability := range declaration.Capabilities {
		if seen[capability] {
			return fmt.Errorf("duplicate source capability %q", capability)
		}
		seen[capability] = true
		supported := false
		for _, available := range SupportedCapabilities() {
			if capability == available {
				supported = true
			}
		}
		if !supported {
			return fmt.Errorf("unsupported source capability %q", capability)
		}
	}
	if !seen["legacy/v1"] {
		return errors.New("source declaration requires capability legacy/v1")
	}
	return nil
}

func readSourceDeclaration(root, ref string) ([]byte, bool, error) {
	tree, err := gitOutput(root, "ls-tree", "-z", ref, "--", SourceDeclarationFile)
	if err != nil {
		return nil, false, fmt.Errorf("inspect source declaration: %w", err)
	}
	if tree == "" {
		return nil, false, nil
	}
	header, name, ok := strings.Cut(strings.TrimSuffix(tree, "\x00"), "\t")
	fields := strings.Fields(header)
	if !ok || name != SourceDeclarationFile || len(fields) != 3 || fields[1] != "blob" || (fields[0] != "100644" && fields[0] != "100755") {
		return nil, false, errors.New("source declaration must be a tracked regular Git blob")
	}
	data, err := gitBytes(root, "cat-file", "blob", fields[2])
	return data, true, err
}

// Reject duplicate keys before decoding: encoding/json otherwise accepts the last
// value, which makes the declared source format ambiguous across implementations.
func decodeSourceJSON(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var value func() error
	value = func() error {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		delim, container := token.(json.Delim)
		if !container {
			return nil
		}
		switch delim {
		case '{':
			keys := map[string]bool{}
			for decoder.More() {
				keyToken, err := decoder.Token()
				if err != nil {
					return err
				}
				key, ok := keyToken.(string)
				if !ok {
					return errors.New("invalid JSON object key")
				}
				if keys[key] {
					return fmt.Errorf("duplicate JSON field %q", key)
				}
				keys[key] = true
				if err := value(); err != nil {
					return err
				}
			}
		case '[':
			for decoder.More() {
				if err := value(); err != nil {
					return err
				}
			}
		default:
			return errors.New("unexpected JSON delimiter")
		}
		_, err = decoder.Token()
		return err
	}
	if err := value(); err != nil {
		return err
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return errors.New("trailing JSON data")
	}
	decoder = json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}
