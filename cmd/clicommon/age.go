package clicommon

import (
	"fmt"
	"time"
)

func Age(date time.Time) string {
	if date.IsZero() {
		return "n/a"
	}

	d := time.Since(date)

	if d <= time.Minute {
		return "1m"
	} else if d < time.Hour {
		return fmt.Sprintf("%dm", d/time.Minute)
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%dh", d/time.Hour)
	} else if d < 365*24*time.Hour {
		return fmt.Sprintf("%dd", d/(24*time.Hour))
	} else {
		return fmt.Sprintf("%dy", d/(365*24*time.Hour))
	}
}
