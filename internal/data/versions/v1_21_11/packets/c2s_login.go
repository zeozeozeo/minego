package packets

import (
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

// C2SHello represents "Login Start".
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Login_Start
type C2SHello struct {
	// Player's Username.
	Name ns.String
	// The UUID of the player logging in. Unused by the vanilla server.
	PlayerUuid ns.UUID
}

func (p *C2SHello) Read(buf *ns.PacketBuffer) error {
	var err error
	if p.Name, err = buf.ReadString(16); err != nil {
		return err
	}
	p.PlayerUuid, err = buf.ReadUUID()
	return err
}

func (p *C2SHello) Write(buf *ns.PacketBuffer) error {
	if err := buf.WriteString(p.Name); err != nil {
		return err
	}
	return buf.WriteUUID(p.PlayerUuid)
}

// C2SKey represents "Encryption Response".
//
// See protocol encryption for details.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Encryption_Response
type C2SKey struct {
	// Shared Secret value, encrypted with the server's public key.
	SharedSecret ns.ByteArray
	// Verify Token value, encrypted with the same public key as the shared secret.
	VerifyToken ns.ByteArray
}

func (p *C2SKey) Read(buf *ns.PacketBuffer) error {
	var err error
	if p.SharedSecret, err = buf.ReadByteArray(256); err != nil {
		return err
	}
	p.VerifyToken, err = buf.ReadByteArray(256)
	return err
}

func (p *C2SKey) Write(buf *ns.PacketBuffer) error {
	if err := buf.WriteByteArray(p.SharedSecret); err != nil {
		return err
	}
	return buf.WriteByteArray(p.VerifyToken)
}

// C2SCustomQueryAnswer represents "Login Plugin Response".
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Login_Plugin_Response
type C2SCustomQueryAnswer struct {
	// Should match ID from server.
	MessageId ns.VarInt
	// Any data, depending on the channel. Only present if the client understood the request.
	Data ns.PrefixedOptional[ns.ByteArray]
}

func (p *C2SCustomQueryAnswer) Read(buf *ns.PacketBuffer) error {
	var err error
	if p.MessageId, err = buf.ReadVarInt(); err != nil {
		return err
	}
	return p.Data.DecodeWith(buf, func(b *ns.PacketBuffer) (ns.ByteArray, error) {
		return b.ReadByteArray(1048576)
	})
}

func (p *C2SCustomQueryAnswer) Write(buf *ns.PacketBuffer) error {
	if err := buf.WriteVarInt(p.MessageId); err != nil {
		return err
	}
	return p.Data.EncodeWith(buf, func(b *ns.PacketBuffer, v ns.ByteArray) error {
		return b.WriteByteArray(v)
	})
}

// C2SLoginAcknowledged represents "Login Acknowledged".
//
// Acknowledgement to the Login Success packet sent by the server.
// This packet switches the connection state to configuration.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Login_Acknowledged
type C2SLoginAcknowledged struct{}

func (p *C2SLoginAcknowledged) Read(*ns.PacketBuffer) error  { return nil }
func (p *C2SLoginAcknowledged) Write(*ns.PacketBuffer) error { return nil }

// C2SCookieResponseLogin represents "Cookie Response (login)".
//
// Response to a Cookie Request (login) from the server.
// The vanilla server only accepts responses of up to 5 kiB in size.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Cookie_Response_(Login)
type C2SCookieResponseLogin struct {
	// The identifier of the cookie.
	Key ns.Identifier
	// The data of the cookie.
	Payload ns.PrefixedOptional[ns.ByteArray]
}

func (p *C2SCookieResponseLogin) Read(buf *ns.PacketBuffer) error {
	var err error
	if p.Key, err = buf.ReadIdentifier(); err != nil {
		return err
	}
	return p.Payload.DecodeWith(buf, func(b *ns.PacketBuffer) (ns.ByteArray, error) {
		return b.ReadByteArray(5120)
	})
}

func (p *C2SCookieResponseLogin) Write(buf *ns.PacketBuffer) error {
	if err := buf.WriteIdentifier(p.Key); err != nil {
		return err
	}
	return p.Payload.EncodeWith(buf, func(b *ns.PacketBuffer, v ns.ByteArray) error {
		return b.WriteByteArray(v)
	})
}
