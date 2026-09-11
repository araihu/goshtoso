package docspages

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/araihu/goshtoso/components/schematree"
	"github.com/araihu/goshtoso/expressions"
)

type expressionSchemaProperty struct {
	Examples    []string                            `json:"examples"`
	Type        string                              `json:"type"`
	Description string                              `json:"description"`
	Default     *string                             `json:"default"`
	Parameter   string                              `json:"x-parameter"`
	Properties  map[string]expressionSchemaProperty `json:"properties"`
}

func expressionSchemaNodes() []schematree.Node {
	var schema expressionSchemaProperty
	if err := json.Unmarshal(expressions.JSONSchema(), &schema); err != nil {
		panic(fmt.Errorf("embedded expression schema: %w", err))
	}
	return expressionPropertyNodes(schema.Properties, "")
}

func expressionPropertyNodes(properties map[string]expressionSchemaProperty, prefix string) []schematree.Node {
	names := make([]string, 0, len(properties))
	for name := range properties {
		names = append(names, name)
	}
	slices.Sort(names)
	nodes := make([]schematree.Node, 0, len(names))
	for _, name := range names {
		property := properties[name]
		path := name
		if prefix != "" {
			path = prefix + "." + name
		}
		node := schematree.Node{Name: name, Path: path, Type: property.Type, Description: property.Description, ShowOptional: true, Collapsed: len(property.Properties) > 0}
		if property.Default != nil {
			node.Constraints = append(node.Constraints, schematree.Constraint{Name: "English", Value: *property.Default})
		}
		for _, example := range property.Examples {
			node.Constraints = append(node.Constraints, schematree.Constraint{Name: "English message", Value: example})
		}
		if property.Parameter != "" {
			node.Constraints = append(node.Constraints, schematree.Constraint{Name: "Placeholder", Value: "{" + property.Parameter + "}"})
		}
		node.Children = expressionPropertyNodes(property.Properties, path)
		nodes = append(nodes, node)
	}
	return nodes
}
