package iof2

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/MetsaApp/iofxml"
)

var errTimeFormat = errors.New("is not MM:SS or HH:MM:SS")

const (
	timeParts        = 3 // HH:MM:SS
	secondsPerMinute = 60
	metresPerKm      = 1000
	metresPerFoot    = 0.3048
)

// seconds reads a 2.0.3 Time: "MM:SS" or "HH:MM:SS", the hours or minutes
// unpadded ("1:23:03", "2:21"), tenths after the seconds ("27:52.5"), or
// bare seconds. The number of parts decides, whatever timeFormat says:
// exporters write the attribute carelessly. It only settles a two-part
// value written as "HH:MM". Empty or dashes is no time, the way several
// exporters write a missing one.
func (t *duration) seconds() (*float64, error) {
	if t == nil {
		return nil, nil //nolint:nilnil // No element is no time, as a nil Time is in 3.0.
	}
	s := strings.TrimSpace(t.Value)
	if s == "" || strings.Trim(s, "-") == "" {
		return nil, nil //nolint:nilnil // An empty Time is no time.
	}
	parts := strings.Split(s, ":")
	if len(parts) > timeParts {
		return nil, fmt.Errorf("time %q %w", s, errTimeFormat)
	}
	var secs float64
	for _, part := range parts {
		v, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil || v < 0 {
			return nil, fmt.Errorf("time %q %w", s, errTimeFormat)
		}
		secs = secs*secondsPerMinute + v
	}
	if len(parts) == timeParts-1 && strings.EqualFold(strings.TrimSpace(t.Format), "HH:MM") {
		secs *= secondsPerMinute
	}
	return &secs, nil
}

// dateTime is an instant as 3.0 writes it, "YYYY-MM-DDTHH:MM:SS", when
// the document dates it. 2.0.3 has no zone, so neither has the result:
// the instant is in the event's local time. A start without a date is
// the clock alone, "HH:MM:SS", a time of day on the event's day, which
// only the caller knows. Empty when there is no clock.
func (t *instant) dateTime() iofxml.DateTime {
	if t == nil {
		return ""
	}
	c := clock(t.Clock)
	if c == "" {
		return ""
	}
	if d := strings.TrimSpace(t.Date); d != "" {
		return iofxml.DateTime(d + "T" + c)
	}
	return iofxml.DateTime(c)
}

// clock is a time of day with its seconds, which 2.0.3 lets an exporter
// leave out ("10:38").
func clock(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	for strings.Count(s, ":") < 2 {
		s += ":00"
	}
	return s
}

// courses is the class's course once a row names a CourseLength, in
// metres; the first row's is kept. 2.0.3 puts the length on each row,
// 3.0 on the class.
func courses(have []iofxml.Course, l *length) []iofxml.Course {
	if len(have) > 0 || l == nil {
		return have
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(l.Value), 64)
	if err != nil || v <= 0 {
		return have
	}
	switch strings.ToLower(strings.TrimSpace(l.Unit)) {
	case "km":
		v *= metresPerKm
	case "ft":
		v *= metresPerFoot
	}
	return []iofxml.Course{{Length: v}}
}

// pick is the first of two strings that is not blank.
func pick(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

// atoi is a 2.0.3 number, 0 when blank or not one: a position or a leg
// the exporter left odd must not fail the whole document.
func atoi(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}
