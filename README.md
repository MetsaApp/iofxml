# iofxml

`iofxml` is a Go module that reads and writes IOF XML. IOF XML is the data standard of the International Orienteering Federation (IOF). Timing software uses it to exchange entries, start lists and results.

- Package `iofxml` reads and writes IOF XML 3.0.
- Package `iof2` reads IOF XML 2.0.3 and converts it to the 3.0 types.

The module uses only the Go standard library.

## IOF Data Standard

This module implements a part of the IOF Data Standard 3.0. The IOF owns and publishes the standard:

- Standard: <https://orienteering.sport/iof/it/data-standard-3-0/>
- Schema (XSD) and examples: <https://github.com/international-orienteering-federation/datastandard-v3>
- Version 2.0.3 (DTD): <https://github.com/international-orienteering-federation/datastandard-v2>

This module does not include the schema or the DTD. The IOF publishes them without a licence statement.

## Install

```sh
go get github.com/MetsaApp/iofxml
```

The module needs Go 1.27 or later.

## Read a document

`iofxml.Read` reads an IOF XML 3.0 document. It finds the root element and returns a `Document` of the correct kind.

```go
doc, err := iofxml.Read(r)
if err != nil {
	return err // wraps iofxml.ErrNotIOF
}
switch doc.Kind {
case iofxml.KindEntryList:
	use(doc.EntryList)
case iofxml.KindStartList:
	use(doc.StartList)
case iofxml.KindResultList:
	use(doc.ResultList)
}
```

To read IOF XML 2.0.3 documents too, use `iof2.Read`. It returns the same `Document`, with the list in the 3.0 types. `Document.Version` tells the version of the source.

```go
doc, err := iof2.Read(r)
```

## Write a document

`iofxml.Marshal` writes an `EntryList`, a `StartList` or a `ResultList`. The output has the XML declaration, the 3.0 namespace and the `iofVersion` attribute. If `createTime` or `creator` is empty, `Marshal` writes the current time or `DefaultCreator`.

```go
list := &iofxml.EntryList{
	Event: iofxml.Event{Name: "Example Open"},
	PersonEntries: []iofxml.PersonEntry{{
		Person:  iofxml.Person{Name: iofxml.PersonName{Family: "Alder", Given: "Ada"}},
		Classes: []iofxml.Class{{ID: iofxml.ID{Value: "1"}, Name: "D21"}},
	}},
}
list.Creator = "My Timing 1.0"
out, err := iofxml.Marshal(list)
```

The struct fields are in the order of the schema. Thus, the output is valid against the schema when the values are valid.

## What the module covers

| Document | Elements |
|---|---|
| All | `iofVersion`, `createTime`, `creator`, `Event` (`Name`, `StartTime`) |
| `EntryList` | `PersonEntry`, `TeamEntry`, `TeamEntryPerson`, `ControlCard`, `Class`, `AssignedFee` |
| `StartList` | `ClassStart`, `Course`, `PersonStart`, `TeamStart`, `TeamMemberStart`, `Start` |
| `ResultList` | `status`, `ClassResult`, `Course`, `PersonResult`, `TeamResult`, `TeamMemberResult`, `Result`, `OverallResult`, `SplitTime` |
| Shared | `Id` with `type`, `Person`, `Organisation`, `Country`, `Class` with `Leg`, `ControlCard` with `punchingSystem` |

The reader skips other elements, for example services, routes, addresses and extensions. The writer does not write them.

## Units

The module keeps the units of the standard. The caller converts them.

- A duration (`Time`, `TimeBehind`, a split time) is seconds, as `*float64`. `nil` means "no time".
- An instant (`StartTime`, `FinishTime`, `createTime`) is `DateTime`, the text of the document. `DateTime.Time` reads it in a time zone.
- A course length or climb is metres.
- An amount of money is text, so it never goes through a float.

## Documents that do not obey the standard

Timing software sometimes writes documents that do not agree with the standard. The reader accepts these common cases:

1. A root without a namespace, or with a different namespace, that declares `iofVersion="3.0"`. The reader reads it as 3.0.
2. Elements in a different order than the schema. For example, some software writes `Given` before `Family`.
3. A `Position` of a team member without the `type` attribute. `TeamMemberRaceResult.Position` returns it when no position has the requested type.
4. An ISO-8859-1 or Windows-1252 encoding declaration. Older software, and most IOF XML 2.0.3 exports, use these encodings.
5. A UTF-8 byte order mark before the XML declaration.

The writer always writes valid 3.0 order and UTF-8.

## IOF XML 2.0.3

Package `iof2` converts 2.0.3 to 3.0 as follows:

| 2.0.3 | 3.0 |
|---|---|
| `ClassShortName` | `Class.Name` and `Class.ShortName` |
| `Class numberInTeam="2"` | Two `Leg` elements, "Leg 1" and "Leg 2" |
| `Club ShortName` | `Organisation.Name` (the long `Name` is not used) |
| `CCardId`, `CCard` | `ControlCard` |
| `StartNumber` | `BibNumber` |
| `StartTime` with `Date` and `Clock` | `DateTime` without a zone, for example `2026-09-01T10:38:00` |
| `StartTime` with `Clock` only | The time of day only, for example `10:38:00` |
| `Time` as `MM:SS` or `HH:MM:SS` | Seconds |
| `CompetitorStatus value="MisPunch"` | `MissingPunch` |
| `CompetitorStatus value="SportWithdr"` | `SportingWithdrawal` |
| A `SplitTime` without a time | `SplitTime` with `status="Missing"` |
| `TeamSequence` | `Leg`. Without it, the position of the member in the team. |
| `TeamResult` `Time` and `TeamStatus` | `OverallResult` of the member on the last leg |
| `CourseLength` on a row | `Course.Length` of the class, in metres |
| A team | The club of the first member that has a club |
| `ResultList status="snapshot"` | `status="Snapshot"` |

A 2.0.3 start without a date is not a valid 3.0 `DateTime`. Only the caller knows the day of the event. `DateTime.Time` returns an error for it.

## Glossary

| Term | Meaning in this document |
|---|---|
| Document | One IOF XML file: an entry list, a start list or a result list |
| List | The root element of a document and its content |
| Root | The first element of a document |
| Entry | The registration of a person or a team for a class |
| Leg | One part of a relay. One team member runs one leg. |
| Split time | The time from the start to one control |
| Bib | The number that a competitor wears |
| Control card | The card or chip that records the controls, for example SPORTident |
| Timing software | Software that records the start times, finish times and results of an event |

## Contributing

Refer to [CONTRIBUTING.md](CONTRIBUTING.md).

## Licence

MIT. Refer to [LICENSE](LICENSE).
