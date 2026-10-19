// Package iofxml reads and writes IOF XML 3.0, the data standard of the
// International Orienteering Federation, see
// https://orienteering.sport/iof/it/data-standard-3-0/.
//
// It covers three documents: an [EntryList], a [StartList] and a
// [ResultList], with the persons, teams, clubs, classes, courses, start
// times, results and split times they carry. Elements outside that
// subset (fees beyond the assigned fee, services, routes, addresses,
// extensions) are skipped on read and not written.
//
// [Read] sniffs the root element and returns a [Document] of the right
// kind. Package iof2 adds IOF XML 2.0.3, converted to the same types.
// [Marshal] writes a document with the namespace, iofVersion, createTime
// and creator that a 3.0 reader expects.
//
// Times stay in the standard's own units: a duration is seconds as a
// float64, an instant is the [DateTime] text as written. The conversion
// to a clock or a time zone is the caller's.
//
// The structs keep the values as the document writes them, whitespace
// included, so a round trip keeps the data. Helpers such as [ID.Key] and
// [PersonName.Full] trim.
package iofxml
