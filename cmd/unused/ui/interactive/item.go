package interactive

import (
	"strings"
	"time"

	"github.com/grafana/unused"
)

type item struct {
	disk    unused.Disk
	verbose bool
}

func (i item) Title() string { return i.disk.Name() }

func (i item) FilterValue() string {
	// TODO this doesn't looks right when filtering
	return i.disk.Name() + " " + i.disk.Meta().String() + " " + i.disk.Provider().Name() + " " + i.disk.Provider().Meta().String()
}

func (i item) Description() string {
	var s strings.Builder

	s.WriteString(i.disk.Provider().Name())
	s.WriteRune('{')
	s.WriteString(i.disk.Provider().Meta().String())
	s.WriteRune('}')

	s.WriteRune(' ')
	s.WriteString("age=")
	s.WriteString(time.Since(i.disk.CreatedAt()).Round(time.Minute).String())

	if i.verbose {
		s.WriteRune(' ')
		s.WriteString(i.disk.Meta().String())
	}

	return s.String()
}
