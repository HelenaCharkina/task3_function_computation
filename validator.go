package main

import "strings"

func IsUnknownFieldError(err error) bool {
	if strings.Contains(err.Error(), "unknown field") {
		return true
	}
	return false
}