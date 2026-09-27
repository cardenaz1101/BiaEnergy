package engine

import (
	"fmt"
	"math"
	"strings"
	"time"
)

func formatNumber(value float64, decimals int) string {
	negative := value < 0
	text := fmt.Sprintf("%.*f", decimals, math.Abs(value))
	integerPart, fraction, hasFraction := strings.Cut(text, ".")

	var grouped strings.Builder
	for i, digit := range integerPart {
		if i > 0 && (len(integerPart)-i)%3 == 0 {
			grouped.WriteByte('.')
		}
		grouped.WriteRune(digit)
	}
	result := grouped.String()
	if hasFraction {
		result += "," + fraction
	}
	if negative {
		result = "-" + result
	}
	return result
}

func formatSignedPercent(value float64) string {
	sign := "+"
	if value < 0 {
		sign = ""
	}
	return sign + formatNumber(value, 1) + "%"
}

func formatUnsignedPercent(value float64) string {
	return formatNumber(math.Abs(value), 1) + "%"
}

func formatDateTime(t time.Time) string {
	return t.Format("02/01 15:04")
}

func formatDay(t time.Time) string {
	return t.Format("02/01")
}

func capitalize(text string) string {
	if text == "" {
		return text
	}
	return strings.ToUpper(text[:1]) + text[1:]
}
