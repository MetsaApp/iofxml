package iofxml

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

// Namespace is the IOF XML 3.0 namespace every document declares.
const Namespace = "http://www.orienteering.org/datastandard/3.0"

// Version is the iofVersion a 3.0 document carries.
const Version = "3.0"

// DefaultCreator is the creator Marshal writes when a document names none.
const DefaultCreator = "github.com/MetsaApp/iofxml"

// Header is the attributes every document's root carries.
type Header struct {
	// IOFVersion is "3.0"; Marshal writes that when it is empty.
	IOFVersion string `xml:"iofVersion,attr"`
	// CreateTime is when the file was made; Marshal writes the current
	// time when it is empty.
	CreateTime DateTime `xml:"createTime,attr,omitempty"`
	// Creator is the software that made the file; Marshal writes
	// DefaultCreator when it is empty.
	Creator string `xml:"creator,attr,omitempty"`
}

// ID is an identifier with the type its issuer gave it. An ID with no
// value is not written.
type ID struct {
	Type  string `xml:"type,attr,omitempty"`
	Value string `xml:",chardata"`
}

// Key is the id as one string, the type folded in when there is one, so
// two issuers' "17" stay apart. Empty when the id has no value.
func (id ID) Key() string {
	v := strings.TrimSpace(id.Value)
	if v == "" {
		return ""
	}
	if t := strings.TrimSpace(id.Type); t != "" {
		return t + ":" + v
	}
	return v
}

// MarshalXML leaves out an id with no value: the standard's ids are all
// optional, and an empty element is no id.
func (id ID) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if strings.TrimSpace(id.Value) == "" {
		return nil
	}
	type plain ID
	return e.EncodeElement(plain(id), start) //nolint:wrapcheck // The encoder's error, as any MarshalXML returns it.
}

// Event is the event a list is for.
type Event struct {
	Name string `xml:"Name"`
	// StartTime is the event's own date, not a competitor's start.
	StartTime *DateAndOptionalTime `xml:"StartTime,omitempty"`
}

// DateAndOptionalTime is a calendar date, "2026-09-01", with an optional
// time of day, "10:00:00+02:00".
type DateAndOptionalTime struct {
	Date string `xml:"Date"`
	Time string `xml:"Time,omitempty"`
}

// DateTime is an instant as the document writes it, an xs:dateTime such
// as "2026-09-01T10:38:00+02:00". It stays text so that an instant
// without a zone keeps that fact; Time reads it in a zone.
type DateTime string

// NewDateTime is t as a DateTime, with its zone offset.
func NewDateTime(t time.Time) DateTime {
	return DateTime(t.Format(time.RFC3339Nano))
}

// localLayout is an xs:dateTime without a zone.
const localLayout = "2006-01-02T15:04:05.999999999"

// Time reads the instant. One without a zone, which the standard allows,
// is read in loc. A time of day without a date is an error: package iof2
// writes one for a 2.0.3 start that names no date, and only the caller
// knows the day.
func (d DateTime) Time(loc *time.Location) (time.Time, error) {
	s := strings.TrimSpace(string(d))
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t, nil
	}
	t, err := time.ParseInLocation(localLayout, s, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("iofxml: date and time %q: %w", s, err)
	}
	return t, nil
}

// PersonName is the standard's two-part name, family first as its
// sequence requires.
type PersonName struct {
	Family string `xml:"Family"`
	Given  string `xml:"Given"`
}

// Full is the given name then the family name, trimmed, the order MeOS
// writes and result lists print.
func (n PersonName) Full() string {
	return strings.TrimSpace(strings.TrimSpace(n.Given) + " " + strings.TrimSpace(n.Family))
}

// Person is a competitor.
type Person struct {
	// IDs is every id the person has; most documents carry one.
	IDs         []ID       `xml:"Id"`
	Name        PersonName `xml:"Name"`
	Nationality *Country   `xml:"Nationality,omitempty"`
}

// ID is the person's first id, the zero ID when there is none.
func (p Person) ID() ID {
	if len(p.IDs) == 0 {
		return ID{}
	}
	return p.IDs[0]
}

// Country is a country by its code, "DNK", with its name as text.
type Country struct {
	Code string `xml:"code,attr"`
	Name string `xml:",chardata"`
}

// Organisation is a club, or any body a competitor runs for.
type Organisation struct {
	ID      ID       `xml:"Id"`
	Name    string   `xml:"Name"`
	Country *Country `xml:"Country,omitempty"`
}

// ControlCard is a card number, with the punching system when the
// document names one ("SI", "Emit").
type ControlCard struct {
	PunchingSystem string `xml:"punchingSystem,attr,omitempty"`
	Value          string `xml:",chardata"`
}

// Cards is the card numbers that are not blank, trimmed.
func Cards(cards []ControlCard) []string {
	out := make([]string, 0, len(cards))
	for _, c := range cards {
		if v := strings.TrimSpace(c.Value); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// Class is a class header. Legs is the relay legs the class declares,
// empty for an individual class.
type Class struct {
	ID        ID     `xml:"Id"`
	Name      string `xml:"Name"`
	ShortName string `xml:"ShortName,omitempty"`
	Legs      []Leg  `xml:"Leg"`
}

// Leg is one declared relay leg.
type Leg struct {
	Name string `xml:"Name,omitempty"`
}

// Course is a class's course as a start or a result list names it.
// Length and climb are in metres; zero is not written.
type Course struct {
	ID     ID      `xml:"Id"`
	Name   string  `xml:"Name,omitempty"`
	Length float64 `xml:"Length,omitempty"`
	Climb  float64 `xml:"Climb,omitempty"`
}

// AssignedFee is a fee charged for an entry.
type AssignedFee struct {
	Fee Fee `xml:"Fee"`
}

// Fee is what a fee is for and what it costs. The standard requires the
// name.
type Fee struct {
	Name   string  `xml:"Name"`
	Amount *Amount `xml:"Amount,omitempty"`
}

// Amount is a sum of money. Value stays text, "120.00", so a decimal is
// written as given and never through a float.
type Amount struct {
	Currency string `xml:"currency,attr,omitempty"`
	Value    string `xml:",chardata"`
}
