package iconpack

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func buildSprite(boundary releaseBoundary, selected []selectedAsset, families []sourceFamily) ([]byte, error) {
	wantedByFamily := map[string]map[string]string{}
	for _, asset := range selected {
		key := asset.family.Namespace + "/" + asset.family.Product
		if wantedByFamily[key] == nil {
			wantedByFamily[key] = map[string]string{}
		}
		wantedByFamily[key][asset.SpriteSymbol] = asset.Dimensions.ViewBox
	}
	extracted := map[string][]byte{}
	for _, family := range families {
		key := family.Namespace + "/" + family.Product
		got, err := extractFamilySymbols(boundary.root, family.SpritePath, wantedByFamily[key])
		if err != nil {
			return nil, err
		}
		for symbol, raw := range got {
			if _, exists := extracted[symbol]; exists {
				return nil, fmt.Errorf("duplicate selected sprite symbol %q across source sprites", symbol)
			}
			extracted[symbol] = raw
		}
	}
	var output bytes.Buffer
	output.WriteString(`<svg xmlns="http://www.w3.org/2000/svg">`)
	output.WriteByte('\n')
	for _, asset := range selected {
		raw, ok := extracted[asset.SpriteSymbol]
		if !ok {
			return nil, fmt.Errorf("exact sprite symbol %q for %q was not found", asset.SpriteSymbol, asset.CanonicalName)
		}
		output.Write(raw)
		output.WriteByte('\n')
	}
	output.WriteString("</svg>\n")
	return output.Bytes(), nil
}

func extractFamilySymbols(root, relative string, wanted map[string]string) (map[string][]byte, error) {
	filename := releasePath(root, relative)
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read source sprite %q: %w", relative, err)
	}
	scanner := spriteScanner{
		relative: relative,
		raw:      b,
		decoder:  xml.NewDecoder(bytes.NewReader(b)),
		wanted:   wanted,
		found:    map[string][]byte{},
		seen:     map[string]struct{}{},
	}
	if err := scanner.scan(); err != nil {
		return nil, err
	}
	return scanner.found, nil
}

type spriteScanner struct {
	relative string
	raw      []byte
	decoder  *xml.Decoder
	wanted   map[string]string
	found    map[string][]byte
	seen     map[string]struct{}
	depth    int
	rootSeen bool
}

func (scanner *spriteScanner) scan() error {
	for {
		startOffset := scanner.decoder.InputOffset()
		token, err := scanner.decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("parse source sprite %q: %w", scanner.relative, err)
		}
		if err := scanner.processToken(token, startOffset); err != nil {
			return err
		}
	}
	if !scanner.rootSeen || scanner.depth != 0 {
		return fmt.Errorf("source sprite %q has incomplete SVG root", scanner.relative)
	}
	for symbol := range scanner.wanted {
		if _, ok := scanner.found[symbol]; !ok {
			return fmt.Errorf("source sprite %q is missing exact catalog spriteSymbol %q", scanner.relative, symbol)
		}
	}
	return nil
}

func (scanner *spriteScanner) processToken(token xml.Token, startOffset int64) error {
	switch value := token.(type) {
	case xml.Directive:
		return fmt.Errorf("source sprite %q contains forbidden XML directive", scanner.relative)
	case xml.StartElement:
		return scanner.processStart(value, startOffset)
	case xml.EndElement:
		scanner.depth--
		if scanner.depth < 0 {
			return fmt.Errorf("source sprite %q has unbalanced XML", scanner.relative)
		}
	}
	return nil
}

func (scanner *spriteScanner) processStart(start xml.StartElement, startOffset int64) error {
	if scanner.depth == 0 {
		if scanner.rootSeen || start.Name.Space != "http://www.w3.org/2000/svg" || start.Name.Local != "svg" {
			return fmt.Errorf("source sprite %q has invalid root element", scanner.relative)
		}
		scanner.rootSeen = true
		scanner.depth++
		return nil
	}
	if scanner.depth != 1 || start.Name.Space != "http://www.w3.org/2000/svg" || start.Name.Local != "symbol" {
		scanner.depth++
		return nil
	}
	id, viewBox, err := symbolAttributes(start)
	if err != nil {
		return fmt.Errorf("source sprite %q: %w", scanner.relative, err)
	}
	if _, exists := scanner.seen[id]; exists {
		return fmt.Errorf("source sprite %q has duplicate symbol %q", scanner.relative, id)
	}
	scanner.seen[id] = struct{}{}
	if err := consumeElement(scanner.decoder); err != nil {
		return fmt.Errorf("parse source sprite symbol %q: %w", id, err)
	}
	expectedViewBox, wanted := scanner.wanted[id]
	if !wanted {
		return nil
	}
	if viewBox != expectedViewBox {
		return fmt.Errorf("sprite symbol %q viewBox %q does not match catalog %q", id, viewBox, expectedViewBox)
	}
	scanner.found[id] = append([]byte(nil), scanner.raw[startOffset:scanner.decoder.InputOffset()]...)
	return nil
}

func symbolAttributes(start xml.StartElement) (string, string, error) {
	var id, viewBox string
	for _, attr := range start.Attr {
		if attr.Name.Space == "" && attr.Name.Local == "id" {
			if id != "" {
				return "", "", fmt.Errorf("symbol has duplicate id attribute")
			}
			id = attr.Value
		}
		if attr.Name.Space == "" && attr.Name.Local == "viewBox" {
			if viewBox != "" {
				return "", "", fmt.Errorf("symbol %q has duplicate viewBox attribute", id)
			}
			viewBox = attr.Value
		}
	}
	if id == "" || viewBox == "" {
		return "", "", fmt.Errorf("symbol must have id and viewBox")
	}
	return id, viewBox, nil
}

func consumeElement(decoder *xml.Decoder) error {
	depth := 1
	for depth > 0 {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		switch token.(type) {
		case xml.Directive:
			return fmt.Errorf("forbidden XML directive")
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		}
	}
	return nil
}

func releasePath(root, relative string) string {
	return filepath.Join(root, filepath.FromSlash(relative))
}
