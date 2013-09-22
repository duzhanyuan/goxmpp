package stream

import (
	"encoding/xml"
	"io"
)

type StreamElement struct {
	XMLName xml.Name
	ID      string
	From    string
	To      string
	Version string
}

func (sw *Wrapper) ReadStreamOpen() (*StreamElement, error) {
	for {
		t, err := sw.StreamDecoder.Token()
		if err != nil {
			return nil, err
		}
		switch t := t.(type) {
		case xml.ProcInst:
			// Good.
		case xml.StartElement:
			if t.Name.Local == "stream" {
				stream := StreamElement{}
				stream.XMLName = t.Name
				for _, attr := range t.Attr {
					switch attr.Name.Local {
					case "to":
						stream.To = attr.Value
					case "from":
						stream.From = attr.Value
					case "version":
						stream.Version = attr.Value
					}
				}

				return &stream, nil
			}
		}
	}
	return nil, nil
}

// TODO(artem): refactor
func (self *Wrapper) WriteStreamOpen(stream *StreamElement, default_namespace string) error {
	data := xml.Header

	data += "<stream:stream xmlns='" + default_namespace + "' xmlns:stream='" + stream.XMLName.Space + "'"
	if stream.ID != "" {
		data += " id='" + stream.ID + "'"
	}
	if stream.From != "" {
		data += " from='" + stream.From + "'"
	}
	if stream.To != "" {
		data += " to='" + stream.To + "'"
	}
	if stream.Version != "" {
		data += " version='" + stream.Version + "'"
	}
	data += ">"

	_, err := io.WriteString(self.rw, data)
	if err == nil {
		self.State["stream_opened"] = true
	}
	return err
}
