package rating

func displayIconState(cfg DisplayConfig, value int) string {
	if cfg.isActive(value) {
		return cfg.activeIconClasses()
	}
	return cfg.inactiveIconClasses()
}
