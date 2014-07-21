package stream

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"

	"github.com/goxmpp/goxmpp/stream/features"
	"github.com/goxmpp/xtream"
)

type StreamHandler func(*Stream) error

var StreamXMLName = xml.Name{Local: "stream:stream"}

type Stream struct {
	XMLName          xml.Name
	ID               string `xml:"id,attr"`
	From             string `xml:"from,attr,omitempty"` // This holds server domain name.
	To               string `xml:"to,attr,omitempty"`   // This holds user JID after bind.
	Version          string `xml:"version,attr"`
	DefaultNamespace string `xml:"-"`
	Opened           bool   `xml:"-"`
	ReOpen           bool   `xml:"-"`
	State            State
	ElementFactory   xtream.Factory
	Connection
	features.FeatureContainable
}

type streamElementFactory struct {
	featuresFactory xtream.Factory
	elementsFactory xtream.Factory
}

func NewStreamElementFactory() *streamElementFactory {
	return &streamElementFactory{xtream.NewFactory(), xtream.NodeFactory}
}

func (sef streamElementFactory) Add(cons xtream.Constructor, outer, inner xml.Name) {
	sef.featuresFactory.Add(cons, outer, inner)
}

func (sef streamElementFactory) Get(outer, inner *xml.Name) xtream.Element {
	if e := sef.featuresFactory.Get(outer, inner); e != nil {
		return e
	}

	e := sef.elementsFactory.Get(outer, inner)
	if innerEl, ok := e.(xtream.Registrable); ok {
		innerEl.SetFactory(sef)
	}

	return e
}

func NewStream(rw io.ReadWriteCloser) *Stream {
	st := &Stream{
		FeatureContainable: features.NewFeatureContainer(),
		ElementFactory:     NewStreamElementFactory(),
	}
	st.SetRW(rw)
	return st
}

func (s *Stream) Open(handler StreamHandler) error {
	s.ReOpen = false

	if err := s.ReadSentOpen(); err != nil {
		return err
	}

	if err := handler(s); err != nil {
		return err
	}

	if !s.ReOpen {
		return s.SendClose()
	}
	return nil
}

func (s *Stream) SendClose() error {
	return s.streamEncoder.EncodeToken(xml.EndElement{xml.Name{Local: "stream:stream"}})
}

func (s *Stream) ReadSentOpen() error {
	s.Opened = false
	if err := s.ReadOpen(); err != nil {
		return err
	}

	if !s.ReOpen {
		if _, err := io.WriteString(s.rw, xml.Header); err != nil {
			return err
		}
	}
	var start xml.StartElement
	start.Name = xml.Name{Local: "stream:stream", Space: "jabber:client"}
	start.Attr = append(start.Attr,
		xml.Attr{Name: xml.Name{Local: "xmlns:stream"}, Value: "http://etherx.jabber.org/streams"},
		xml.Attr{Name: xml.Name{Local: "xmlns:xml"}, Value: "http://www.w3.org/XML/1998/namespace"},
		xml.Attr{Name: xml.Name{Local: "to"}, Value: s.To},
		xml.Attr{Name: xml.Name{Local: "version"}, Value: s.Version},
		xml.Attr{Name: xml.Name{Local: "from"}, Value: s.From},
	)
	if err := s.streamEncoder.EncodeToken(start); err != nil {
		return err
	}
	// xml.Encoder doesn't flush until it generated end tag
	// so we flush here to make it send stream's open tag
	if err := s.streamEncoder.Flush(); err != nil {
		return err
	}

	s.Opened = true
	s.ReOpen = false
	return nil
}

func (self *Stream) ReadOpen() error {
	for {
		t, err := self.streamDecoder.Token()
		if err != nil {
			return err
		}
		switch t := t.(type) {
		case xml.ProcInst:
			// Good.
		case xml.StartElement:
			if t.Name.Local == "stream" {
				self.XMLName = t.Name
				self.XMLName.Local = "stream:stream"
				for _, attr := range t.Attr {
					switch attr.Name.Local {
					case "to":
						self.From = attr.Value
					case "from":
						self.To = attr.Value
					case "version":
						self.Version = attr.Value
					}
				}
				log.Printf("got <stream> from: %v, to: %v, version: %v\n", self.From, self.To, self.Version)
				return nil
			}
		}
	}
}

func (self *Stream) Close() error {
	self.Opened = false
	return self.Connection.Close()
}

func (self *Stream) WriteElement(element xtream.Element) error {
	err := self.streamEncoder.Encode(element)
	if err != nil {
		log.Println("Error sending rely:", err)
	}
	return err
}

func (self *Stream) ReadElement() (xtream.Element, error) {
	var err error
	var token xml.Token

	for token, err = self.streamDecoder.Token(); err == nil; token, err = self.streamDecoder.Token() {
		if start, ok := token.(xml.StartElement); ok {
			log.Printf("got element: %v (ns %v)\n", start.Name.Local, start.Name.Space)

			element := self.ElementFactory.Get(&StreamXMLName, &start.Name)
			if element == nil {
				return nil, fmt.Errorf("Unknown node encountered: %s", start.Name.Local)
			}

			err := self.streamDecoder.DecodeElement(element, &start)
			return element, err
		}
	}

	return nil, err
}
