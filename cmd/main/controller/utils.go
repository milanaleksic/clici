package controller

import (
	"sort"
	"strings"
)

func joinInCSV(m []string) string {
	sep := ", "
	if len(m) == 0 {
		return ""
	}
	b := make([]string, len(m))
	bp := 0
	for _, user := range m {
		b[bp] = user
		bp++
	}
	sort.Strings(b)
	return strings.Join(b, sep)
}
