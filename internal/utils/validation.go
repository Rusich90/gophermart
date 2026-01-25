package utils

import (
	"strconv"
	"strings"
)

func IsValidLuhn(s string) bool {
	s = strings.ReplaceAll(s, " ", "")
	if _, err := strconv.Atoi(s); err != nil || len(s) < 2 {
		return false
	}

	total := 0
	double := false

	for i := len(s) - 1; i >= 0; i-- {
		digit := int(s[i] - '0')
		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		total += digit
		double = !double
	}

	return total%10 == 0
}
