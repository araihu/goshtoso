package iconlibrary

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	_ "golang.org/x/image/webp"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"strings"
)

const MaxImageBytes = 2 << 20
const MaxImageDimension = 4096

// ValidateImage checks actual bytes, not an extension or a supplied content type.
// SVG is deliberately restricted to passive vector markup. Rejected files must
// be converted by their owner, never silently treated as trusted markup.
func ValidateImage(data []byte) (mime string, width, height int, err error) {
	if len(data) == 0 || len(data) > MaxImageBytes {
		return "", 0, 0, errors.New("icon must be between 1 byte and 2 MiB")
	}
	if c, format, e := image.DecodeConfig(bytes.NewReader(data)); e == nil {
		if c.Width <= 0 || c.Height <= 0 || c.Width > MaxImageDimension || c.Height > MaxImageDimension {
			return "", 0, 0, errors.New("icon dimensions must not exceed 4096 pixels")
		}
		if _, _, e = image.Decode(bytes.NewReader(data)); e != nil {
			return "", 0, 0, errors.New("invalid image data")
		}
		return "image/" + format, c.Width, c.Height, nil
	}
	if err = validateSVG(data); err != nil {
		return "", 0, 0, err
	}
	return "image/svg+xml", 0, 0, nil
}

func validateSVG(data []byte) error {
	d := xml.NewDecoder(bytes.NewReader(data))
	depth, roots := 0, 0
	allowed := map[string]bool{}
	for name := range strings.FieldsSeq("svg g path rect circle ellipse line polyline polygon defs linearGradient radialGradient stop clipPath mask title desc use symbol") {
		allowed[name] = true
	}
	for {
		tok, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.New("invalid SVG")
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if depth == 0 {
				roots++
				if err := validateSVGRoot(t, roots); err != nil {
					return err
				}
			}
			if !allowed[t.Name.Local] || t.Name.Space != "http://www.w3.org/2000/svg" {
				return fmt.Errorf("unsupported SVG element %q", t.Name.Local)
			}
			if err := validateSVGImageAttributes(t); err != nil {
				return err
			}
			depth++
		case xml.EndElement:
			depth--
		case xml.Directive:
			return errors.New("SVG directives are not allowed")
		case xml.ProcInst:
			if t.Target != "xml" {
				return errors.New("SVG processing instructions are not allowed")
			}
		case xml.CharData:
			if depth == 0 && strings.TrimSpace(string(t)) != "" {
				return errors.New("text outside SVG root")
			}
		}
	}
	if roots != 1 || depth != 0 {
		return errors.New("incomplete SVG")
	}
	return nil
}

func validateSVGImageAttributes(t xml.StartElement) error {
	for _, a := range t.Attr {
		name, value := strings.ToLower(a.Name.Local), strings.ToLower(strings.TrimSpace(a.Value))
		if strings.HasPrefix(name, "on") || name == "style" || name == "base" || strings.Contains(value, "url(") || strings.Contains(value, "javascript:") {
			return errors.New("active SVG content is not allowed")
		}
		if (name == "href" || name == "src") && !strings.HasPrefix(value, "#") {
			return errors.New("external SVG references are not allowed")
		}
	}
	return nil
}

func validateSVGRoot(t xml.StartElement, roots int) error {
	if roots != 1 || t.Name.Local != "svg" || t.Name.Space != "http://www.w3.org/2000/svg" {
		return errors.New("expected SVG root")
	}
	return nil
}
