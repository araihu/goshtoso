package tooltip

func tooltipTriggerLabel(cfg config, inherited string) string {
	if cfg.triggerLabelSet {
		return cfg.triggerLabel
	}
	return inherited
}
