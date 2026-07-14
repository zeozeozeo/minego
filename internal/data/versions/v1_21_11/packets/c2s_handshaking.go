//go:generate go run generate.go .

package packets

import (
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

// C2SIntention represents "Handshake".
//
// This packet causes the server to switch into the target state. It should be
// sent right after opening the TCP connection to prevent the server from disconnecting.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Handshake
type C2SIntention struct {
	// See protocol version numbers (currently 775 in Minecraft 26.1).
	ProtocolVersion ns.VarInt
	// Hostname or IP, e.g. localhost or 127.0.0.1, that was used to connect.
	ServerAddress ns.String
	// Default is 25565.
	ServerPort ns.Uint16
	// 1 for Status, 2 for Login, 3 for Transfer.
	Intent ns.VarInt
}

func (p *C2SIntention) Read(buf *ns.PacketBuffer) error {
	var err error
	if p.ProtocolVersion, err = buf.ReadVarInt(); err != nil {
		return err
	}
	if p.ServerAddress, err = buf.ReadString(255); err != nil {
		return err
	}
	if p.ServerPort, err = buf.ReadUint16(); err != nil {
		return err
	}
	p.Intent, err = buf.ReadVarInt()
	return err
}

func (p *C2SIntention) Write(buf *ns.PacketBuffer) error {
	if err := buf.WriteVarInt(p.ProtocolVersion); err != nil {
		return err
	}
	if err := buf.WriteString(p.ServerAddress); err != nil {
		return err
	}
	if err := buf.WriteUint16(p.ServerPort); err != nil {
		return err
	}
	return buf.WriteVarInt(p.Intent)
}
