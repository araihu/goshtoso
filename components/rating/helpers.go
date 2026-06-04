package rating

func readOnlyIconState(cfg Config, value int) string {
	if cfg.IsActive(value) {
		return cfg.ActiveIconClasses()
	}
	return cfg.InactiveIconClasses()
}
