package iof2_test

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/MetsaApp/iofxml"
	"github.com/MetsaApp/iofxml/iof2"
)

// The fixtures in testdata are synthetic, in the shapes of the standard's
// own 2.0.3 examples: a DOCTYPE, an IOFVersion element, and two of them
// in ISO-8859-1.

func readFile(t *testing.T, name string) *iofxml.Document {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	doc, err := iof2.Read(f)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return doc
}

func readString(t *testing.T, body string) *iofxml.Document {
	t.Helper()
	doc, err := iof2.Read(strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func eq[T comparable](t *testing.T, what string, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %v, want %v", what, got, want)
	}
}

func seconds(t *testing.T, what string, got *float64, want float64) {
	t.Helper()
	if got == nil {
		t.Errorf("%s: got no time, want %v", what, want)
		return
	}
	if d := *got - want; d > 0.001 || d < -0.001 {
		t.Errorf("%s: got %v, want %v", what, *got, want)
	}
}

func TestReadEveryFixture(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("testdata", "*.xml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) < 4 {
		t.Fatalf("got %d fixtures, want at least 4", len(files))
	}
	for _, f := range files {
		doc := readFile(t, filepath.Base(f))
		eq(t, f+" version", doc.Version, iof2.Version)
	}
}

func TestStartList(t *testing.T) {
	doc := readFile(t, "startlist.xml")
	eq(t, "kind", doc.Kind, iofxml.KindStartList)
	l := doc.StartList
	eq(t, "converted version", l.IOFVersion, iofxml.Version)
	// The body is ISO-8859-1.
	eq(t, "event", l.Event.Name, "Skogsö Open")
	eq(t, "event date", l.Event.StartTime.Date, "2026-09-01")
	eq(t, "event clock", l.Event.StartTime.Time, "10:00:00")
	d21 := l.ClassStarts[0]
	eq(t, "class id", d21.Class.ID.Key(), "")
	eq(t, "class", d21.Class.Name, "D21")
	eq(t, "course length", d21.Courses[0].Length, 5200)
	if len(d21.PersonStarts) != 3 {
		t.Fatalf("got %d starts, want 3", len(d21.PersonStarts))
	}
	ada := d21.PersonStarts[0]
	eq(t, "person id", ada.Person.ID().Key(), "nat:12345")
	eq(t, "name", ada.Person.Name.Full(), "Ada Alder")
	eq(t, "nationality", ada.Person.Nationality.Code, "DEN")
	eq(t, "club id", ada.Organisation.ID.Key(), "34")
	// ShortName is the club's name; the long Name is the optional one.
	eq(t, "club", ada.Organisation.Name, "Northwood OK")
	eq(t, "club country", ada.Organisation.Country.Code, "DEN")
	s := ada.Starts[0]
	eq(t, "card", strings.Join(iofxml.Cards(s.ControlCards), ","), "8123456")
	eq(t, "bib", s.BibNumber, "101")
	// No date on the start: the clock alone.
	eq(t, "start", s.StartTime, "10:38:00")
	// An empty PersonId is no id; a CCard wrapper is a card like CCardId.
	sara := d21.PersonStarts[2]
	eq(t, "no id", sara.Person.ID().Key(), "")
	eq(t, "no id name", sara.Person.Name.Full(), "Sara Spruce")
	eq(t, "wrapped card", iofxml.Cards(sara.Starts[0].ControlCards)[0], "8123458")
	eq(t, "start 3", sara.Starts[0].StartTime, "10:42:00")
	h21 := l.ClassStarts[1]
	eq(t, "class 2", h21.Class.Name, "H21")
	eq(t, "h21 name", h21.PersonStarts[0].Person.Name.Full(), "Jon Juniper")
	eq(t, "h21 club", h21.PersonStarts[0].Organisation.Name, "Riverside SK")
	eq(t, "h21 bib", h21.PersonStarts[1].Starts[0].BibNumber, "202")
}

func TestResultList(t *testing.T) {
	doc := readFile(t, "resultlist.xml")
	l := doc.ResultList
	eq(t, "status", l.EffectiveStatus(), iofxml.ListComplete)
	d21 := l.ClassResults[0].PersonResults
	if len(d21) != 3 {
		t.Fatalf("got %d results, want 3", len(d21))
	}
	ada := d21[0].Results[0]
	eq(t, "bib", ada.BibNumber, "101")
	eq(t, "card", iofxml.Cards(ada.ControlCards)[0], "8123456")
	eq(t, "start", ada.StartTime, "10:38:00")
	seconds(t, "time", ada.Time, 2112)
	eq(t, "status", ada.Status, iofxml.StatusOK)
	eq(t, "position", ada.Position, 1)
	eq(t, "splits", len(ada.SplitTimes), 4)
	eq(t, "control", ada.SplitTimes[0].ControlCode, "31")
	seconds(t, "split", ada.SplitTimes[0].Time, 723)
	seconds(t, "split 4", ada.SplitTimes[3].Time, 2050)
	mia := d21[1].Results[0]
	eq(t, "running", mia.Status, iofxml.StatusActive)
	if mia.Time != nil {
		t.Error("a running competitor has no time")
	}
	eq(t, "no position", mia.Position, 0)
	eq(t, "mia splits", len(mia.SplitTimes), 2)
	eq(t, "mia control", mia.SplitTimes[1].ControlCode, "45")
	seconds(t, "mia split", mia.SplitTimes[1].Time, 1490)
	eq(t, "dns", d21[2].Results[0].Status, iofxml.StatusDidNotStart)
	eq(t, "dns splits", len(d21[2].Results[0].SplitTimes), 0)
	h21 := l.ClassResults[1].PersonResults
	jon := h21[0].Results[0]
	// MisPunch is the 3.0 MissingPunch; a split without a time is missing.
	eq(t, "mispunch", jon.Status, iofxml.StatusMissingPunch)
	seconds(t, "jon time", jon.Time, 2165)
	eq(t, "jon splits", len(jon.SplitTimes), 3)
	eq(t, "missing control", jon.SplitTimes[1].ControlCode, "32")
	if jon.SplitTimes[1].Time != nil {
		t.Error("a missing split has no time")
	}
	eq(t, "missing", jon.SplitTimes[1].Status, iofxml.SplitMissing)
	eq(t, "punched", jon.SplitTimes[0].Status, "")
	seconds(t, "paul time", h21[1].Results[0].Time, 2021)
	eq(t, "h21 course", l.ClassResults[1].Courses[0].Length, 6100)
}

func TestRelay(t *testing.T) {
	doc := readFile(t, "resultlist-relay.xml")
	l := doc.ResultList
	eq(t, "status", l.Status, iofxml.ListSnapshot)
	cls := l.ClassResults[0]
	eq(t, "class id", cls.Class.ID.Key(), "3")
	eq(t, "class", cls.Class.Name, "Mixed relay")
	eq(t, "legs", len(cls.Class.Legs), 2)
	eq(t, "leg name", cls.Class.Legs[0].Name, "Leg 1")
	team := cls.TeamResults[0]
	eq(t, "entry id", team.EntryID.Key(), "")
	eq(t, "team", team.Name, "Northwood OK 1")
	eq(t, "bib", team.BibNumber, "1")
	// The team's club is its first member's: a 2.0.3 team has a country only.
	eq(t, "team club", team.Organisations[0].Name, "Northwood OK")
	ada := team.Members[0].Results[0]
	eq(t, "leg", ada.Leg, 1)
	eq(t, "member bib", ada.BibNumber, "1-1")
	eq(t, "dated start", ada.StartTime, "2026-09-01T11:30:00")
	seconds(t, "time", ada.Time, 1230)
	eq(t, "status", ada.Status, iofxml.StatusOK)
	pos, _ := ada.Position(iofxml.PlaceLeg)
	eq(t, "leg place", pos, 1)
	seconds(t, "split", ada.SplitTimes[0].Time, 410)
	if ada.OverallResult != nil {
		t.Error("only the furthest leg carries the team's result")
	}
	paul := team.Members[1]
	eq(t, "leg 2", paul.Results[0].Leg, 2)
	eq(t, "running", paul.Results[0].Status, iofxml.StatusActive)
	// A member who names no club keeps none; the team's applies.
	if paul.Organisation != nil {
		t.Error("the member names no club")
	}
	// The furthest leg carries the team's own time and status.
	overall := paul.Results[0].OverallResult
	eq(t, "overall status", overall.Status, iofxml.StatusActive)
	seconds(t, "overall time", overall.Time, 1230)
	eq(t, "team 2", cls.TeamResults[1].Name, "Riverside SK 1")
	eq(t, "team 2 member", cls.TeamResults[1].Members[1].Person.Name.Full(), "Jon Juniper")
}

func TestEntryList(t *testing.T) {
	doc := readFile(t, "entrylist.xml")
	l := doc.EntryList
	if len(l.PersonEntries) != 3 || len(l.TeamEntries) != 1 {
		t.Fatalf("got %d persons and %d teams, want 3 and 1", len(l.PersonEntries), len(l.TeamEntries))
	}
	ada := l.PersonEntries[0]
	eq(t, "class", ada.Classes[0].Name, "D21")
	eq(t, "entry id", ada.ID.Key(), "501")
	eq(t, "person id", ada.Person.ID().Key(), "nat:12345")
	eq(t, "card", iofxml.Cards(ada.ControlCards)[0], "8123456")
	// The ClubEntry's club is every entry's.
	eq(t, "club", ada.Organisation.Name, "Northwood OK")
	eq(t, "sara", l.PersonEntries[1].Person.Name.Full(), "Sara Spruce")
	eq(t, "sara cards", len(l.PersonEntries[1].ControlCards), 0)
	jon := l.PersonEntries[2]
	eq(t, "jon class", jon.Classes[0].Name, "H21")
	eq(t, "jon club", jon.Organisation.Name, "Riverside SK")
	eq(t, "jon entry", jon.ID.Key(), "503")
	team := l.TeamEntries[0]
	eq(t, "team class", team.Classes[0].Name, "Mixed relay")
	eq(t, "team id", team.ID.Key(), "601")
	eq(t, "team", team.Name, "Northwood OK 1")
	eq(t, "team club", team.Organisations[0].Name, "Northwood OK")
	eq(t, "leg 1", team.Members[0].Leg, 1)
	eq(t, "leg 1 name", team.Members[0].Person.Name.Full(), "Ada Alder")
	eq(t, "leg 2", team.Members[1].Leg, 2)
	eq(t, "leg 2 name", team.Members[1].Person.Name.Full(), "Paul Pine")
	eq(t, "leg 2 card", iofxml.Cards(team.Members[1].ControlCards)[0], "8223457")
}

func TestDispatch(t *testing.T) {
	// No namespace and no IOFVersion is still read as 2.0.3.
	doc := readString(t, `<ResultList><ClassResult><ClassShortName>D21</ClassShortName></ClassResult></ResultList>`)
	eq(t, "kind", doc.Kind, iofxml.KindResultList)
	eq(t, "version", doc.Version, iof2.Version)
	eq(t, "status", doc.ResultList.EffectiveStatus(), iofxml.ListComplete)
	eq(t, "class", doc.ResultList.ClassResults[0].Class.Name, "D21")

	// The list status is honoured in 2.0.3's lower case.
	doc = readString(t, `<ResultList status="delta"><IOFVersion version="2.0.3"/></ResultList>`)
	eq(t, "delta", doc.ResultList.Status, iofxml.ListDelta)

	// 3.0 in its namespace is unchanged.
	doc = readString(t, `<StartList xmlns="http://www.orienteering.org/datastandard/3.0" iofVersion="3.0"/>`)
	eq(t, "3.0", doc.Version, iofxml.Version)

	for name, body := range map[string]string{
		"namespace": `<ResultList xmlns="http://www.orienteering.org/datastandard/2.0"/>`,
		"version":   `<ResultList><IOFVersion version="1.0"/></ResultList>`,
		"root":      `<CourseData><IOFVersion version="2.0.3"/></CourseData>`,
		"status":    `<ResultList status="partial"/>`,
		"time":      `<ResultList><ClassResult><PersonResult><Result><Time>abc</Time></Result></PersonResult></ClassResult></ResultList>`,
		"team time": `<ResultList><ClassResult><TeamResult><Time>1:2:3:4</Time></TeamResult></ClassResult></ResultList>`,
		"charset":   `<?xml version="1.0" encoding="EBCDIC"?><ResultList/>`,
	} {
		_, err := iof2.Read(strings.NewReader(body))
		if !errors.Is(err, iofxml.ErrNotIOF) {
			t.Errorf("%s: got %v, want ErrNotIOF", name, err)
		}
	}
}

func TestLatin1(t *testing.T) {
	body := "<?xml version=\"1.0\" encoding=\"ISO-8859-1\"?><StartList><ClassStart><ClassShortName>D21</ClassShortName>" +
		"<PersonStart><Person><PersonName><Family>S\xf8rensen</Family><Given>Mads</Given><Given>K.</Given></PersonName>" +
		"<PersonId/></Person><Start><StartTime><Date>2026-09-01</Date><Clock>10:38</Clock></StartTime>" +
		"<CourseLength unit=\"km\">5.2</CourseLength></Start></PersonStart></ClassStart></StartList>"
	doc := readString(t, body)
	c := doc.StartList.ClassStarts[0]
	p := c.PersonStarts[0]
	eq(t, "name", p.Person.Name.Full(), "Mads K. Sørensen")
	eq(t, "no id", len(p.Person.IDs), 0)
	eq(t, "start", p.Starts[0].StartTime, "2026-09-01T10:38:00")
	eq(t, "course", c.Courses[0].Length, 5200)
}

func TestSeconds(t *testing.T) {
	for _, tc := range []struct {
		xml  string
		want float64
	}{
		{`<Time>35:12</Time>`, 2112},
		{`<Time>2:21</Time>`, 141},
		{`<Time>27:52.5</Time>`, 1672.5},
		{`<Time>1:23:03</Time>`, 4983},
		{`<Time timeFormat="HH:MM:SS">00:33:41</Time>`, 2021},
		{`<Time timeFormat="MM:SS">35:12</Time>`, 2112},
		{`<Time timeFormat="HH:MM">1:05</Time>`, 3900},
		{`<Time> 90 </Time>`, 90},
	} {
		doc := readString(t, `<ResultList><ClassResult><ClassShortName>D21</ClassShortName>`+
			`<PersonResult><PersonId>1</PersonId><Result>`+tc.xml+`<CompetitorStatus value="OK"/>`+
			`</Result></PersonResult></ClassResult></ResultList>`)
		seconds(t, tc.xml, doc.ResultList.ClassResults[0].PersonResults[0].Results[0].Time, tc.want)
	}
	for _, empty := range []string{``, `<Time></Time>`, `<Time>-----</Time>`} {
		doc := readString(t, `<ResultList><ClassResult><ClassShortName>D21</ClassShortName>`+
			`<PersonResult><PersonId>1</PersonId><Result>`+empty+`<CompetitorStatus value="SportWithdr"/>`+
			`</Result></PersonResult></ClassResult></ResultList>`)
		r := doc.ResultList.ClassResults[0].PersonResults[0]
		if r.Results[0].Time != nil {
			t.Errorf("%q: want no time", empty)
		}
		eq(t, "status", r.Results[0].Status, iofxml.StatusSportingWithdrawal)
		// A PersonId on its own is a person with an id and no name.
		eq(t, "id only", r.Person.ID().Key(), "1")
	}
}

// A converted document writes as 3.0 and reads back the same.
func TestConvertedRoundTrip(t *testing.T) {
	header := iofxml.Header{IOFVersion: iofxml.Version, CreateTime: "2026-09-01T12:00:00Z", Creator: "test"}
	for _, name := range []string{"entrylist.xml", "startlist.xml", "resultlist.xml", "resultlist-relay.xml"} {
		doc := readFile(t, name)
		var l any
		switch doc.Kind {
		case iofxml.KindEntryList:
			doc.EntryList.Header = header
			l = doc.EntryList
		case iofxml.KindStartList:
			doc.StartList.Header = header
			l = doc.StartList
		case iofxml.KindResultList:
			doc.ResultList.Header = header
			l = doc.ResultList
		}
		out, err := iofxml.Marshal(l)
		if err != nil {
			t.Fatal(err)
		}
		back, err := iofxml.Read(strings.NewReader(string(out)))
		if err != nil {
			t.Fatalf("%s: %v\n%s", name, err, out)
		}
		doc.Version = iofxml.Version
		if !reflect.DeepEqual(doc, back) {
			t.Errorf("%s: the round trip changed the document\n%s", name, out)
		}
	}
}
