package jenkins

func mapKeysToSlice(m map[string]bool) (b []string) {
	if len(m) == 0 {
		return nil
	}
	b = make([]string, len(m))
	bp := 0
	for key := range m {
		b[bp] = key
		bp++
	}
	return
}
