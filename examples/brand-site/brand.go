package main

import _ "embed"

//go:embed brand.css
var brandCSS string

func BrandCSS() string { return brandCSS }
