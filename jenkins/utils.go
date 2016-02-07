package jenkins

import (
	"sort"
	"strings"
)

func joinKeysInCsv(m map[string]bool) string {
	sep := ", "
	if len(m) == 0 {
		return ""
	}
	b := make([]string, len(m))
	bp := 0
	for key := range m {
		b[bp] = key
		bp++
	}
	sort.Strings(b)
	return strings.Join(b, sep)
}
