package contracts

import "regexp"

// This deliberately narrow grammar is shared by draft validation and base relocation.
var referenceDestination = `(?:<([^<>\\\r\n]+)>|([^\s()<>\\]+))`
var referenceTitle = `(?:[ \t]+(?:"[^"\r\n]*"|'[^'\r\n]*'|\([^()\r\n]*\)))?`
var draftInlineReference = regexp.MustCompile(`\][ \t]*\([ \t]*` + referenceDestination + referenceTitle + `[ \t]*\)`)
var draftReferenceDefinition = regexp.MustCompile(`(?m)^[ \t]{0,3}\[[^\]\r\n]+\]:[ \t]*` + referenceDestination + referenceTitle + `[ \t]*\r?$`)
var draftReferenceSyntax = regexp.MustCompile(`\]\s*[(:]`)
var draftAutolink = regexp.MustCompile(`<(?:(?:https?://)|mailto:)[^<>\\\s]+>`)
var draftHTML = regexp.MustCompile(`<[!/A-Za-z]`)
var referenceEntity = regexp.MustCompile(`&(?:#[0-9]+|#x[0-9a-fA-F]+|[A-Za-z][A-Za-z0-9]+);`)
