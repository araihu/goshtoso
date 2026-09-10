package form

import "encoding/json"

// validationValues encodes inert request metadata without evaluating JavaScript.
func (c FieldGroupConfig) validationValues() string {
	field := ""
	if c.Meta != nil {
		field = c.Meta.FieldName
	}
	values, _ := json.Marshal(map[string]string{
		"X-Goshtoso-Validation": "field",
		"X-Goshtoso-Field":      field,
	})
	return string(values)
}
