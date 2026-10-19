package iofxml_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/MetsaApp/iofxml"
)

// The fixtures in testdata are synthetic: invented people, clubs and an
// invented event, in the shapes common timing software writes.

func readFile(t *testing.T, name string) *iofxml.Document {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	doc, err := iofxml.Read(f)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return doc
}

func readString(t *testing.T, body string) *iofxml.Document {
	t.Helper()
	doc, err := iofxml.Read(strings.NewReader(body))
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
	eq(t, what, *got, want)
}

func TestReadEveryFixture(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("testdata", "*.xml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) < 5 {
		t.Fatalf("got %d fixtures, want at least 5", len(files))
	}
	for _, f := range files {
		doc := readFile(t, filepath.Base(f))
		eq(t, f+" version", doc.Version, iofxml.Version)
	}
}

func TestStartList(t *testing.T) {
	doc := readFile(t, "startlist.xml")
	eq(t, "kind", doc.Kind, iofxml.KindStartList)
	l := doc.StartList
	if l == nil || doc.ResultList != nil || doc.EntryList != nil {
		t.Fatal("want the start list only")
	}
	eq(t, "creator", l.Creator, "Example Timing 1.0")
	eq(t, "createTime", l.CreateTime, "2026-09-01T07:30:00+02:00")
	eq(t, "event", l.Event.Name, "Skogsö Open")
	eq(t, "event date", l.Event.StartTime.Date, "2026-09-01")
	if len(l.ClassStarts) != 2 {
		t.Fatalf("got %d classes, want 2", len(l.ClassStarts))
	}
	d21 := l.ClassStarts[0]
	eq(t, "class id", d21.Class.ID.Key(), "1")
	eq(t, "class", d21.Class.Name, "D21")
	eq(t, "course length", d21.Courses[0].Length, 5200)
	eq(t, "course climb", d21.Courses[0].Climb, 120)
	ada := d21.PersonStarts[0]
	eq(t, "person id", ada.Person.ID().Key(), "FED:12345")
	eq(t, "name", ada.Person.Name.Full(), "Ada Alder")
	eq(t, "nationality", ada.Person.Nationality.Code, "DNK")
	eq(t, "club", ada.Organisation.Name, "Northwood OK")
	eq(t, "club country", ada.Organisation.Country.Code, "DNK")
	start := ada.Starts[0]
	eq(t, "bib", start.BibNumber, "101")
	eq(t, "start", start.StartTime, "2026-09-01T10:38:00+02:00")
	eq(t, "punching system", start.ControlCards[0].PunchingSystem, "SI")
	eq(t, "cards", strings.Join(iofxml.Cards(start.ControlCards), ","), "8123456")
	// A person without an id is still a person.
	eq(t, "no id", d21.PersonStarts[2].Person.ID().Key(), "")
	eq(t, "no id name", d21.PersonStarts[2].Person.Name.Full(), "Sara Spruce")
}

func TestResultListSnapshot(t *testing.T) {
	doc := readFile(t, "resultlist-snapshot.xml")
	l := doc.ResultList
	eq(t, "status", l.EffectiveStatus(), iofxml.ListSnapshot)
	persons := l.ClassResults[0].PersonResults
	if len(persons) != 3 {
		t.Fatalf("got %d persons, want 3", len(persons))
	}
	ada := persons[0].Results[0]
	seconds(t, "time", ada.Time, 2112.4)
	seconds(t, "time behind", ada.TimeBehind, 0)
	eq(t, "finish", ada.FinishTime, "2026-09-01T11:13:12+02:00")
	eq(t, "status", ada.Status, iofxml.StatusOK)
	eq(t, "position", ada.Position, 1)
	if len(ada.SplitTimes) != 4 {
		t.Fatalf("got %d splits, want 4", len(ada.SplitTimes))
	}
	eq(t, "control", ada.SplitTimes[0].ControlCode, "31")
	seconds(t, "split", ada.SplitTimes[0].Time, 723)
	mia := persons[1].Results[0]
	eq(t, "running", mia.Status, iofxml.StatusActive)
	if mia.Time != nil {
		t.Error("a running competitor has no time")
	}
	eq(t, "not started", persons[2].Results[0].Status, iofxml.StatusInactive)
}

func TestResultListDeltaAndMissingSplit(t *testing.T) {
	doc := readFile(t, "resultlist-delta.xml")
	eq(t, "status", doc.ResultList.EffectiveStatus(), iofxml.ListDelta)
	mia := doc.ResultList.ClassResults[0].PersonResults[0].Results[0]
	eq(t, "status", mia.Status, iofxml.StatusMissingPunch)
	missing := mia.SplitTimes[1]
	eq(t, "split status", missing.Status, iofxml.SplitMissing)
	if missing.Time != nil {
		t.Error("a missing split has no time")
	}
}

func TestResultListDefaultsToComplete(t *testing.T) {
	doc := readString(t, `<ResultList xmlns="http://www.orienteering.org/datastandard/3.0" iofVersion="3.0">`+
		`<Event><Name>x</Name></Event></ResultList>`)
	eq(t, "status as written", doc.ResultList.Status, "")
	eq(t, "status", doc.ResultList.EffectiveStatus(), iofxml.ListComplete)
	eq(t, "classes", len(doc.ResultList.ClassResults), 0)
}

func TestRelay(t *testing.T) {
	doc := readFile(t, "resultlist-relay.xml")
	cls := doc.ResultList.ClassResults[0]
	eq(t, "legs", len(cls.Class.Legs), 2)
	eq(t, "leg name", cls.Class.Legs[1].Name, "Leg 2")
	team := cls.TeamResults[0]
	eq(t, "entry id", team.EntryID.Key(), "601")
	eq(t, "team", team.Name, "Northwood OK 1")
	eq(t, "bib", team.BibNumber, "1")
	eq(t, "team club", team.Organisations[0].Name, "Northwood OK")
	ada := team.Members[0]
	r := ada.Results[0]
	eq(t, "leg", r.Leg, 1)
	seconds(t, "overall", r.OverallResult.Time, 1230)
	eq(t, "overall status", r.OverallResult.Status, iofxml.StatusOK)
	// The fixture writes the place without its type, as some exporters do.
	pos, ok := r.Position(iofxml.PlaceLeg)
	eq(t, "untyped place found", ok, true)
	eq(t, "untyped place", pos, 1)
	// A member without an organisation of their own: the field stays nil,
	// and the team's organisation is the one that applies.
	if ada.Organisation != nil {
		t.Error("the member names no organisation")
	}
	paul := team.Members[1].Results[0]
	eq(t, "leg 2", paul.Leg, 2)
	eq(t, "running", paul.Status, iofxml.StatusActive)
	if paul.OverallResult.Time != nil {
		t.Error("a running team has no overall time")
	}
}

func TestTypedPosition(t *testing.T) {
	r := iofxml.TeamMemberRaceResult{Positions: []iofxml.TypedPlace{
		{Type: iofxml.PlaceCourse, Value: 7}, {Type: iofxml.PlaceLeg, Value: 3},
	}}
	pos, _ := r.Position(iofxml.PlaceLeg)
	eq(t, "leg", pos, 3)
	pos, _ = r.Position(iofxml.PlaceCourse)
	eq(t, "course", pos, 7)
	_, ok := (&iofxml.TeamMemberRaceResult{}).Position(iofxml.PlaceLeg)
	eq(t, "none", ok, false)
}

func TestEntryList(t *testing.T) {
	doc := readFile(t, "entrylist.xml")
	l := doc.EntryList
	eq(t, "persons", len(l.PersonEntries), 3)
	ada := l.PersonEntries[0]
	// The entry id is the entry's, the person's id the person's.
	eq(t, "entry id", ada.ID.Key(), "501")
	eq(t, "person id", ada.Person.ID().Key(), "FED:12345")
	eq(t, "class", ada.Classes[0].Name, "D21")
	eq(t, "card", iofxml.Cards(ada.ControlCards)[0], "8123456")
	eq(t, "no card", len(l.PersonEntries[1].ControlCards), 0)
	team := l.TeamEntries[0]
	eq(t, "team id", team.ID.Key(), "601")
	eq(t, "team class", team.Classes[0].Name, "Mixed relay")
	eq(t, "leg", team.Members[1].Leg, 2)
	eq(t, "member card", iofxml.Cards(team.Members[1].ControlCards)[0], "8223457")
}

func TestNotIOF(t *testing.T) {
	for name, body := range map[string]string{
		"mop":       `<MOPComplete xmlns="http://www.melin.nu/mop"></MOPComplete>`,
		"namespace": `<ResultList xmlns="http://www.orienteering.org/datastandard/2.0"/>`,
		"root":      `<CourseData xmlns="http://www.orienteering.org/datastandard/3.0"/>`,
		"status":    `<ResultList xmlns="http://www.orienteering.org/datastandard/3.0" status="Partial"/>`,
		"version":   `<ResultList xmlns="http://www.orienteering.org/datastandard/3.0" iofVersion="2.0.3"/>`,
		"time":      `<ResultList iofVersion="3.0"><ClassResult><PersonResult><Result><Time>abc</Time></Result></PersonResult></ClassResult></ResultList>`,
		"v2":        `<ResultList><IOFVersion version="2.0.3"/></ResultList>`,
		"charset":   `<?xml version="1.0" encoding="EBCDIC"?><ResultList iofVersion="3.0"/>`,
		"empty":     ``,
		"garbage":   `not xml at all`,
	} {
		_, err := iofxml.Read(strings.NewReader(body))
		if !errors.Is(err, iofxml.ErrNotIOF) {
			t.Errorf("%s: got %v, want ErrNotIOF", name, err)
		}
	}
	_, err := iofxml.Read(strings.NewReader(`<MOPComplete xmlns="http://www.melin.nu/mop"></MOPComplete>`))
	if !strings.Contains(err.Error(), "MOPComplete") {
		t.Errorf("the error names the root: %v", err)
	}
	_, err = iofxml.Read(strings.NewReader(`<StartList/>`))
	if !strings.Contains(err.Error(), "iof2") {
		t.Errorf("a 2.0.3 document points to package iof2: %v", err)
	}
}

// Some exporters leave the namespace out, or write another one, but
// declare iofVersion 3.0: the version decides.
func TestVersionDecidesWithoutTheNamespace(t *testing.T) {
	for name, root := range map[string]string{
		"none":    `<ResultList iofVersion="3.0">`,
		"another": `<ResultList xmlns="http://example.com/iof/3.0" iofVersion="3.0">`,
	} {
		doc := readString(t, root+`<Event><Name>E</Name><StartTime><Date>2026-10-04</Date></StartTime></Event>`+
			`<ClassResult><Class><Name>H21</Name></Class><PersonResult><Person><Name>`+
			`<Given>Ada</Given><Family>Alder</Family></Name></Person>`+
			`<Result><Time>1800.7</Time><Status>OK</Status></Result></PersonResult></ClassResult></ResultList>`)
		eq(t, name+" kind", doc.Kind, iofxml.KindResultList)
		eq(t, name+" date", doc.ResultList.Event.StartTime.Date, "2026-10-04")
		p := doc.ResultList.ClassResults[0].PersonResults[0]
		// Given before Family is out of the schema's order; it reads all the same.
		eq(t, name+" name", p.Person.Name.Full(), "Ada Alder")
		seconds(t, name+" time", p.Results[0].Time, 1800.7)
	}
}

func TestByteOrderMark(t *testing.T) {
	doc := readString(t, string([]byte{0xef, 0xbb, 0xbf})+`<?xml version="1.0" encoding="UTF-8"?>`+
		`<StartList xmlns="http://www.orienteering.org/datastandard/3.0" iofVersion="3.0"/>`)
	eq(t, "kind", doc.Kind, iofxml.KindStartList)
}

func TestLatin1(t *testing.T) {
	doc := readString(t, "<?xml version=\"1.0\" encoding=\"ISO-8859-1\"?>"+
		"<StartList xmlns=\"http://www.orienteering.org/datastandard/3.0\" iofVersion=\"3.0\">"+
		"<Event><Name>Sk\xf8vde \xc5ben</Name></Event></StartList>")
	eq(t, "event", doc.StartList.Event.Name, "Skøvde Åben")
}

func TestIDKey(t *testing.T) {
	eq(t, "empty", iofxml.ID{}.Key(), "")
	eq(t, "type only", iofxml.ID{Type: "FED"}.Key(), "")
	eq(t, "trimmed", iofxml.ID{Value: " 17 "}.Key(), "17")
	eq(t, "typed", iofxml.ID{Type: "FED", Value: "17"}.Key(), "FED:17")
}

func TestResultStatus(t *testing.T) {
	for _, s := range []iofxml.ResultStatus{iofxml.StatusMoved, iofxml.StatusMovedUp, iofxml.StatusDidNotEnter} {
		eq(t, string(s)+" in class", s.InClass(), false)
	}
	eq(t, "OK in class", iofxml.StatusOK.InClass(), true)
	eq(t, "trimmed", iofxml.ResultStatus(" Moved ").InClass(), false)
	eq(t, "known", iofxml.StatusSportingWithdrawal.Known(), true)
	eq(t, "unknown", iofxml.ResultStatus("Teleported").Known(), false)
}

func TestDateTime(t *testing.T) {
	cph, err := time.LoadLocation("Europe/Copenhagen")
	if err != nil {
		t.Skip("no time zone data:", err)
	}
	got, err := iofxml.DateTime("2026-09-01T10:38:00+02:00").Time(time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	eq(t, "zoned", got.UTC().Format(time.RFC3339), "2026-09-01T08:38:00Z")
	got, err = iofxml.DateTime("2026-09-01T10:38:00.5").Time(cph)
	if err != nil {
		t.Fatal(err)
	}
	eq(t, "local", got.UTC().Format(time.RFC3339Nano), "2026-09-01T08:38:00.5Z")
	if _, err = iofxml.DateTime("10:38:00").Time(cph); err == nil {
		t.Error("a time of day alone is not an instant")
	}
	eq(t, "new", iofxml.NewDateTime(time.Date(2026, 9, 1, 10, 38, 0, 0, cph)), "2026-09-01T10:38:00+02:00")
}
