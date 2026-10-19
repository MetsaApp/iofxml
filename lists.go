package iofxml

import "strings"

// The documents. Field order is the schema's sequence, which a validating
// reader enforces: a document with every element present but in another
// order fails validation. The same holds for every type below.

// EntryList is the entries for an event, written before the start list
// is drawn.
type EntryList struct {
	Header

	Event         Event         `xml:"Event"`
	TeamEntries   []TeamEntry   `xml:"TeamEntry"`
	PersonEntries []PersonEntry `xml:"PersonEntry"`
}

// PersonEntry is one person's entry.
type PersonEntry struct {
	ID           ID            `xml:"Id"`
	Person       Person        `xml:"Person"`
	Organisation *Organisation `xml:"Organisation,omitempty"`
	ControlCards []ControlCard `xml:"ControlCard"`
	Classes      []Class       `xml:"Class"`
	AssignedFees []AssignedFee `xml:"AssignedFee"`
}

// TeamEntry is one relay team's entry.
type TeamEntry struct {
	ID            ID                `xml:"Id"`
	Name          string            `xml:"Name"`
	Organisations []Organisation    `xml:"Organisation"`
	Members       []TeamEntryPerson `xml:"TeamEntryPerson"`
	Classes       []Class           `xml:"Class"`
	AssignedFees  []AssignedFee     `xml:"AssignedFee"`
}

// TeamEntryPerson is a team member and the leg they run. Person is nil
// for a leg nobody is named for yet.
type TeamEntryPerson struct {
	Person       *Person       `xml:"Person,omitempty"`
	Organisation *Organisation `xml:"Organisation,omitempty"`
	Leg          int           `xml:"Leg,omitempty"`
	LegOrder     int           `xml:"LegOrder,omitempty"`
	ControlCards []ControlCard `xml:"ControlCard"`
}

// StartList is the start times of an event, class by class.
type StartList struct {
	Header

	Event       Event        `xml:"Event"`
	ClassStarts []ClassStart `xml:"ClassStart"`
}

// ClassStart is one class's part of a start list.
type ClassStart struct {
	Class        Class         `xml:"Class"`
	Courses      []Course      `xml:"Course"`
	PersonStarts []PersonStart `xml:"PersonStart"`
	TeamStarts   []TeamStart   `xml:"TeamStart"`
}

// PersonStart is one person's start, one Start per race.
type PersonStart struct {
	EntryID      ID                `xml:"EntryId"`
	Person       *Person           `xml:"Person,omitempty"`
	Organisation *Organisation     `xml:"Organisation,omitempty"`
	Starts       []PersonRaceStart `xml:"Start"`
}

// PersonRaceStart is a person's start in one race.
type PersonRaceStart struct {
	BibNumber    string        `xml:"BibNumber,omitempty"`
	StartTime    DateTime      `xml:"StartTime,omitempty"`
	ControlCards []ControlCard `xml:"ControlCard"`
}

// TeamStart is a relay team's start.
type TeamStart struct {
	EntryID       ID                `xml:"EntryId"`
	Name          string            `xml:"Name,omitempty"`
	Organisations []Organisation    `xml:"Organisation"`
	BibNumber     string            `xml:"BibNumber,omitempty"`
	Members       []TeamMemberStart `xml:"TeamMemberStart"`
}

// TeamMemberStart is a team member's start, one Start per race.
type TeamMemberStart struct {
	EntryID      ID                    `xml:"EntryId"`
	Person       *Person               `xml:"Person,omitempty"`
	Organisation *Organisation         `xml:"Organisation,omitempty"`
	Starts       []TeamMemberRaceStart `xml:"Start"`
}

// TeamMemberRaceStart is a team member's start in one race, with the leg
// they run.
type TeamMemberRaceStart struct {
	Leg          int           `xml:"Leg,omitempty"`
	LegOrder     int           `xml:"LegOrder,omitempty"`
	BibNumber    string        `xml:"BibNumber,omitempty"`
	StartTime    DateTime      `xml:"StartTime,omitempty"`
	ControlCards []ControlCard `xml:"ControlCard"`
}

// ResultList is the results of an event, class by class.
type ResultList struct {
	Header

	// Status is the list's status as written; EffectiveStatus applies the
	// standard's default.
	Status       ListStatus    `xml:"status,attr,omitempty"`
	Event        Event         `xml:"Event"`
	ClassResults []ClassResult `xml:"ClassResult"`
}

// EffectiveStatus is the list's status, Complete when it declares none.
func (l ResultList) EffectiveStatus() ListStatus {
	if s := ListStatus(strings.TrimSpace(string(l.Status))); s != "" {
		return s
	}
	return ListComplete
}

// ClassResult is one class's part of a result list.
type ClassResult struct {
	Class         Class          `xml:"Class"`
	Courses       []Course       `xml:"Course"`
	PersonResults []PersonResult `xml:"PersonResult"`
	TeamResults   []TeamResult   `xml:"TeamResult"`
}

// PersonResult is one person's result, one Result per race.
type PersonResult struct {
	EntryID      ID                 `xml:"EntryId"`
	Person       Person             `xml:"Person"`
	Organisation *Organisation      `xml:"Organisation,omitempty"`
	Results      []PersonRaceResult `xml:"Result"`
}

// PersonRaceResult is a person's result in one race. Time is the running
// time in seconds, nil when there is none.
type PersonRaceResult struct {
	BibNumber  string   `xml:"BibNumber,omitempty"`
	StartTime  DateTime `xml:"StartTime,omitempty"`
	FinishTime DateTime `xml:"FinishTime,omitempty"`
	Time       *float64 `xml:"Time,omitempty"`
	TimeBehind *float64 `xml:"TimeBehind,omitempty"`
	// Position is the place in the class, 0 when there is none.
	Position      int            `xml:"Position,omitempty"`
	Status        ResultStatus   `xml:"Status"`
	OverallResult *OverallResult `xml:"OverallResult,omitempty"`
	SplitTimes    []SplitTime    `xml:"SplitTime"`
	ControlCards  []ControlCard  `xml:"ControlCard"`
}

// TeamResult is a relay team's result.
type TeamResult struct {
	EntryID       ID                 `xml:"EntryId"`
	Name          string             `xml:"Name"`
	Organisations []Organisation     `xml:"Organisation"`
	BibNumber     string             `xml:"BibNumber,omitempty"`
	Members       []TeamMemberResult `xml:"TeamMemberResult"`
}

// TeamMemberResult is a team member's result, one Result per race.
// Person is nil for a leg nobody is named for.
type TeamMemberResult struct {
	EntryID      ID                     `xml:"EntryId"`
	Person       *Person                `xml:"Person,omitempty"`
	Organisation *Organisation          `xml:"Organisation,omitempty"`
	Results      []TeamMemberRaceResult `xml:"Result"`
}

// TeamMemberRaceResult is a team member's result in one race: the leg
// they ran and, in OverallResult, the team's result through that leg.
type TeamMemberRaceResult struct {
	Leg        int          `xml:"Leg,omitempty"`
	LegOrder   int          `xml:"LegOrder,omitempty"`
	BibNumber  string       `xml:"BibNumber,omitempty"`
	StartTime  DateTime     `xml:"StartTime,omitempty"`
	FinishTime DateTime     `xml:"FinishTime,omitempty"`
	Time       *float64     `xml:"Time,omitempty"`
	TimeBehind []TypedTime  `xml:"TimeBehind"`
	Positions  []TypedPlace `xml:"Position"`
	Status     ResultStatus `xml:"Status"`
	// OverallResult is the team's result through this leg.
	OverallResult *OverallResult `xml:"OverallResult,omitempty"`
	SplitTimes    []SplitTime    `xml:"SplitTime"`
	ControlCards  []ControlCard  `xml:"ControlCard"`
}

// Position is the member's place of the given type, PlaceLeg or
// PlaceCourse, and whether the result names one. The schema requires the
// type, but some exporters leave it out; a place without a type is
// returned when none has the type asked for.
func (r *TeamMemberRaceResult) Position(placeType string) (int, bool) {
	untyped, found := 0, false
	for _, p := range r.Positions {
		switch t := strings.TrimSpace(p.Type); {
		case strings.EqualFold(t, placeType):
			return p.Value, true
		case t == "" && !found:
			untyped, found = p.Value, true
		}
	}
	return untyped, found
}

// The types of a team member's place and time behind: among everyone on
// the same leg, or among everyone on the same course.
const (
	PlaceLeg    = "Leg"
	PlaceCourse = "Course"
)

// TypedPlace is a team member's place of one type.
type TypedPlace struct {
	Type  string `xml:"type,attr,omitempty"`
	Value int    `xml:",chardata"`
}

// TypedTime is a team member's time behind of one type, in seconds.
type TypedTime struct {
	Type  string  `xml:"type,attr,omitempty"`
	Value float64 `xml:",chardata"`
}

// OverallResult is a cumulative result: a team's through a leg, or a
// person's over several races.
type OverallResult struct {
	Time       *float64     `xml:"Time,omitempty"`
	TimeBehind *float64     `xml:"TimeBehind,omitempty"`
	Position   int          `xml:"Position,omitempty"`
	Status     ResultStatus `xml:"Status"`
}

// SplitTime is one control passing: the control's code and the seconds
// from the start. Time is nil when the status is Missing.
type SplitTime struct {
	Status      SplitStatus `xml:"status,attr,omitempty"`
	ControlCode string      `xml:"ControlCode"`
	Time        *float64    `xml:"Time,omitempty"`
}
