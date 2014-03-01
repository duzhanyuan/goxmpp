package iq

import "encoding/xml"
import "github.com/dotdoom/goxmpp/stream"
import "github.com/dotdoom/goxmpp/stream/elements"
import "github.com/dotdoom/goxmpp/stream/elements/stanzas"

const (
	STREAM_NODE = "iq"
)

func init() {
	stream.Factory.AddConstructor(" "+STREAM_NODE, func() elements.Element {
		return &IQ{InnerElements: elements.InnerElements{ElementFactory: ElementFactory}}
	})
}

var ElementFactory = elements.NewFactory()

type IQ struct {
	XMLName xml.Name `xml:"iq"`
	stanzas.Base
	elements.InnerElements
}

func (iq *IQ) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	iq.XMLName = start.Name

	iq.SetFromStartElement(start)

	return iq.HandlerInnerElements(d, start.Name.Local)
}
