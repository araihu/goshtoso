// internal/examples/expense/format.go
package expense

import "strconv"

// FormatCents renders an integer count of cents as a "$1,234.56" string. It is
// the single money formatter shared by the row, the total badge, and tests, so
// display never reintroduces floating point.
func FormatCents(cents int) string {
	neg := cents < 0
	if neg {
		cents = -cents
	}
	dollars := cents / 100
	frac := cents % 100

	// Group the integer part with thousands separators.
	digits := strconv.Itoa(dollars)
	var grouped []byte
	for i, d := range []byte(digits) {
		if i > 0 && (len(digits)-i)%3 == 0 {
			grouped = append(grouped, ',')
		}
		grouped = append(grouped, d)
	}

	out := "$" + string(grouped) + "." + twoDigits(frac)
	if neg {
		return "-" + out
	}
	return out
}

// twoDigits zero-pads a value in [0,99] to two characters.
func twoDigits(n int) string {
	if n < 10 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}
