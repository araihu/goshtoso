// Package iconlibrary describes searchable, file-based icon collections.
package iconlibrary

// Catalog is the portable output of iconpack -library. Paths are relative to
// the directory containing catalog.json. Source identifiers never depend on labels.
type Catalog struct {
	SchemaVersion int    `json:"schemaVersion"`
	Icons         []Icon `json:"icons"`
}

type Icon struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Source    string    `json:"source"`
	Reference string    `json:"reference"`
	Tags      []string  `json:"tags,omitempty"`
	Variants  []Variant `json:"variants"`
	License   string    `json:"license"`
	SourceURL string    `json:"sourceUrl"`
}

type Variant struct {
	Appearance string `json:"appearance"`
	Path       string `json:"path"`
	MIME       string `json:"mime"`
	SHA256     string `json:"sha256"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
}
