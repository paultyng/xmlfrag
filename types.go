package xmlfrag

import "encoding/xml"

// Element represents the data of an XML element.
type Element struct {
	InnerXML string
	Chardata string
	Comment  string
	Name     xml.Name
	Attr     []xml.Attr
}

// Fragment represents the collected fragment data parsed from the decoder.
type Fragment struct {
	Root    xml.StartElement
	Headers []Element
	Body    Element
}

// FragmentFunc is a function invoked on a Fragment.
type FragmentFunc func(*Fragment) error

// Parser is an interface for the Parse method that invokes a callback for each
// decoded XML Fragment.
type Parser interface {
	Parse(Decoder, FragmentFunc) error
}

// Config represents the configuration options the Parser uses to fragment
// the XML document.
type Config struct {
	// Root is the local name of the root element (optional, if unset, will be
	// set to the same value as Body).
	Root string
	// Body is local name of the element matched for each fragment.
	Body string
	// Headers is the list of parent elements between Root and Body you want
	// captured on the Fragment.
	Headers []string
}

// Decoder represents the functionality needed from an XML decoder the
// Parser uses.  Go's xml.Decoder satisfies this interface.
type Decoder interface {
	DecodeElement(interface{}, *xml.StartElement) error
	Token() (xml.Token, error)
}
