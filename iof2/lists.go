package iof2

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/MetsaApp/iofxml"
)

func decodeResultList(dec *xml.Decoder, root xml.StartElement) (*iofxml.Document, error) {
	var x resultList
	if err := dec.DecodeElement(&x, &root); err != nil {
		return nil, err //nolint:wrapcheck // iofxml.Reader wraps it with the root's name.
	}
	v, err := checkVersion(x.Version)
	if err != nil {
		return nil, err
	}
	l := &iofxml.ResultList{Header: header(), Event: eventOf(x.Event)}
	if l.Status, err = listStatus(x.Status); err != nil {
		return nil, err
	}
	for _, cr := range x.Classes {
		c, cerr := classResultOf(cr)
		if cerr != nil {
			return nil, cerr
		}
		l.ClassResults = append(l.ClassResults, c)
	}
	return &iofxml.Document{Kind: iofxml.KindResultList, Version: v, ResultList: l}, nil
}

func classResultOf(cr classResult) (iofxml.ClassResult, error) {
	out := iofxml.ClassResult{Class: cr.class()}
	for _, pr := range cr.Persons {
		p, org := personOf(pr.competitor)
		pr3 := iofxml.PersonResult{Organisation: org}
		if p != nil {
			pr3.Person = *p
		}
		r, err := raceResultOf(pr.Result)
		if err != nil {
			return iofxml.ClassResult{}, fmt.Errorf("%s: %w", pr3.Person.Name.Full(), err)
		}
		pr3.Results = []iofxml.PersonRaceResult{r}
		out.PersonResults = append(out.PersonResults, pr3)
		out.Courses = courses(out.Courses, pr.Result.Length)
	}
	for _, tr := range cr.Teams {
		t, err := teamResultOf(tr)
		if err != nil {
			return iofxml.ClassResult{}, err
		}
		out.TeamResults = append(out.TeamResults, t)
		for _, m := range tr.Members {
			out.Courses = courses(out.Courses, m.Result.Length)
		}
	}
	return out, nil
}

func raceResultOf(r result) (iofxml.PersonRaceResult, error) {
	t, err := r.Time.seconds()
	if err != nil {
		return iofxml.PersonRaceResult{}, err
	}
	splits, err := splitsOf(r.Splits)
	if err != nil {
		return iofxml.PersonRaceResult{}, fmt.Errorf("split: %w", err)
	}
	return iofxml.PersonRaceResult{
		BibNumber:    strings.TrimSpace(pick(r.Bib, r.StartNumber)),
		StartTime:    r.StartTime.dateTime(),
		FinishTime:   r.FinishTime.dateTime(),
		Time:         t,
		Position:     atoi(r.Position),
		Status:       statusOf(r.Status.Value),
		SplitTimes:   splits,
		ControlCards: cardsOf(r.CardID, r.Card),
	}, nil
}

// teamResultOf is a TeamResult as 3.0 has it: one member per
// PersonResult, on the leg its TeamSequence says or its place in the
// team. The team's own time and status become the overall result of the
// furthest leg, where 3.0 carries a team's result.
func teamResultOf(tr teamResult) (iofxml.TeamResult, error) {
	t := iofxml.TeamResult{
		Name:      strings.TrimSpace(tr.Name),
		BibNumber: strings.TrimSpace(pick(tr.Bib, tr.StartNumber)),
	}
	clubs := make([]*iofxml.Organisation, 0, len(tr.Members))
	for i, m := range tr.Members {
		p, org := personOf(m.competitor)
		r, err := raceResultOf(m.Result)
		if err != nil {
			name := ""
			if p != nil {
				name = p.Name.Full()
			}
			return iofxml.TeamResult{}, fmt.Errorf("team %s: %s: %w", t.Name, name, err)
		}
		mr := iofxml.TeamMemberRaceResult{
			Leg:          legOf(atoi(m.Result.Sequence), i),
			BibNumber:    r.BibNumber,
			StartTime:    r.StartTime,
			FinishTime:   r.FinishTime,
			Time:         r.Time,
			Status:       r.Status,
			SplitTimes:   r.SplitTimes,
			ControlCards: r.ControlCards,
		}
		if r.Position > 0 {
			mr.Positions = []iofxml.TypedPlace{{Type: iofxml.PlaceLeg, Value: r.Position}}
		}
		t.Members = append(t.Members, iofxml.TeamMemberResult{
			Person: p, Organisation: org, Results: []iofxml.TeamMemberRaceResult{mr},
		})
		clubs = append(clubs, org)
	}
	t.Organisations = orgs(clubs)
	overall, err := tr.Time.seconds()
	if err != nil {
		return iofxml.TeamResult{}, fmt.Errorf("team %s: %w", t.Name, err)
	}
	st := statusOf(tr.Status.Value)
	if st == "" && overall == nil {
		return t, nil
	}
	furthest := -1
	for i, m := range t.Members {
		if furthest < 0 || m.Results[0].Leg >= t.Members[furthest].Results[0].Leg {
			furthest = i
		}
	}
	if furthest >= 0 {
		t.Members[furthest].Results[0].OverallResult = &iofxml.OverallResult{Time: overall, Status: st}
	}
	return t, nil
}

// splitsOf is the splits as 3.0 has them. One without a time is a
// missing punch, which 2.0.3 has no other way to say.
func splitsOf(xs []split) ([]iofxml.SplitTime, error) {
	var out []iofxml.SplitTime
	for _, s := range xs {
		st, err := s.Time.seconds()
		if err != nil {
			return nil, err
		}
		sp := iofxml.SplitTime{ControlCode: strings.TrimSpace(pick(s.Code, s.Control)), Time: st}
		if st == nil {
			sp.Status = iofxml.SplitMissing
		}
		out = append(out, sp)
	}
	return out, nil
}

func decodeStartList(dec *xml.Decoder, root xml.StartElement) (*iofxml.Document, error) {
	var x startList
	if err := dec.DecodeElement(&x, &root); err != nil {
		return nil, err //nolint:wrapcheck // iofxml.Reader wraps it with the root's name.
	}
	v, err := checkVersion(x.Version)
	if err != nil {
		return nil, err
	}
	l := &iofxml.StartList{Header: header(), Event: eventOf(x.Event)}
	for _, cs := range x.Classes {
		c := iofxml.ClassStart{Class: cs.class()}
		for _, ps := range cs.Persons {
			c.PersonStarts = append(c.PersonStarts, personStartOf(ps))
			c.Courses = startCourses(c.Courses, ps)
		}
		for _, ts := range cs.Teams {
			c.TeamStarts = append(c.TeamStarts, teamStartOf(ts))
			for _, m := range ts.Members {
				c.Courses = startCourses(c.Courses, m)
			}
		}
		l.ClassStarts = append(l.ClassStarts, c)
	}
	return &iofxml.Document{Kind: iofxml.KindStartList, Version: v, StartList: l}, nil
}

// personStartOf is an individual's start. A TeamSequence on it has no
// place in 3.0, where only a team member runs a leg, and is dropped.
func personStartOf(ps personStart) iofxml.PersonStart {
	p, org := personOf(ps.competitor)
	out := iofxml.PersonStart{Person: p, Organisation: org}
	if s := ps.start(); s != nil {
		out.Starts = []iofxml.PersonRaceStart{{
			BibNumber:    strings.TrimSpace(pick(s.Bib, s.StartNumber)),
			StartTime:    s.StartTime.dateTime(),
			ControlCards: cardsOf(s.CardID, s.Card),
		}}
	}
	return out
}

// teamStartOf is a TeamStart as 3.0 has it: one member per PersonStart,
// on the leg its TeamSequence says or its place in the team.
func teamStartOf(ts teamStart) iofxml.TeamStart {
	t := iofxml.TeamStart{Name: strings.TrimSpace(ts.Name)}
	clubs := make([]*iofxml.Organisation, 0, len(ts.Members))
	for i, m := range ts.Members {
		p, org := personOf(m.competitor)
		var rs iofxml.TeamMemberRaceStart
		seq := 0
		if s := m.start(); s != nil {
			rs.BibNumber = strings.TrimSpace(pick(s.Bib, s.StartNumber))
			rs.StartTime = s.StartTime.dateTime()
			rs.ControlCards = cardsOf(s.CardID, s.Card)
			seq = atoi(s.Sequence)
		}
		rs.Leg = legOf(seq, i)
		t.Members = append(t.Members, iofxml.TeamMemberStart{
			Person: p, Organisation: org, Starts: []iofxml.TeamMemberRaceStart{rs},
		})
		clubs = append(clubs, org)
	}
	t.Organisations = orgs(clubs)
	return t
}

func startCourses(have []iofxml.Course, ps personStart) []iofxml.Course {
	if s := ps.start(); s != nil {
		return courses(have, s.Length)
	}
	return have
}

// decodeEntryList converts the entries club by club. 3.0 keeps the team
// entries apart from the person entries, so the order between the two
// is not kept. An entry that names nobody is dropped: 3.0 requires a
// person.
func decodeEntryList(dec *xml.Decoder, root xml.StartElement) (*iofxml.Document, error) {
	var x entryList
	if err := dec.DecodeElement(&x, &root); err != nil {
		return nil, err //nolint:wrapcheck // iofxml.Reader wraps it with the root's name.
	}
	v, err := checkVersion(x.Version)
	if err != nil {
		return nil, err
	}
	l := &iofxml.EntryList{Header: header()}
	for _, ce := range x.Clubs {
		org := orgOf(ce.ClubID, ce.Club)
		for _, e := range ce.Entries {
			switch {
			case len(e.Members) == 0:
			case strings.TrimSpace(e.TeamName) == "":
				l.PersonEntries = append(l.PersonEntries, personEntryOf(e, org))
			default:
				l.TeamEntries = append(l.TeamEntries, teamEntryOf(e, org))
			}
		}
	}
	return &iofxml.Document{Kind: iofxml.KindEntryList, Version: v, EntryList: l}, nil
}

func classesOf(hs []classHead) []iofxml.Class {
	var out []iofxml.Class
	for _, h := range hs {
		out = append(out, h.class())
	}
	return out
}

// personEntryOf is an individual entry: its one member, of the club the
// ClubEntry names unless the member names one.
func personEntryOf(e entry, clubOrg *iofxml.Organisation) iofxml.PersonEntry {
	m := e.Members[0]
	p, org := personOf(m.competitor)
	out := iofxml.PersonEntry{ID: e.ID, Organisation: org, Classes: classesOf(e.Classes)}
	if p != nil {
		out.Person = *p
	}
	if out.Organisation == nil {
		out.Organisation = clubOrg
	}
	for _, c := range m.Cards {
		out.ControlCards = append(out.ControlCards, iofxml.ControlCard{Value: c})
	}
	return out
}

// teamEntryOf is a team entry of the ClubEntry's club: a member per
// person group, on the leg its TeamSequence says or its place in the
// team. A member keeps a club only when they name their own.
func teamEntryOf(e entry, clubOrg *iofxml.Organisation) iofxml.TeamEntry {
	t := iofxml.TeamEntry{ID: e.ID, Name: strings.TrimSpace(e.TeamName), Classes: classesOf(e.Classes)}
	if clubOrg != nil {
		t.Organisations = []iofxml.Organisation{*clubOrg}
	}
	for i, m := range e.Members {
		p, org := personOf(m.competitor)
		tp := iofxml.TeamEntryPerson{Person: p, Organisation: org, Leg: legOf(m.Sequence, i)}
		for _, c := range m.Cards {
			tp.ControlCards = append(tp.ControlCards, iofxml.ControlCard{Value: c})
		}
		t.Members = append(t.Members, tp)
	}
	return t
}
