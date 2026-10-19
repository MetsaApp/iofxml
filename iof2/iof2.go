// Package iof2 reads IOF XML 2.0.3, the standard's previous version, and
// converts it to the IOF XML 3.0 types of package iofxml, so nothing past
// the read knows which version a document spoke.
//
// A 2.0.3 document has no namespace. Its root carries an
// <IOFVersion version="2.0.3"/> child and, as the standard's own examples
// and several Windows exporters write it, a DOCTYPE and an ISO-8859-1
// encoding declaration. Package iofxml reads that encoding.
package iof2

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/MetsaApp/iofxml"
)

// Version is the version this package reads, and the one a document
// that declares none is taken to be.
const Version = "2.0.3"

// Read reads one document, IOF XML 3.0 or 2.0.3. It is iofxml.Read with
// Decode for the documents without a namespace.
func Read(r io.Reader) (*iofxml.Document, error) {
	return iofxml.Reader{V2: Decode}.Read(r) //nolint:wrapcheck // iofxml's error, already wrapping iofxml.ErrNotIOF.
}

// Decode decodes the rest of a 2.0.3 document whose root start element
// the reader has read: an EntryList, a StartList or a ResultList. It is
// an iofxml.DecodeFunc.
func Decode(dec *xml.Decoder, root xml.StartElement) (*iofxml.Document, error) {
	switch iofxml.Kind(root.Name.Local) {
	case iofxml.KindResultList:
		return decodeResultList(dec, root)
	case iofxml.KindStartList:
		return decodeStartList(dec, root)
	case iofxml.KindEntryList:
		return decodeEntryList(dec, root)
	}
	return nil, fmt.Errorf("root element %q, want EntryList, StartList or ResultList", root.Name.Local)
}

// checkVersion accepts any 2.x, and a document that declares none: the
// root's lack of a namespace already said which standard it speaks. It
// is the version the document is taken to be.
func checkVersion(v version) (string, error) {
	s := strings.TrimSpace(v.Version)
	switch {
	case s == "":
		return Version, nil
	case strings.HasPrefix(s, "2."):
		return s, nil
	}
	return "", fmt.Errorf("IOFVersion %q, want %s", s, Version)
}

func header() iofxml.Header {
	return iofxml.Header{IOFVersion: iofxml.Version}
}

func eventOf(e event) iofxml.Event {
	out := iofxml.Event{Name: strings.TrimSpace(e.Name)}
	if e.StartDate != nil && strings.TrimSpace(e.StartDate.Date) != "" {
		out.StartTime = &iofxml.DateAndOptionalTime{
			Date: strings.TrimSpace(e.StartDate.Date),
			Time: clock(e.StartDate.Clock),
		}
	}
	return out
}

// listStatus is the list's status attribute, which 2.0.3 spells in
// lower case. None stays none, which 3.0 reads as Complete.
func listStatus(s string) (iofxml.ListStatus, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "":
		return "", nil
	case "complete":
		return iofxml.ListComplete, nil
	case "snapshot":
		return iofxml.ListSnapshot, nil
	case "delta":
		return iofxml.ListDelta, nil
	}
	return "", fmt.Errorf("status %q, want complete, snapshot or delta", s)
}

// statusOf is the 3.0 ResultStatus for a 2.0.3 CompetitorStatus or
// TeamStatus: two are spelled differently, the rest the same.
func statusOf(s string) iofxml.ResultStatus {
	switch s = strings.TrimSpace(s); s {
	case "MisPunch":
		return iofxml.StatusMissingPunch
	case "SportWithdr":
		return iofxml.StatusSportingWithdrawal
	}
	return iofxml.ResultStatus(s)
}

// personOf is a person row's head as 3.0 has it: the person with their
// nationality, nil when the row names nobody, and their club, nil when
// they name none.
func personOf(c competitor) (*iofxml.Person, *iofxml.Organisation) {
	var p iofxml.Person
	var id iofxml.ID
	switch {
	case c.Person != nil:
		p.Name.Family = strings.TrimSpace(c.Person.Family)
		given := make([]string, 0, len(c.Person.Given))
		for _, g := range c.Person.Given {
			if g = strings.TrimSpace(g); g != "" {
				given = append(given, g)
			}
		}
		p.Name.Given = strings.Join(given, " ")
		id = c.Person.ID
		if code := c.Person.Nationality.code(); code != "" {
			p.Nationality = &iofxml.Country{Code: code}
		}
	case c.PersonID != nil:
		id = *c.PersonID
	}
	if id.Key() != "" {
		p.IDs = []iofxml.ID{id}
	}
	if code := c.code(); p.Nationality == nil && code != "" {
		p.Nationality = &iofxml.Country{Code: code}
	}
	org := orgOf(c.ClubID, c.Club)
	if p.IDs == nil && p.Name == (iofxml.PersonName{}) && p.Nationality == nil {
		return nil, org
	}
	return &p, org
}

// orgOf is a club by id or in full. The ShortName is the name, since the
// DTD requires it and makes Name the optional long form.
func orgOf(id *iofxml.ID, c *club) *iofxml.Organisation {
	switch {
	case c != nil:
		o := &iofxml.Organisation{ID: c.ID, Name: strings.TrimSpace(pick(c.ShortName, c.Name))}
		if code := c.code(); code != "" {
			o.Country = &iofxml.Country{Code: code}
		}
		return o
	case id != nil && id.Key() != "":
		return &iofxml.Organisation{ID: *id}
	}
	return nil
}

// cardsOf is the card of the DTD's (CCardId|CCard) choice.
func cardsOf(id string, c *card) []iofxml.ControlCard {
	if c != nil {
		id = pick(id, c.ID)
	}
	if id = strings.TrimSpace(id); id != "" {
		return []iofxml.ControlCard{{Value: id}}
	}
	return nil
}

// orgs is a team's organisations: its first member's club, since the
// DTD gives a team a country but no club of its own.
func orgs(members []*iofxml.Organisation) []iofxml.Organisation {
	for _, o := range members {
		if o != nil {
			return []iofxml.Organisation{*o}
		}
	}
	return nil
}

// legOf is the leg a TeamSequence names, or the member's place in the
// team when it names none.
func legOf(sequence int, i int) int {
	if sequence > 0 {
		return sequence
	}
	return i + 1
}
