package expressions

import (
	"fmt"
	"reflect"
	"strings"
)

type messagePart struct {
	text     string
	argument int
}

// parseMessage deliberately supports substitution only: no evaluation or plural rules.
func parseMessage(text string, parameters ...string) ([]messagePart, error) {
	if len(parameters) == 0 || parameters[0] == "" {
		return nil, fmt.Errorf("missing placeholder definition")
	}
	var parts []messagePart
	var literal strings.Builder
	flush := func() {
		if literal.Len() > 0 {
			parts = append(parts, messagePart{text: literal.String()})
			literal.Reset()
		}
	}
	for i := 0; i < len(text); {
		c := text[i]
		if c != '{' && c != '}' {
			literal.WriteByte(c)
			i++
			continue
		}
		if i+1 < len(text) && text[i+1] == c {
			literal.WriteByte(c)
			i += 2
			continue
		}
		matched := false
		for index, parameter := range parameters {
			token := "{" + parameter + "}"
			if strings.HasPrefix(text[i:], token) {
				flush()
				parts = append(parts, messagePart{argument: index + 1})
				i += len(token)
				matched = true
				break
			}
		}
		if !matched {
			return nil, fmt.Errorf("invalid placeholder at byte %d; use %s, {{, or }}", i, strings.Join(parameters, ", "))
		}
	}
	flush()
	return parts, nil
}

func messageFunc(typ reflect.Type, parts []messagePart) reflect.Value {
	return reflect.MakeFunc(typ, func(args []reflect.Value) []reflect.Value {
		var out strings.Builder
		for _, part := range parts {
			if part.argument > 0 {
				_, _ = fmt.Fprint(&out, args[part.argument-1].Interface())
			} else {
				out.WriteString(part.text)
			}
		}
		return []reflect.Value{reflect.ValueOf(out.String())}
	})
}
