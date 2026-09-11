module github.com/araihu/goshtoso/examples/getting-started

go 1.27.0

require (
	github.com/a-h/templ v0.3.1020
	github.com/araihu/goshtoso v0.0.0-00010101000000-000000000000
)

require go.yaml.in/yaml/v3 v3.0.5 // indirect

replace github.com/araihu/goshtoso => ../..
