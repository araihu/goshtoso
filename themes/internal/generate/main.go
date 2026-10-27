// Command generate freezes the source theme declarations and compiled palette.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var comments = regexp.MustCompile(`(?s)/\*.*?\*/`)
var declarations = regexp.MustCompile(`(--[a-zA-Z0-9-]+)\s*:\s*([^;{}]+)(?:;|})`)
var themeBlocks = regexp.MustCompile(`(?ms)^[ \t]*\[data-theme\s*=\s*["']?([a-z0-9-]+)["']?\]\s*\{([^{}]*)\}`)

func read(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return comments.ReplaceAllString(string(b), "")
}
func values(s string) map[string]string {
	out := map[string]string{}
	for _, m := range declarations.FindAllStringSubmatch(s, -1) {
		out[m[1]] = strings.TrimSpace(m[2])
	}
	return out
}
func main() {
	source := read("../all-themes.css")
	before, _, ok := strings.Cut(source, "@layer base")
	if !ok {
		panic("missing theme source boundary")
	}
	defaults := values(before)
	// css/main.css supplies the final root font defaults.
	maps.Copy(defaults, values(read("../css/main.css")))
	overrides := map[string]map[string]string{}
	for _, m := range themeBlocks.FindAllStringSubmatch(source, -1) {
		overrides[m[1]] = values(m[2])
	}
	palette := map[string]string{"--color-black": "#000000", "--color-white": "#ffffff"}
	for _, m := range declarations.FindAllStringSubmatch(read("../assets/styles.css"), -1) {
		k, v := m[1], strings.TrimSpace(m[2])
		if (strings.HasPrefix(k, "--color-") || strings.HasPrefix(k, "--radius-")) && !strings.Contains(v, "var(") && !strings.Contains(v, "color-mix(") {
			if _, ok := palette[k]; !ok {
				palette[k] = v
			}
		}
	}
	files, err := filepath.Glob("*.go")
	if err != nil {
		panic(err)
	}
	hash := sha256.New()
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		data, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}
		hash.Write([]byte(file + "\x00"))
		hash.Write(data)
	}
	payload := struct {
		MaterializerDigest string                       `json:"materializerDigest"`
		Defaults           map[string]string            `json:"defaults"`
		Overrides          map[string]map[string]string `json:"overrides"`
		Palette            map[string]string            `json:"palette"`
	}{hex.EncodeToString(hash.Sum(nil)), defaults, overrides, palette}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		panic(err)
	}
	if err = os.WriteFile("source_catalog.json", append(data, '\n'), 0644); err != nil {
		panic(err)
	}
	fmt.Printf("materialization source: %d presets, %d palette entries\n", len(overrides), len(palette))
}
