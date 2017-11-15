package xmlfrag

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestElementUnmarshal(t *testing.T) {
	type simple struct {
		Foo string `xml:"foo"`
		Bar string `xml:"bar"`
	}

	type complex struct {
		Foo string `xml:"foo,attr"`
		Baz int    `xml:"bar>baz"`
	}

	{
		expected := &simple{
			Foo: "foo1",
			Bar: "bar1",
		}
		actual := &simple{}
		assert.NoError(t, Element{
			Name:     xml.Name{Local: "Root"},
			InnerXML: "<foo>foo1</foo><bar>bar1</bar>",
		}.Unmarshal(actual))
		assert.Equal(t, expected, actual)
	}

	{
		expected := &complex{
			Foo: "foo1",
			Baz: 3,
		}
		actual := &complex{}
		assert.NoError(t, Element{
			Name: xml.Name{Local: "Root"},
			Attr: []xml.Attr{
				{Name: xml.Name{Local: "foo"}, Value: "foo1"},
			},
			InnerXML: "<bar><baz>3</baz></bar>",
		}.Unmarshal(actual))
		assert.Equal(t, expected, actual)
	}
}
