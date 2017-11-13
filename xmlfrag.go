package xmlfrag

import (
	"encoding/xml"
	"errors"
	"io"
)

// Element represents the data of an XML element.
type Element struct {
	InnerXML string   `xml:",innerxml"`
	Chardata string   `xml:",chardata"`
	Comment  string   `xml:",comment"`
	XMLName  xml.Name `xml:",name"`
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
	Root    string
	Body    string
	Headers []string
}

// Decoder represents the functionality needed from an XML decoder the
// Parser uses.  Go's xml.Decoder satisfies this interface.
type Decoder interface {
	DecodeElement(interface{}, *xml.StartElement) error
	Token() (xml.Token, error)
}

type xmlDecoderParser struct {
	root    string
	body    string
	headers map[string]bool
}

// New returns a value implementing the Parser interface for the given
// configuration.
func New(conf *Config) Parser {
	root := conf.Root
	if root == "" {
		root = conf.Body
	}

	headers := make(map[string]bool, len(conf.Headers))
	for _, h := range conf.Headers {
		headers[h] = true
	}

	return &xmlDecoderParser{
		root:    root,
		body:    conf.Body,
		headers: headers,
	}
}

// Parse reads an XML file and triggers
// the callback with the parsed XML fragment
func (p *xmlDecoderParser) Parse(d Decoder, cb FragmentFunc) error {
	return p.parseFragments(d, cb)
}

// nolint: gocyclo
func (p *xmlDecoderParser) parseFragments(d Decoder, cb FragmentFunc) error {
	var template *Fragment
	for {
		el, err := getNextElement(d)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		if el.Name.Local == p.root {
			template = &Fragment{
				Root:    el.Copy(),
				Headers: make([]Element, 0, len(p.headers)),
			}
		}

		if template == nil {
			continue
		}

		if p.headers[el.Name.Local] {
			var n Element
			err := d.DecodeElement(&n, el)
			if err != nil {
				return err
			}
			template.Headers = append(template.Headers, n)
			continue
		}

		if el.Name.Local == p.body {
			f := &Fragment{
				Root:    template.Root,
				Headers: template.Headers,
			}
			err := d.DecodeElement(&f.Body, el)
			if err != nil {
				return err
			}
			err = cb(f)
			if err != nil {
				return err
			}
		}
	}
}

func getNextElement(d Decoder) (*xml.StartElement, error) {
	for {
		t, err := d.Token()
		if err != nil {
			return nil, err
		}
		if t == nil {
			return nil, errors.New("unexpected nil token")
		}
		el, ok := t.(xml.StartElement)
		if !ok {
			continue
		}
		return &el, nil
	}
}
