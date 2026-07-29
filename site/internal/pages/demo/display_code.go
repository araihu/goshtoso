package demo

// DisplayCode marks source text rendered by ComponentDemo or DemoSection as an
// escaped example, not executable page markup. It is also a detector precision
// boundary; never pass its result to templ.Raw.
func DisplayCode(source string) string {
	return source
}
