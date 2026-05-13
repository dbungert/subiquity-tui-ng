package ui

import (
	"os"
	"strings"
)

var useHalfBlocks = detectHalfBlockSupport()

func detectHalfBlockSupport() bool {
	switch os.Getenv("SUBIQUITY_NG_HEADER") {
	case "blocks":
		return true
	case "plain":
		return false
	}
	switch os.Getenv("TERM") {
	case "linux", "dumb", "":
		return false
	}
	return isUTF8Locale()
}

func isUTF8Locale() bool {
	for _, k := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		v := os.Getenv(k)
		if v == "" {
			continue
		}
		lv := strings.ToLower(v)
		return strings.Contains(lv, "utf-8") || strings.Contains(lv, "utf8")
	}
	return false
}
