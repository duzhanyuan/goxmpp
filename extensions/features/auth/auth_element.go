package auth

import (
	"encoding/xml"
	"fmt"

	"github.com/dotdoom/goxmpp/stream"
	"github.com/dotdoom/goxmpp/stream/elements"
)

type AuthElement struct {
	XMLName   xml.Name `xml:"auth"`
	Mechanism string   `xml:"mechanism,attr"`
	Data      string   `xml:",chardata,omitempty"`
}

type InvalidMechanism struct {
	XMLName xml.Name `xml:"invalid-mechanism"`
}

type Handler func(*AuthElement, *stream.Stream) error

func (self *AuthElement) Handle(stream *stream.Stream) error {
	if handler := mechanism_handlers[self.Mechanism]; handler != nil {
		return handler(self, stream)
	} else {
		if err := stream.WriteElement(InvalidMechanism{}); err != nil {
			return err
		}
		return fmt.Errorf("No handler for mechanism %v", self.Mechanism)
	}
}

func init() {
	stream.StreamFactory.AddConstructor(func() elements.Element {
		return &AuthElement{}
	})
}
