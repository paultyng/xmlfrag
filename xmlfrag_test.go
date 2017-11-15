package xmlfrag_test

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tdewolff/minify"
	minifyxml "github.com/tdewolff/minify/xml"

	"github.com/paultyng/xmlfrag"
)

func attrValues(values ...string) []xml.Attr {
	if len(values)%2 != 0 {
		panic("values must have an even number of values")
	}
	attr := make([]xml.Attr, 0, len(values)/2)
	for i := 0; i < len(values)/2; i += 2 {
		attr = append(attr, xml.Attr{
			Name: xml.Name{
				Local: values[i],
			},
			Value: values[i+1],
		})
	}
	return attr
}

func newStartElement(local string, attr ...string) xml.StartElement {
	return xml.StartElement{
		Name: xml.Name{Local: local},
		Attr: attrValues(attr...),
	}
}

func newFragment(local, innerXML, chardata string, attr ...string) xmlfrag.Element {
	return xmlfrag.Element{
		Name:     xml.Name{Local: local},
		InnerXML: innerXML,
		Chardata: chardata,
		Attr:     attrValues(attr...),
	}
}

const xmlMediaType = "text/xml"

func minifyXML(xml string) (string, error) {
	m := minify.New()
	m.AddFunc(xmlMediaType, minifyxml.Minify)
	return m.String(xmlMediaType, xml)
}

func TestParse(t *testing.T) {
	for i, c := range []struct {
		xml      string
		conf     *xmlfrag.Config
		expected []xmlfrag.Fragment
	}{
		{
			`<head attr="value">
				<foo>bar</foo>
			</head>`,
			&xmlfrag.Config{
				Body: "head",
			},
			[]xmlfrag.Fragment{
				{
					Root:    newStartElement("head", "attr", "value"),
					Headers: []xmlfrag.Element{},
					Body:    newFragment("head", "<foo>bar</foo>", "", "attr", "value"),
				},
			},
		},
		{
			`<list>
				<item>
					<foo>foo1</foo>
					<bar>
						<baz>baz1</baz>
					</bar>
				</item>
				<item>
					<foo>foo2</foo>
				</item>
			</list>`,
			&xmlfrag.Config{
				Body: "item",
			},
			[]xmlfrag.Fragment{
				{
					Root:    newStartElement("item"),
					Headers: []xmlfrag.Element{},
					Body:    newFragment("item", "<foo>foo1</foo><bar><baz>baz1</baz></bar>", ""),
				},
				{
					Root:    newStartElement("item"),
					Headers: []xmlfrag.Element{},
					Body:    newFragment("item", "<foo>foo2</foo>", ""),
				},
			},
		},

		{
			`<xmlRoot>
				<rootTag>
					<head2>
						<name attr="head2attr">head2name</name>
						<value>head2value</value>
					</head2>
					<body>foo1</body>
				</rootTag>
				<rootTag>
					<head1>
						<name>head1name</name>
						<value>head1value</value>
					</head1>
					<head2>
						<name attr="head2attr">head2name</name>
						<value>head2value</value>
					</head2>
					<body>foo2</body>
					<body>foo3</body>
				</rootTag>
			</xmlRoot>`,
			&xmlfrag.Config{
				Root:    "rootTag",
				Body:    "body",
				Headers: []string{"head1", "head2"},
			},
			[]xmlfrag.Fragment{
				{
					Root: newStartElement("rootTag"),
					Headers: []xmlfrag.Element{
						newFragment("head2", "<name attr=\"head2attr\">head2name</name><value>head2value</value>", ""),
					},
					Body: newFragment("body", "foo1", "foo1"),
				},
				{
					Root: newStartElement("rootTag"),
					Headers: []xmlfrag.Element{
						newFragment("head1", "<name>head1name</name><value>head1value</value>", ""),
						newFragment("head2", "<name attr=\"head2attr\">head2name</name><value>head2value</value>", ""),
					},
					Body: newFragment("body", "foo2", "foo2"),
				},
				{
					Root: newStartElement("rootTag"),
					Headers: []xmlfrag.Element{
						newFragment("head1", "<name>head1name</name><value>head1value</value>", ""),
						newFragment("head2", "<name attr=\"head2attr\">head2name</name><value>head2value</value>", ""),
					},
					Body: newFragment("body", "foo3", "foo3"),
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			assert := require.New(t)
			var fragments = make([]xmlfrag.Fragment, 0)
			cb := func(f *xmlfrag.Fragment) error {
				if f == nil {
					return errors.New("nil message")
				}
				fragments = append(fragments, *f)
				return nil
			}
			s, err := minifyXML(c.xml)
			assert.NoError(err)
			d := xml.NewDecoder(strings.NewReader(s))
			p := xmlfrag.New(c.conf)

			err = p.Parse(d, cb)
			assert.NoError(err)
			assert.Equal(c.expected, fragments)
		})
	}
}

func TestParseCallbackError(t *testing.T) {
	for i, c := range []struct {
		xml  string
		conf *xmlfrag.Config
	}{
		{`<foo>bar</foo>`, &xmlfrag.Config{Root: "foo", Body: "foo"}},
		{`<foo><bar>baz</bar></foo>`, &xmlfrag.Config{Root: "foo", Body: "bar"}},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			expectedErr := errors.New("expected error!")
			cb := func(f *xmlfrag.Fragment) error {
				return expectedErr
			}
			d := xml.NewDecoder(strings.NewReader(c.xml))
			p := xmlfrag.New(c.conf)
			actualErr := p.Parse(d, cb)

			assert.Equal(t, expectedErr, actualErr)
		})
	}
}
