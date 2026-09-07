package contracts

import (
	"bytes"
	"errors"
	"regexp"
)

// This deliberately narrow grammar is shared by draft validation and base relocation.
var referenceDestination = `(?:<([^<>\\\r\n]+)>|([^\s()<>\\]+))`
var referenceTitle = `(?:[ \t]+(?:"[^"\r\n]*"|'[^'\r\n]*'|\([^()\r\n]*\)))?`
var draftInlineReference = regexp.MustCompile(`\][ \t]*\([ \t]*` + referenceDestination + referenceTitle + `[ \t]*\)`)
var draftReferenceDefinition = regexp.MustCompile(`(?m)^[ \t]{0,3}\[[^\]\r\n]+\]:[ \t]*` + referenceDestination + referenceTitle + `[ \t]*\r?$`)
var draftReferenceSyntax = regexp.MustCompile(`\]\s*[(:]`)
var draftAutolink = regexp.MustCompile(`<(?:(?:https?://)|mailto:)[^<>\\\s]+>`)
var draftHTML = regexp.MustCompile(`<[!/A-Za-z]`)
var referenceEntity = regexp.MustCompile(`&(?:#[0-9]+|#x[0-9a-fA-F]+|[A-Za-z][A-Za-z0-9]+);`)

// Drafts are copied verbatim: any reference that relocation would change is
// rejected. This also validates the exact YAML and Markdown syntax once.
func ValidateDraftCopy(d Document, target string) error {
	relocated, err := RelocateBaseDocument(d.Raw, d.Path, target)
	if err != nil {
		return err
	}
	if !bytes.Equal(relocated, d.Raw) {
		return errors.New("cross-directory draft has relative references; use repository-absolute references")
	}
	return nil
}
