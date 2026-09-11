package expressions

import _ "embed"

//go:generate go run ../cmd/expressionschema -source set.go -output schema.json

//go:embed schema.json
var schemaJSON []byte

// JSONSchema returns an independent copy of the JSON Schema for expression files.
// Editors and JSON Schema validators can use it for both JSON and YAML documents.
// The parser never fetches a document's $schema reference.
func JSONSchema() []byte { return append([]byte(nil), schemaJSON...) }
