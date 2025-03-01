package format

import (
	"fmt"
	"time"
)

// FormatDuration formats the duration in a human-readable way
func FormatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24

	if days > 0 {
		if hours > 0 {
			return fmt.Sprintf("%dd %dh", days, hours)
		}
		return fmt.Sprintf("%dd", days)
	}

	if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}

	minutes := int(d.Minutes()) % 60
	if minutes > 0 {
		return fmt.Sprintf("%dm", minutes)
	}

	return "just now"
}
