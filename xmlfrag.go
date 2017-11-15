package xmlfrag

import (
	"encoding/xml"
	"errors"
	"io"
)

type unmarshalElement struct {
	InnerXML string `xml:",innerxml"`
	Chardata string `xml:",chardata"`
	Comment  string `xml:",comment"`
	XMLName  xml.Name
	Attr     []xml.Attr `xml:"-"`
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

func decodeElement(d Decoder, s *xml.StartElement) (Element, error) {
	var el unmarshalElement
	err := d.DecodeElement(&el, s)
	if err != nil {
		return Element{}, err
	}
	el.Attr = append(make([]xml.Attr, 0, len(s.Attr)), s.Attr...)
	return Element{
		Attr:     el.Attr,
		Chardata: el.Chardata,
		Comment:  el.Comment,
		InnerXML: el.InnerXML,
		Name:     el.XMLName,
	}, nil
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
			n, err := decodeElement(d, el)
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
			f.Body, err = decodeElement(d, el)
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
