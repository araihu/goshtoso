package iconpack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Selection is the JSON-first input manifest. YAML uses the same field names.
type Selection struct {
	SchemaVersion int      `json:"schemaVersion" yaml:"schemaVersion"`
	Names         []string `json:"names" yaml:"names"`
}

func selectionNames(opts Options) ([]string, error) {
	if len(opts.Names) > 0 && opts.SelectionManifest != "" {
		return nil, fmt.Errorf("name and manifest selection are mutually exclusive")
	}
	if len(opts.Names) == 0 && opts.SelectionManifest == "" {
		if opts.ConfigPath != "" {
			return nil, nil
		}
		return nil, fmt.Errorf("at least one exact canonical name or a selection manifest is required")
	}
	names := append([]string(nil), opts.Names...)
	if opts.SelectionManifest != "" {
		manifest, err := loadSelection(opts.SelectionManifest)
		if err != nil {
			return nil, err
		}
		names = manifest.Names
	}
	seen := make(map[string]struct{}, len(names))
	for _, name := range names {
		if name == "" || name != strings.TrimSpace(name) || strings.ContainsAny(name, "\r\n\t\x00") {
			return nil, fmt.Errorf("invalid canonical name %q", name)
		}
		if _, exists := seen[name]; exists {
			return nil, fmt.Errorf("duplicate canonical name %q", name)
		}
		seen[name] = struct{}{}
	}
	return names, nil
}

func loadSelection(path string) (Selection, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Selection{}, fmt.Errorf("read selection manifest: %w", err)
	}
	var selection Selection
	trimmed := bytes.TrimSpace(b)
	if filepath.Ext(path) == ".json" || len(trimmed) > 0 && trimmed[0] == '{' {
		if err := decodeSelectionJSON(trimmed, &selection); err != nil {
			return Selection{}, fmt.Errorf("decode JSON selection manifest: %w", err)
		}
	} else {
		decoder := yaml.NewDecoder(bytes.NewReader(b))
		decoder.KnownFields(true)
		if err := decoder.Decode(&selection); err != nil {
			return Selection{}, fmt.Errorf("decode YAML selection manifest: %w", err)
		}
		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			if err == nil {
				return Selection{}, fmt.Errorf("YAML selection manifest contains more than one document")
			}
			return Selection{}, fmt.Errorf("decode YAML selection manifest: %w", err)
		}
	}
	if selection.SchemaVersion != 1 {
		return Selection{}, fmt.Errorf("unsupported selection schemaVersion %d: want 1", selection.SchemaVersion)
	}
	if len(selection.Names) == 0 {
		return Selection{}, fmt.Errorf("selection names must not be empty")
	}
	return selection, nil
}

func decodeSelectionJSON(b []byte, selection *Selection) error {
	decoder := json.NewDecoder(bytes.NewReader(b))
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '{' {
		return fmt.Errorf("top level must be an object")
	}
	seen := map[string]struct{}{}
	fields := map[string]json.RawMessage{}
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		key, ok := token.(string)
		if !ok {
			return fmt.Errorf("object key must be a string")
		}
		if key != "schemaVersion" && key != "names" {
			return fmt.Errorf("unknown key %q", key)
		}
		if _, exists := seen[key]; exists {
			return fmt.Errorf("duplicate key %q", key)
		}
		seen[key] = struct{}{}
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			return err
		}
		fields[key] = raw
	}
	if _, err := decoder.Token(); err != nil {
		return err
	}
	if len(fields) != 2 {
		return fmt.Errorf("schemaVersion and names are required")
	}
	if err := json.Unmarshal(fields["schemaVersion"], &selection.SchemaVersion); err != nil {
		return fmt.Errorf("schemaVersion: %w", err)
	}
	if err := json.Unmarshal(fields["names"], &selection.Names); err != nil {
		return fmt.Errorf("names: %w", err)
	}
	if token, err := decoder.Token(); err != io.EOF {
		if err != nil {
			return err
		}
		return fmt.Errorf("more than one JSON value (%v)", token)
	}
	return nil
}
