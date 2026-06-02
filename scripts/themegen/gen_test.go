package main

import (
	"strings"
	"testing"
)

func TestGenerateTheme(t *testing.T) {
	mainCSS := `@custom-variant dark (&:where(.dark, .dark *));
@import "tailwindcss";
@source "../components/**/*.templ";
@source inline("md:w-64");
@import "../all-themes.css";
@import "./codeblock.css";
@theme { --font-body: x; }`

	imports := map[string]string{
		"all-themes.css": "@theme { --color-primary: red; }",
		"codeblock.css":  ".ch-x { color: red; }",
	}
	out := generateTheme(mainCSS, imports)

	if strings.Contains(out, `@import "tailwindcss"`) {
		t.Error("tailwind import must be stripped")
	}
	if strings.Contains(out, `@source "../components`) {
		t.Error("repo path @source globs must be stripped")
	}
	if !strings.Contains(out, `@source inline("md:w-64")`) {
		t.Error("@source inline safelists must be kept")
	}
	if !strings.Contains(out, "--color-primary: red") {
		t.Error("all-themes.css must be inlined")
	}
	if !strings.Contains(out, ".ch-x") {
		t.Error("codeblock.css must be inlined")
	}
	if !strings.Contains(out, "@custom-variant dark") {
		t.Error("@custom-variant must be preserved")
	}
	if !strings.Contains(out, "--font-body: x") {
		t.Error("@theme blocks must be preserved")
	}
	if strings.Contains(out, `@import "../all-themes.css"`) || strings.Contains(out, `@import "./codeblock.css"`) {
		t.Error("relative @imports must be replaced, not left in place")
	}
}
