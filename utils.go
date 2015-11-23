package main
import (
	"log"
	"github.com/jroimartin/gocui"
)

func itoidrune(i int) rune {
	if i < 10 {
		return rune(48 + i)
	}
	return rune(87 + i)
}

func idtoindex(i byte) byte {
	if i >= 87 {
		return i - 87
	}
	if i >= 48 && i <= 57 {
		return i - 48
	}
	log.Fatalf("Not allowed id: %v", i)
	return 255
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
	n := len(sep) * (len(m) - 1)
	for key, _ := range m {
		n += len(key)
	}
	b := make([]byte, n+1)
	bp := 0
	for key, _ := range m {
		bp += copy(b[bp:], key)
		bp += copy(b[bp:], sep)
	}
	return string(b[:bp-1])
}