package wizard

// Plan describes one selectable subscription tier. PriceCents is an integer
// (no floats in money) and is formatted for display in the templ layer.
type Plan struct {
	Key        string
	Name       string
	PriceCents int
	Blurb      string
}

// Plans is the canonical, ordered list of selectable tiers.
var Plans = []Plan{
	{Key: "free", Name: "Free", PriceCents: 0, Blurb: "For trying things out. One project."},
	{Key: "pro", Name: "Pro", PriceCents: 1200, Blurb: "For individuals shipping real work."},
	{Key: "team", Name: "Team", PriceCents: 4000, Blurb: "Shared workspaces and seats."},
}

// PlanByKey returns the Plan for a key and whether it was found.
func PlanByKey(key string) (Plan, bool) {
	for _, p := range Plans {
		if p.Key == key {
			return p, true
		}
	}
	return Plan{}, false
}

// ValidPlan reports whether key names a real tier.
func ValidPlan(key string) bool {
	_, ok := PlanByKey(key)
	return ok
}
