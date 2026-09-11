package expressions

import (
	"fmt"
	"reflect"
	"strings"
)

type messagePart struct {
	text     string
	argument bool
}

// parseMessage deliberately supports substitution only: no evaluation or plural rules.
func parseMessage(text, parameter string) ([]messagePart, error) {
	if parameter == "" {
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
		token := "{" + parameter + "}"
		if strings.HasPrefix(text[i:], token) {
			flush()
			parts = append(parts, messagePart{argument: true})
			i += len(token)
			continue
		}
		return nil, fmt.Errorf("invalid placeholder at byte %d; use %s, {{, or }}", i, token)
	}
	flush()
	return parts, nil
}

func messageFunc(typ reflect.Type, parts []messagePart) reflect.Value {
	return reflect.MakeFunc(typ, func(args []reflect.Value) []reflect.Value {
		var out strings.Builder
		argument := fmt.Sprint(args[0].Interface())
		for _, part := range parts {
			if part.argument {
				out.WriteString(argument)
			} else {
				out.WriteString(part.text)
			}
		}
		return []reflect.Value{reflect.ValueOf(out.String())}
	})
}
