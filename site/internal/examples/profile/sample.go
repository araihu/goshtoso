package profile

// Sample returns a starter profile shown on a visitor's first load (no cookie)
// so the example never opens empty.
func Sample() State {
	return State{
		Name: "Ada Lovelace",
		Bio:  "Mathematician and writer, known for work on Babbage's Analytical Engine — arguably the first computer programmer.",
	}
}
