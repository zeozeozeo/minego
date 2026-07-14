// Package normalized contains version-independent protocol commands and events
// used by MineGo's runtime. Version packs assign their wire IDs and codecs.
package normalized

import ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"

type Handshake struct {
	Protocol int32
	Address  string
	Port     uint16
	Intent   int32
}

func (p Handshake) Read(*ns.PacketBuffer) error { return nil }
func (p Handshake) Write(b *ns.PacketBuffer) error {
	if err := b.WriteVarInt(ns.VarInt(p.Protocol)); err != nil {
		return err
	}
	if err := b.WriteString(ns.String(p.Address)); err != nil {
		return err
	}
	if err := b.WriteUint16(ns.Uint16(p.Port)); err != nil {
		return err
	}
	return b.WriteVarInt(ns.VarInt(p.Intent))
}

type LoginStart struct {
	Name string
	UUID ns.UUID
}

func (p LoginStart) Read(*ns.PacketBuffer) error { return nil }
func (p LoginStart) Write(b *ns.PacketBuffer) error {
	if err := b.WriteString(ns.String(p.Name)); err != nil {
		return err
	}
	return b.WriteUUID(p.UUID)
}

type LoginAcknowledged struct{}

func (LoginAcknowledged) Read(*ns.PacketBuffer) error  { return nil }
func (LoginAcknowledged) Write(*ns.PacketBuffer) error { return nil }

type LoginPluginResponse struct{ MessageID int32 }

func (p LoginPluginResponse) Read(*ns.PacketBuffer) error { return nil }
func (p LoginPluginResponse) Write(b *ns.PacketBuffer) error {
	if err := b.WriteVarInt(ns.VarInt(p.MessageID)); err != nil {
		return err
	}
	return b.WriteBool(false)
}

type EncryptionResponse struct{ SharedSecret, VerifyToken []byte }

func (p EncryptionResponse) Read(*ns.PacketBuffer) error { return nil }
func (p EncryptionResponse) Write(b *ns.PacketBuffer) error {
	if err := b.WriteByteArray(ns.ByteArray(p.SharedSecret)); err != nil {
		return err
	}
	return b.WriteByteArray(ns.ByteArray(p.VerifyToken))
}
