package quota

// CrossedThreshold only alerts on a known transition, never on startup or missing data.
func CrossedThreshold(previous, current []Window, threshold int) bool {
	for _, next := range current {
		for _, old := range previous {
			if old.Minutes == next.Minutes && old.Remaining >= threshold && next.Remaining < threshold {
				return true
			}
		}
	}
	return false
}
