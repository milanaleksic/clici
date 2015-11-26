package main

import (
	"github.com/jroimartin/gocui"
	"log"
	"sort"
	"strings"
)

func itoidrune(i int) rune {
	if i < 10 {
		return rune(48 + i)
	}
	return rune(87 + i)
}

func check_cui(err error) {
	if err != gocui.ErrorUnkView {
		log.Panicf("Unexpected error occured: %v", err)
	}
}

func joinKeysInCsv(m map[string]bool) string {
	sep := ", "
	if len(m) == 0 {
		return ""
	}
	b := make([]string, len(m))
	bp := 0
	for key, _ := range m {
		b[bp] = key
		bp += 1
	}
	sort.Strings(b)
	return strings.Join(b, sep)
}
