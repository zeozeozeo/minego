package packets

import (
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

// C2SStatusRequest represents "Status Request".
//
// The status can only be requested once, immediately after the handshake,
// before any ping. The server won't respond otherwise.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Status_Request
type C2SStatusRequest struct{}

func (p *C2SStatusRequest) Read(*ns.PacketBuffer) error  { return nil }
func (p *C2SStatusRequest) Write(*ns.PacketBuffer) error { return nil }

// C2SPingRequestStatus represents "Ping Request (status)".
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Ping_Request_(Status)
type C2SPingRequestStatus struct {
	// May be any number, but vanilla clients will always use the timestamp in milliseconds.
	Timestamp ns.Int64
}

func (p *C2SPingRequestStatus) Read(buf *ns.PacketBuffer) error {
	var err error
	p.Timestamp, err = buf.ReadInt64()
	return err
}

func (p *C2SPingRequestStatus) Write(buf *ns.PacketBuffer) error {
	return buf.WriteInt64(p.Timestamp)
}
