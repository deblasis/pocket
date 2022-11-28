package messaging

import "strings"

const (
	DebugMessageEventType                    = "pocket.DebugMessage"
	HotstuffMessageContentType               = "consensus.HotstuffMessage"
	UtilityMessageContentType                = "consensus.UtilityMessage"
	P2PAddressBookSnapshotMessageContentType = "p2p.P2PAddressBookSnapshotMessage"
)

func (x *PocketEnvelope) GetContentType() string {
	return strings.Split(x.Content.GetTypeUrl(), "/")[1]
}
