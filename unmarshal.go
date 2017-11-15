package xmlfrag

import (
	"bytes"
	"encoding/xml"
)

func (el Element) start() xml.StartElement {
	return xml.StartElement{
		Name: el.Name,
		Attr: el.Attr,
	}
}

func (el Element) end() xml.EndElement {
	return xml.EndElement{
		Name: el.Name,
	}
}

func (el Element) Unmarshal(v interface{}) error {
	buf := bytes.NewBuffer(nil)
	enc := xml.NewEncoder(buf)
	// err := enc.EncodeToken(xml.ProcInst{Target: "xml"})
	// if err != nil {
	// 	return err
	// }
	err := enc.EncodeToken(el.start())
	if err != nil {
		return err
	}
	err = enc.Flush()
	if err != nil {
		return err
	}
	_, err = buf.WriteString(el.InnerXML)
	if err != nil {
		return err
	}
	_, err = buf.WriteString(el.Chardata)
	if err != nil {
		return err
	}
	enc.EncodeToken(el.end())
	enc.Flush()
	return xml.Unmarshal(buf.Bytes(), v)
}
