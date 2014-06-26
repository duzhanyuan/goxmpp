package auth

import (
	"encoding/xml"
	"fmt"
	"log"

	"github.com/goxmpp/goxmpp/stream"
	"github.com/goxmpp/xtream"
)

type AuthElement struct {
	XMLName   xml.Name `xml:"auth"`
	Mechanism string   `xml:"mechanism,attr"`
	Data      string   `xml:",chardata"`
}

type Handler func(*AuthElement, *stream.Stream) error

func (self *AuthElement) Handle(stream *stream.Stream) error {
	if handler := mechanism_handlers[self.Mechanism]; handler != nil {
		if err := handler(self, stream); err != nil {
			log.Println("Authorization failed:", err)
			if err := stream.WriteElement(NewFailute(NotAuthorized{})); err != nil {
				return err
			}
			return err
		}
	} else {
		if err := stream.WriteElement(NewFailute(InvalidMechanism{})); err != nil {
			return err
		}
		return fmt.Errorf("No handler for mechanism %v", self.Mechanism)
	}

	return nil
}

func init() {
	xtream.NodeFactory.Add(func() xtream.Element {
		return &AuthElement{}
	}, xml.Name{Local: "stream:stream"}, xml.Name{Local: "auth"})
}
