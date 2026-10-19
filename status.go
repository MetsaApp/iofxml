package iofxml

import "strings"

// ResultStatus is a competitor's or a team's status in a result.
type ResultStatus string

// The standard's result statuses.
const (
	StatusOK                 ResultStatus = "OK"
	StatusFinished           ResultStatus = "Finished"
	StatusMissingPunch       ResultStatus = "MissingPunch"
	StatusDisqualified       ResultStatus = "Disqualified"
	StatusDidNotFinish       ResultStatus = "DidNotFinish"
	StatusActive             ResultStatus = "Active"
	StatusInactive           ResultStatus = "Inactive"
	StatusOverTime           ResultStatus = "OverTime"
	StatusSportingWithdrawal ResultStatus = "SportingWithdrawal"
	StatusNotCompeting       ResultStatus = "NotCompeting"
	StatusMoved              ResultStatus = "Moved"
	StatusMovedUp            ResultStatus = "MovedUp"
	StatusDidNotStart        ResultStatus = "DidNotStart"
	StatusDidNotEnter        ResultStatus = "DidNotEnter"
	StatusCancelled          ResultStatus = "Cancelled"
)

// Trim is the status without surrounding whitespace.
func (s ResultStatus) Trim() ResultStatus {
	return ResultStatus(strings.TrimSpace(string(s)))
}

// Known reports whether the status is one the standard defines.
func (s ResultStatus) Known() bool {
	switch s.Trim() {
	case StatusOK, StatusFinished, StatusMissingPunch, StatusDisqualified, StatusDidNotFinish,
		StatusActive, StatusInactive, StatusOverTime, StatusSportingWithdrawal, StatusNotCompeting,
		StatusMoved, StatusMovedUp, StatusDidNotStart, StatusDidNotEnter, StatusCancelled:
		return true
	}
	return false
}

// InClass reports whether the row belongs to the class it is listed in.
// Moved, MovedUp and DidNotEnter say it does not: the person is in the
// class the document moved them to, and a reader that shows the row in
// both places counts the person twice.
func (s ResultStatus) InClass() bool {
	t := s.Trim()
	return t != StatusMoved && t != StatusMovedUp && t != StatusDidNotEnter
}

// ListStatus is a ResultList's status attribute: what the list is to
// the results known so far.
type ListStatus string

// The standard's list statuses. Complete is the default.
const (
	ListComplete ListStatus = "Complete"
	ListDelta    ListStatus = "Delta"
	ListSnapshot ListStatus = "Snapshot"
)

// SplitStatus is a split time's status attribute.
type SplitStatus string

// The standard's split statuses. OK is the default: the control was
// punched. Missing has no time; Additional is a control not on the
// course.
const (
	SplitOK         SplitStatus = "OK"
	SplitMissing    SplitStatus = "Missing"
	SplitAdditional SplitStatus = "Additional"
)
