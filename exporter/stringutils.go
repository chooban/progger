package exporter

import "strings"

func ContainsAll(s string, t []string) bool {
	if len(t) == 0 {
		return true
	}
	for _, v := range t {
		if !strings.Contains(s, v) {
			return false
		}
	}
	return true
}
