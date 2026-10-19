package iofxml

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// Kind is the root element of a document.
type Kind string

// The three roots this package reads.
const (
	KindEntryList  Kind = "EntryList"
	KindStartList  Kind = "StartList"
	KindResultList Kind = "ResultList"
)

// ErrNotIOF is wrapped by Read when the body is not a usable IOF XML
// EntryList, StartList or ResultList, whether the root is something else
// or the document inside does not decode. The message says which.
var ErrNotIOF = errors.New("iofxml: not an IOF XML document")

// Document is one document read: its kind, the standard version it was
// written in, and the list itself, in the field its kind names.
type Document struct {
	Kind Kind
	// Version is the version the document declared: "3.0", or for a
	// document read through package iof2, its 2.x version ("2.0.3" when
	// it declares none). The list is in the 3.0 types either way.
	Version    string
	EntryList  *EntryList
	StartList  *StartList
	ResultList *ResultList
}

// DecodeFunc decodes the rest of a document whose root start element the
// reader has read. Package iof2's Decode is one, for IOF XML 2.0.3.
type DecodeFunc func(dec *xml.Decoder, root xml.StartElement) (*Document, error)

// Reader reads documents. The zero Reader reads IOF XML 3.0 only.
type Reader struct {
	// V2 decodes a document without a namespace and without a 3.x
	// iofVersion, which is IOF XML 2.0.3. Nil refuses such a document
	// with a message that names package iof2.
	V2 DecodeFunc
}

// Read reads one IOF XML 3.0 document with the zero Reader. Use
// iof2.Read to read 2.0.3 documents too.
func Read(r io.Reader) (*Document, error) {
	return Reader{}.Read(r)
}

// Read reads one document. The root must be an EntryList, a StartList
// or a ResultList. Which standard it speaks is decided as follows:
//
//   - in the 3.0 namespace, it is 3.0, and an iofVersion, when present,
//     must be 3.x;
//   - without a namespace, a 3.x iofVersion makes it 3.0 (some exporters
//     leave the namespace out), and no iofVersion makes it 2.0.3, which
//     has no namespace and puts its version in an IOFVersion element;
//   - in another namespace, it is 3.0 only when it declares a 3.x
//     iofVersion, and otherwise not IOF.
//
// Anything else is ErrNotIOF naming what it was, so a document of another
// format gets a useful message. A body declared as ISO-8859-1 or
// Windows-1252, which older exporters write, is read as such.
func (rd Reader) Read(r io.Reader) (*Document, error) {
	dec := xml.NewDecoder(r)
	dec.CharsetReader = charsetReader
	root, err := rootElement(dec)
	if err != nil {
		return nil, fmt.Errorf("%w: read root element: %w", ErrNotIOF, err)
	}
	version := strings.TrimSpace(attr(root, "iofVersion"))
	v3 := strings.HasPrefix(version, "3")
	switch {
	case root.Name.Space == Namespace && version != "" && !v3:
		return nil, fmt.Errorf("%w: iofVersion %q in the 3.0 namespace, want %s", ErrNotIOF, version, Version)
	case root.Name.Space == Namespace, v3:
	case root.Name.Space == "":
		return rd.readV2(dec, root)
	default:
		return nil, fmt.Errorf("%w: root %q is in namespace %q, want %q or none for IOF XML 2.0.3",
			ErrNotIOF, root.Name.Local, root.Name.Space, Namespace)
	}
	doc, err := decode3(dec, root)
	if err != nil {
		return nil, err
	}
	doc.Version = version
	if doc.Version == "" {
		doc.Version = Version
	}
	return doc, nil
}

func (rd Reader) readV2(dec *xml.Decoder, root xml.StartElement) (*Document, error) {
	if err := checkKind(root); err != nil {
		return nil, err
	}
	if rd.V2 == nil {
		return nil, fmt.Errorf("%w: %s has no namespace and no 3.x iofVersion, so it is IOF XML 2.0.3; "+
			"read it with package iof2", ErrNotIOF, root.Name.Local)
	}
	doc, err := rd.V2(dec, root)
	if err != nil {
		return nil, fmt.Errorf("%w: decode %s (IOF XML 2.0.3): %w", ErrNotIOF, root.Name.Local, err)
	}
	return doc, nil
}

func decode3(dec *xml.Decoder, root xml.StartElement) (*Document, error) {
	if err := checkKind(root); err != nil {
		return nil, err
	}
	doc := &Document{Kind: Kind(root.Name.Local)}
	var err error
	switch doc.Kind {
	case KindEntryList:
		doc.EntryList = new(EntryList)
		err = dec.DecodeElement(doc.EntryList, &root)
	case KindStartList:
		doc.StartList = new(StartList)
		err = dec.DecodeElement(doc.StartList, &root)
	case KindResultList:
		doc.ResultList = new(ResultList)
		if err = dec.DecodeElement(doc.ResultList, &root); err == nil {
			err = checkListStatus(doc.ResultList.Status)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("%w: decode %s: %w", ErrNotIOF, root.Name.Local, err)
	}
	return doc, nil
}

func checkKind(root xml.StartElement) error {
	switch Kind(root.Name.Local) {
	case KindEntryList, KindStartList, KindResultList:
		return nil
	}
	return fmt.Errorf("%w: root element %q, want EntryList, StartList or ResultList", ErrNotIOF, root.Name.Local)
}

func checkListStatus(s ListStatus) error {
	switch ListStatus(strings.TrimSpace(string(s))) {
	case "", ListComplete, ListDelta, ListSnapshot:
		return nil
	}
	return fmt.Errorf("status %q, want Complete, Delta or Snapshot", s)
}

// rootElement is the first start element. A byte order mark, the XML
// declaration, comments and a DOCTYPE before it are skipped.
func rootElement(dec *xml.Decoder) (xml.StartElement, error) {
	for {
		tok, err := dec.Token()
		if err != nil {
			return xml.StartElement{}, err //nolint:wrapcheck // Read wraps it with ErrNotIOF.
		}
		if se, ok := tok.(xml.StartElement); ok {
			return se, nil
		}
	}
}

func attr(se xml.StartElement, name string) string {
	for _, a := range se.Attr {
		if a.Name.Space == "" && a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

// charsetReader reads the encodings older exporters declare: the 2.0.3
// standard is ISO-8859-1, and so are the exports of several Windows
// timing packages. Windows-1252 is read as ISO-8859-1, the two differing
// only in the C1 range, which names do not use.
func charsetReader(label string, r io.Reader) (io.Reader, error) {
	switch strings.ToLower(strings.TrimSpace(label)) {
	case "iso-8859-1", "iso8859-1", "latin1", "latin-1", "windows-1252", "cp1252", "us-ascii", "ascii":
		return latin1Reader{r: r}, nil
	}
	return nil, fmt.Errorf("encoding %q is not supported", label)
}

// latin1Reader re-encodes ISO-8859-1 bytes as UTF-8: a byte is one rune.
type latin1Reader struct {
	r io.Reader
}

// latin1Width is the most UTF-8 bytes one ISO-8859-1 byte becomes.
const latin1Width = 2

func (l latin1Reader) Read(p []byte) (int, error) {
	if len(p) < latin1Width {
		return 0, io.ErrShortBuffer
	}
	buf := make([]byte, len(p)/latin1Width)
	n, err := l.r.Read(buf)
	out := 0
	for _, b := range buf[:n] {
		out += utf8.EncodeRune(p[out:], rune(b))
	}
	return out, err //nolint:wrapcheck // The reader's own error, passed through as io.Reader asks.
}
