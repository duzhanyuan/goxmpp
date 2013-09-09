package disco_info

import "enconding/xml"

type DiscoInfoQuery struct {
	// http://xmpp.org/extensions/xep-0030.html
	XMLName xml.Name `xml:"http://jabber.org/protocol/disco#info query"`
}
