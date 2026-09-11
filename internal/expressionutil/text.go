package expressionutil

// Text selects an explicit component value before its inherited expression.
func Text(explicit, inherited string) string {
	if explicit != "" {
		return explicit
	}
	return inherited
}
