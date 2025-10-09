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

	for _, ud := range []struct {
		u string
		d time.Duration
	}{
		{"y", 365 * 24 * time.Hour},
		{"d", 24 * time.Hour},
	} {
		before, after, found := strings.Cut(s, ud.u)
		if found {
			n, err := strconv.Atoi(before)
			if err != nil {
				return 0, fmt.Errorf("%w: parsing %s: %w", ErrInvalidAge, ud.u, err)
			}
			age += time.Duration(n) * ud.d
			s = after
		}
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
