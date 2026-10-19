package iofxml_test

import (
	"bytes"
	"encoding/xml"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/MetsaApp/iofxml"
)

var update = flag.Bool( //nolint:gochecknoglobals // A test flag is a global by nature.
	"update", false, "rewrite the golden files in testdata/golden")

// list is the one list a document carries.
func list(doc *iofxml.Document) any {
	switch doc.Kind {
	case iofxml.KindEntryList:
		return doc.EntryList
	case iofxml.KindStartList:
		return doc.StartList
	case iofxml.KindResultList:
		return doc.ResultList
	}
	return nil
}

// A document read, written and read again is the same document.
func TestRoundTrip(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("testdata", "*.xml"))
	if err != nil {
		t.Fatal(err)
	}
	files = append(files, goldenFiles(t)...)
	for _, f := range files {
		name, rerr := filepath.Rel("testdata", f)
		if rerr != nil {
			t.Fatal(rerr)
		}
		first := readFile(t, name)
		out, merr := iofxml.Marshal(list(first))
		if merr != nil {
			t.Fatalf("%s: %v", f, merr)
		}
		second, rerr := iofxml.Read(bytes.NewReader(out))
		if rerr != nil {
			t.Fatalf("%s: read back: %v\n%s", f, rerr, out)
		}
		if !reflect.DeepEqual(first, second) {
			t.Errorf("%s: the round trip changed the document\n%s", f, out)
		}
	}
}

func goldenFiles(t *testing.T) []string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join("testdata", "golden", "*.xml"))
	if err != nil {
		t.Fatal(err)
	}
	return files
}

// The golden documents fill every field this package writes, so a change
// to the element order shows in the diff. Run with -update to rewrite
// them.
func TestGolden(t *testing.T) {
	header := iofxml.Header{CreateTime: "2026-09-01T12:00:00Z", Creator: "Example Timing 1.0"}
	club := &iofxml.Organisation{
		ID: iofxml.ID{Value: "34"}, Name: "Northwood OK", Country: &iofxml.Country{Code: "DNK", Name: "Denmark"},
	}
	ada := iofxml.Person{
		IDs:         []iofxml.ID{{Type: "FED", Value: "12345"}, {Type: "Other", Value: "A-1"}},
		Name:        iofxml.PersonName{Family: "Alder", Given: "Ada"},
		Nationality: &iofxml.Country{Code: "DNK", Name: "Denmark"},
	}
	paul := iofxml.Person{Name: iofxml.PersonName{Family: "Pine", Given: "Paul"}}
	card := []iofxml.ControlCard{{PunchingSystem: "SI", Value: "8123456"}}
	d21 := iofxml.Class{ID: iofxml.ID{Value: "1"}, Name: "D21", ShortName: "D21"}
	relay := iofxml.Class{
		ID: iofxml.ID{Value: "3"}, Name: "Mixed relay", Legs: []iofxml.Leg{{Name: "Leg 1"}, {Name: "Leg 2"}},
	}
	course := []iofxml.Course{{ID: iofxml.ID{Value: "A"}, Name: "Long", Length: 5200, Climb: 120}}
	event := iofxml.Event{
		Name: "Skogsö Open", StartTime: &iofxml.DateAndOptionalTime{Date: "2026-09-01", Time: "10:00:00+02:00"},
	}
	fee := []iofxml.AssignedFee{{
		Fee: iofxml.Fee{Name: "D21", Amount: &iofxml.Amount{Currency: "DKK", Value: "120.00"}},
	}}

	lists := map[string]any{
		"entrylist.xml": iofxml.EntryList{
			Header: header, Event: event,
			TeamEntries: []iofxml.TeamEntry{{
				ID: iofxml.ID{Value: "601"}, Name: "Northwood OK 1", Organisations: []iofxml.Organisation{*club},
				Members: []iofxml.TeamEntryPerson{
					{Person: &ada, Organisation: club, Leg: 1, LegOrder: 1, ControlCards: card},
					{Leg: 2},
				},
				Classes: []iofxml.Class{relay}, AssignedFees: fee,
			}},
			PersonEntries: []iofxml.PersonEntry{{
				ID: iofxml.ID{Value: "501"}, Person: ada, Organisation: club, ControlCards: card,
				Classes: []iofxml.Class{d21}, AssignedFees: fee,
			}},
		},
		"startlist.xml": iofxml.StartList{
			Header: header, Event: event,
			ClassStarts: []iofxml.ClassStart{
				{
					Class: d21, Courses: course,
					PersonStarts: []iofxml.PersonStart{{
						EntryID: iofxml.ID{Value: "501"}, Person: &ada, Organisation: club,
						Starts: []iofxml.PersonRaceStart{{
							BibNumber: "101", StartTime: "2026-09-01T10:38:00+02:00", ControlCards: card,
						}},
					}},
				},
				{
					Class: relay,
					TeamStarts: []iofxml.TeamStart{{
						EntryID: iofxml.ID{Value: "601"}, Name: "Northwood OK 1",
						Organisations: []iofxml.Organisation{*club}, BibNumber: "1",
						Members: []iofxml.TeamMemberStart{{
							Person: &paul, Organisation: club,
							Starts: []iofxml.TeamMemberRaceStart{{
								Leg: 1, LegOrder: 1, BibNumber: "1-1", StartTime: "2026-09-01T11:30:00+02:00",
								ControlCards: card,
							}},
						}},
					}},
				},
			},
		},
		"resultlist.xml": iofxml.ResultList{
			Header: header, Status: iofxml.ListSnapshot, Event: event,
			ClassResults: []iofxml.ClassResult{
				{
					Class: d21, Courses: course,
					PersonResults: []iofxml.PersonResult{{
						EntryID: iofxml.ID{Value: "501"}, Person: ada, Organisation: club,
						Results: []iofxml.PersonRaceResult{{
							BibNumber: "101", StartTime: "2026-09-01T10:38:00+02:00",
							FinishTime: "2026-09-01T11:13:12.4+02:00", Time: new(2112.4), TimeBehind: new(0.0),
							Position: 1, Status: iofxml.StatusOK,
							OverallResult: &iofxml.OverallResult{
								Time: new(4000.0), TimeBehind: new(12.0), Position: 2, Status: iofxml.StatusOK,
							},
							SplitTimes: []iofxml.SplitTime{
								{ControlCode: "31", Time: new(723.0)},
								{Status: iofxml.SplitMissing, ControlCode: "32"},
								{Status: iofxml.SplitAdditional, ControlCode: "99", Time: new(1000.0)},
							},
							ControlCards: card,
						}},
					}},
				},
				{
					Class: relay,
					TeamResults: []iofxml.TeamResult{{
						EntryID: iofxml.ID{Value: "601"}, Name: "Northwood OK 1",
						Organisations: []iofxml.Organisation{*club}, BibNumber: "1",
						Members: []iofxml.TeamMemberResult{
							{
								Person: &paul, Organisation: club,
								Results: []iofxml.TeamMemberRaceResult{
									{
										Leg:        1,
										LegOrder:   1,
										BibNumber:  "1-1",
										StartTime:  "2026-09-01T11:30:00+02:00",
										FinishTime: "2026-09-01T11:50:30+02:00",
										Time:       new(1230.0),
										TimeBehind: []iofxml.TypedTime{
											{Type: iofxml.PlaceLeg, Value: 0}, {Type: iofxml.PlaceCourse, Value: 5.5},
										},
										Positions: []iofxml.TypedPlace{
											{Type: iofxml.PlaceLeg, Value: 1}, {Type: iofxml.PlaceCourse, Value: 2},
										},
										Status: iofxml.StatusOK,
										OverallResult: &iofxml.OverallResult{
											Time:     new(1230.0),
											Position: 1,
											Status:   iofxml.StatusOK,
										},
										SplitTimes:   []iofxml.SplitTime{{ControlCode: "31", Time: new(410.0)}},
										ControlCards: card,
									},
								},
							},
							{Results: []iofxml.TeamMemberRaceResult{{Leg: 2, Status: iofxml.StatusInactive}}},
						},
					}},
				},
			},
		},
	}
	for name, l := range lists {
		got, err := iofxml.Marshal(l)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		path := filepath.Join("testdata", "golden", name)
		if *update {
			if err = os.WriteFile(path, got, 0o600); err != nil {
				t.Fatal(err)
			}
			continue
		}
		want, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("%s differs from the golden file; run go test -update and check the diff\n%s", name, got)
		}
	}
}

func TestMarshalHeaderDefaults(t *testing.T) {
	out, err := iofxml.Marshal(&iofxml.EntryList{Event: iofxml.Event{Name: "E"}})
	if err != nil {
		t.Fatal(err)
	}
	body := string(out)
	if !strings.HasPrefix(body, xml.Header) {
		t.Errorf("no XML declaration:\n%s", body)
	}
	for _, want := range []string{
		`<EntryList xmlns="http://www.orienteering.org/datastandard/3.0" iofVersion="3.0" createTime="`,
		`creator="github.com/MetsaApp/iofxml"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("want %s in\n%s", want, body)
		}
	}
	doc := readString(t, body)
	if _, err = doc.EntryList.CreateTime.Time(nil); err != nil {
		t.Errorf("createTime: %v", err)
	}
	// xml.Marshal of the value writes the namespace too.
	raw, err := xml.Marshal(iofxml.ResultList{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(raw), `<ResultList xmlns="http://www.orienteering.org/datastandard/3.0"`) {
		t.Errorf("xml.Marshal: %s", raw)
	}
}

func TestMarshalRefusesOtherTypes(t *testing.T) {
	if _, err := iofxml.Marshal(iofxml.Event{}); err == nil {
		t.Error("an Event is not a document")
	}
}

// What a document does not know, it does not state: no empty id, club,
// card, start time or fee.
func TestMarshalOmitsWhatIsNotKnown(t *testing.T) {
	out, err := iofxml.Marshal(&iofxml.EntryList{
		Event: iofxml.Event{Name: "E"},
		PersonEntries: []iofxml.PersonEntry{{
			Person:  iofxml.Person{Name: iofxml.PersonName{Family: "Pine"}},
			Classes: []iofxml.Class{{ID: iofxml.ID{Type: "x"}, Name: "Open"}},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, entry, _ := strings.Cut(string(out), "<PersonEntry>")
	for _, absent := range []string{"<Id", "<Organisation", "<ControlCard", "<StartTime", "<AssignedFee", "<Nationality"} {
		if strings.Contains(entry, absent) {
			t.Errorf("want no %s in\n%s", absent, entry)
		}
	}
	// The schema requires both name parts, so an empty given name is written.
	if !strings.Contains(entry, "<Given></Given>") {
		t.Errorf("want an empty Given in\n%s", entry)
	}
}

// A name ISO-8859-1 cannot hold is written as UTF-8 and reads back.
func TestMarshalUnicode(t *testing.T) {
	out, err := iofxml.Marshal(&iofxml.EntryList{
		Event: iofxml.Event{Name: "E"},
		PersonEntries: []iofxml.PersonEntry{{
			Person: iofxml.Person{Name: iofxml.PersonName{Family: "名前", Given: "Ævø Ünïcode"}},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `encoding="UTF-8"`) {
		t.Error("the declaration names UTF-8")
	}
	doc := readString(t, string(out))
	eq(t, "name", doc.EntryList.PersonEntries[0].Person.Name.Full(), "Ævø Ünïcode 名前")
}
