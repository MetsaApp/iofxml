package iof2

import (
	"encoding/xml"
	"strconv"
	"strings"

	"github.com/MetsaApp/iofxml"
)

// The IOF XML 2.0.3 shapes, as the standard's DTD (IOFdata.dtd) has
// them. The names differ from 3.0 throughout: ClassShortName for the
// class, PersonId and PersonName>Family/Given for the person, Club with
// ClubId and ShortName for the organisation, CCardId for the card,
// StartNumber for the bib, StartTime>Clock for the start, Time as
// "MM:SS" or "HH:MM:SS" text, CompetitorStatus value="..." for the
// status with MisPunch and SportWithdr spelled its way, and TeamSequence
// for a relay member's leg.

type version struct {
	Version string `xml:"version,attr"`
}

type event struct {
	Name      string   `xml:"Name"`
	StartDate *instant `xml:"StartDate"`
}

type countryID struct {
	Value string `xml:"value,attr"`
}

// country is the DTD's (CountryId|Country) choice: a bare CountryId, or
// a Country wrapping one.
type country struct {
	CountryID *countryID `xml:"CountryId"`
	Country   *struct {
		ID countryID `xml:"CountryId"`
	} `xml:"Country"`
}

func (c country) code() string {
	switch {
	case c.CountryID != nil:
		return strings.TrimSpace(c.CountryID.Value)
	case c.Country != nil:
		return strings.TrimSpace(c.Country.ID.Value)
	}
	return ""
}

type person struct {
	Family      string    `xml:"PersonName>Family"`
	Given       []string  `xml:"PersonName>Given"`
	ID          iofxml.ID `xml:"PersonId"`
	Nationality country   `xml:"Nationality"`
}

type club struct {
	country

	ID        iofxml.ID `xml:"ClubId"`
	Name      string    `xml:"Name"`
	ShortName string    `xml:"ShortName"`
}

// competitor is the head every person row shares: the person by id or
// in full, and the club by id or in full, or a country in its place.
type competitor struct {
	country

	PersonID *iofxml.ID `xml:"PersonId"`
	Person   *person    `xml:"Person"`
	ClubID   *iofxml.ID `xml:"ClubId"`
	Club     *club      `xml:"Club"`
}

type card struct {
	ID string `xml:"CCardId"`
}

type duration struct {
	Format string `xml:"timeFormat,attr"`
	Value  string `xml:",chardata"`
}

type instant struct {
	Date  string `xml:"Date"`
	Clock string `xml:"Clock"`
}

type status struct {
	Value string `xml:"value,attr"`
}

type split struct {
	Code    string    `xml:"ControlCode"`
	Control string    `xml:"Control>ControlCode"`
	Time    *duration `xml:"Time"`
}

type length struct {
	Unit  string `xml:"unit,attr"`
	Value string `xml:",chardata"`
}

type result struct {
	StartNumber string    `xml:"StartNumber"`
	Bib         string    `xml:"BibNumber"`
	CardID      string    `xml:"CCardId"`
	Card        *card     `xml:"CCard"`
	StartTime   *instant  `xml:"StartTime"`
	FinishTime  *instant  `xml:"FinishTime"`
	Time        *duration `xml:"Time"`
	Position    string    `xml:"ResultPosition"`
	Status      status    `xml:"CompetitorStatus"`
	Sequence    string    `xml:"TeamSequence"`
	Length      *length   `xml:"CourseLength"`
	Splits      []split   `xml:"SplitTime"`
}

type start struct {
	StartNumber string   `xml:"StartNumber"`
	Bib         string   `xml:"BibNumber"`
	CardID      string   `xml:"CCardId"`
	Card        *card    `xml:"CCard"`
	StartTime   *instant `xml:"StartTime"`
	Sequence    string   `xml:"TeamSequence"`
	Length      *length  `xml:"CourseLength"`
}

type class struct {
	ID           iofxml.ID `xml:"ClassId"`
	Name         string    `xml:"Name"`
	ShortName    string    `xml:"ClassShortName"`
	NumberInTeam string    `xml:"numberInTeam,attr"`
}

// classHead is the DTD's (ClassId|ClassShortName|EventClass) choice that
// names the class of a ClassStart, a ClassResult or an EntryClass.
type classHead struct {
	ClassID   *iofxml.ID `xml:"ClassId"`
	ShortName string     `xml:"ClassShortName"`
	Event     *struct {
		ClassID *iofxml.ID `xml:"ClassId"`
		Class   *class     `xml:"Class"`
	} `xml:"EventClass"`
}

// class is the header as 3.0 has it. ClassShortName is the name, since
// it is the one the DTD requires and every other shape names a class by.
// The legs come from a full Class's numberInTeam, named "Leg 1", "Leg 2".
func (h classHead) class() iofxml.Class {
	var c iofxml.Class
	if h.ClassID != nil {
		c.ID = *h.ClassID
	}
	c.Name = h.ShortName
	if h.Event != nil {
		if h.Event.ClassID != nil {
			c.ID = *h.Event.ClassID
		}
		if cls := h.Event.Class; cls != nil {
			c.ID = cls.ID
			c.Name = pick(cls.ShortName, cls.Name)
			for n := atoi(cls.NumberInTeam); n > 1 && len(c.Legs) < n; {
				c.Legs = append(c.Legs, iofxml.Leg{Name: "Leg " + strconv.Itoa(len(c.Legs)+1)})
			}
		}
	}
	c.Name = strings.TrimSpace(c.Name)
	c.ShortName = c.Name
	return c
}

type personResult struct {
	competitor

	Result result `xml:"Result"`
}

type teamResult struct {
	Name        string         `xml:"TeamName"`
	StartNumber string         `xml:"StartNumber"`
	Bib         string         `xml:"BibNumber"`
	Time        *duration      `xml:"Time"`
	Status      status         `xml:"TeamStatus"`
	Members     []personResult `xml:"PersonResult"`
}

type classResult struct {
	classHead

	Persons []personResult `xml:"PersonResult"`
	Teams   []teamResult   `xml:"TeamResult"`
}

type resultList struct {
	Status  string        `xml:"status,attr"`
	Version version       `xml:"IOFVersion"`
	Event   event         `xml:"Event"`
	Classes []classResult `xml:"ClassResult"`
}

type personStart struct {
	competitor

	Start      *start `xml:"Start"`
	RaceStarts []struct {
		Start start `xml:"Start"`
	} `xml:"RaceStart"`
}

// start is the Start, or a multi-day document's first RaceStart.
func (p personStart) start() *start {
	if p.Start == nil && len(p.RaceStarts) > 0 {
		return &p.RaceStarts[0].Start
	}
	return p.Start
}

type teamStart struct {
	Name    string        `xml:"TeamName"`
	Members []personStart `xml:"PersonStart"`
}

type classStart struct {
	classHead

	Persons []personStart `xml:"PersonStart"`
	Teams   []teamStart   `xml:"TeamStart"`
}

type startList struct {
	Version version      `xml:"IOFVersion"`
	Event   event        `xml:"Event"`
	Classes []classStart `xml:"ClassStart"`
}

type entryList struct {
	Version version     `xml:"IOFVersion"`
	Clubs   []clubEntry `xml:"ClubEntry"`
}

type clubEntry struct {
	ClubID  *iofxml.ID `xml:"ClubId"`
	Club    *club      `xml:"Club"`
	Entries []entry    `xml:"Entry"`
}

// entry is one entry of a ClubEntry: a person with their cards, or a
// team, TeamName followed by (person, club?, cards, TeamSequence?)
// groups. The groups are told apart by order alone, so it is decoded by
// hand.
type entry struct {
	ID       iofxml.ID
	TeamName string
	Members  []entryMember
	Classes  []classHead
}

type entryMember struct {
	competitor

	Cards    []string
	Sequence int
}

func (e *entry) UnmarshalXML(d *xml.Decoder, _ xml.StartElement) error {
	for {
		tok, err := d.Token()
		if err != nil {
			return err //nolint:wrapcheck // The decoder's error, as any UnmarshalXML returns it.
		}
		switch t := tok.(type) {
		case xml.EndElement:
			return nil
		case xml.StartElement:
			if err = e.child(d, t); err != nil {
				return err
			}
		}
	}
}

// child reads one child of the Entry: a person opens a new member, and a
// club, card or sequence belongs to the member opened last.
func (e *entry) child(d *xml.Decoder, se xml.StartElement) error {
	last := func() *entryMember {
		if len(e.Members) == 0 {
			e.Members = append(e.Members, entryMember{})
		}
		return &e.Members[len(e.Members)-1]
	}
	var err error
	switch se.Name.Local {
	case "EntryId":
		err = d.DecodeElement(&e.ID, &se)
	case "TeamName":
		err = d.DecodeElement(&e.TeamName, &se)
	case "PersonId":
		e.Members = append(e.Members, entryMember{})
		last().PersonID = new(iofxml.ID)
		err = d.DecodeElement(last().PersonID, &se)
	case "Person":
		e.Members = append(e.Members, entryMember{})
		last().Person = new(person)
		err = d.DecodeElement(last().Person, &se)
	case "ClubId":
		last().ClubID = new(iofxml.ID)
		err = d.DecodeElement(last().ClubID, &se)
	case "Club":
		last().Club = new(club)
		err = d.DecodeElement(last().Club, &se)
	case "CCard":
		var c card
		if err = d.DecodeElement(&c, &se); err == nil {
			if v := strings.TrimSpace(c.ID); v != "" {
				last().Cards = append(last().Cards, v)
			}
		}
	case "TeamSequence":
		var s string
		if err = d.DecodeElement(&s, &se); err == nil {
			last().Sequence = atoi(s)
		}
	case "EntryClass":
		var h classHead
		if err = d.DecodeElement(&h, &se); err == nil {
			e.Classes = append(e.Classes, h)
		}
	default:
		err = d.Skip()
	}
	if err != nil {
		return err //nolint:wrapcheck // The decoder's error, as any UnmarshalXML returns it.
	}
	return nil
}
