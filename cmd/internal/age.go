package internal

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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

var ErrInvalidAge = errors.New("invalid age")

func ParseAge(s string) (time.Duration, error) {
	if s == "" {
		return 0, ErrInvalidAge
	}

	var age time.Duration

	days, s, ok := strings.Cut(s, "d")
	if ok {
		days, err := strconv.Atoi(days)
		if err != nil {
			return 0, fmt.Errorf("%w: %w", ErrInvalidAge, err)
		}

		age = time.Duration(days) * 24 * time.Hour
	}

	if s != "" {
		dur, err := time.ParseDuration(s)
		if err != nil {
			return 0, fmt.Errorf("%w: %w", ErrInvalidAge, err)
		}

		age += dur
	}

	return age, nil
}
