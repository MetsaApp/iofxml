package iofxml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"time"
)

// Marshal writes an *EntryList, a *StartList or a *ResultList (or the
// value) as a whole document: the XML declaration, then the list indented
// by two spaces, then a newline. The root carries the 3.0 namespace and
// the header, with the defaults Header describes.
//
// The output is UTF-8 and declares it. Some older timing packages read
// only ISO-8859-1, but that encoding cannot hold every name, and an
// export that fails on a valid name is worse than one a package cannot
// read.
func Marshal(list any) ([]byte, error) {
	switch list.(type) {
	case EntryList, *EntryList, StartList, *StartList, ResultList, *ResultList:
	default:
		return nil, fmt.Errorf("iofxml: marshal %T, want an EntryList, a StartList or a ResultList", list)
	}
	body, err := xml.MarshalIndent(list, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("iofxml: marshal: %w", err)
	}
	var buf bytes.Buffer
	buf.Grow(len(xml.Header) + len(body) + 1)
	buf.WriteString(xml.Header)
	buf.Write(body)
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

// The roots' MarshalXML methods have value receivers, so [xml.Marshal]
// of a value writes the namespace too.

// MarshalXML writes the list as an EntryList root in the 3.0 namespace.
func (l EntryList) MarshalXML(e *xml.Encoder, _ xml.StartElement) error {
	type plain EntryList
	l.Header = l.withDefaults()
	return encodeRoot(e, plain(l), KindEntryList)
}

// MarshalXML writes the list as a StartList root in the 3.0 namespace.
func (l StartList) MarshalXML(e *xml.Encoder, _ xml.StartElement) error {
	type plain StartList
	l.Header = l.withDefaults()
	return encodeRoot(e, plain(l), KindStartList)
}

// MarshalXML writes the list as a ResultList root in the 3.0 namespace.
func (l ResultList) MarshalXML(e *xml.Encoder, _ xml.StartElement) error {
	type plain ResultList
	l.Header = l.withDefaults()
	return encodeRoot(e, plain(l), KindResultList)
}

// encodeRoot writes v as the root element of kind k, with the namespace
// as a plain xmlns attribute, so the encoder writes one default
// namespace on the root and no prefixes, the shape every reference
// document has.
func encodeRoot(e *xml.Encoder, v any, k Kind) error {
	start := xml.StartElement{
		Name: xml.Name{Local: string(k)},
		Attr: []xml.Attr{{Name: xml.Name{Local: "xmlns"}, Value: Namespace}},
	}
	return e.EncodeElement(v, start) //nolint:wrapcheck // The encoder's error, as any MarshalXML returns it.
}

// withDefaults is the header with the version, the creation time and
// the creator filled in where empty.
func (h Header) withDefaults() Header {
	if h.IOFVersion == "" {
		h.IOFVersion = Version
	}
	if h.CreateTime == "" {
		h.CreateTime = NewDateTime(time.Now().UTC().Truncate(time.Second))
	}
	if h.Creator == "" {
		h.Creator = DefaultCreator
	}
	return h
}
