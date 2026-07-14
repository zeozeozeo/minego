package packets

import (
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
	"github.com/zeozeozeo/minego/internal/protocol/nbt"
)

// C2SClientInformationConfiguration represents "Client Information (configuration)".
//
// Sent when the player connects, or when settings are changed.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Client_Information_(Configuration)
type C2SClientInformationConfiguration struct {
	// e.g. en_GB.
	Locale ns.String
	// Client-side render distance, in chunks.
	ViewDistance ns.Int8
	// 0: enabled, 1: commands only, 2: hidden.
	ChatMode ns.VarInt
	// "Colors" multiplayer setting.
	ChatColors ns.Boolean
	// Bit mask for displayed skin parts.
	DisplayedSkinParts ns.Uint8
	// 0: Left, 1: Right.
	MainHand ns.VarInt
	// Enables filtering of text on signs and written book titles.
	EnableTextFiltering ns.Boolean
	// Servers usually list online players; this option should let you not show up in that list.
	AllowServerListings ns.Boolean
	// 0: all, 1: decreased, 2: minimal.
	ParticleStatus ns.VarInt
}

func (p *C2SClientInformationConfiguration) Read(buf *ns.PacketBuffer) error {
	var err error
	if p.Locale, err = buf.ReadString(16); err != nil {
		return err
	}
	if p.ViewDistance, err = buf.ReadInt8(); err != nil {
		return err
	}
	if p.ChatMode, err = buf.ReadVarInt(); err != nil {
		return err
	}
	if p.ChatColors, err = buf.ReadBool(); err != nil {
		return err
	}
	if p.DisplayedSkinParts, err = buf.ReadUint8(); err != nil {
		return err
	}
	if p.MainHand, err = buf.ReadVarInt(); err != nil {
		return err
	}
	if p.EnableTextFiltering, err = buf.ReadBool(); err != nil {
		return err
	}
	if p.AllowServerListings, err = buf.ReadBool(); err != nil {
		return err
	}
	p.ParticleStatus, err = buf.ReadVarInt()
	return err
}

func (p *C2SClientInformationConfiguration) Write(buf *ns.PacketBuffer) error {
	if err := buf.WriteString(p.Locale); err != nil {
		return err
	}
	if err := buf.WriteInt8(p.ViewDistance); err != nil {
		return err
	}
	if err := buf.WriteVarInt(p.ChatMode); err != nil {
		return err
	}
	if err := buf.WriteBool(p.ChatColors); err != nil {
		return err
	}
	if err := buf.WriteUint8(p.DisplayedSkinParts); err != nil {
		return err
	}
	if err := buf.WriteVarInt(p.MainHand); err != nil {
		return err
	}
	if err := buf.WriteBool(p.EnableTextFiltering); err != nil {
		return err
	}
	if err := buf.WriteBool(p.AllowServerListings); err != nil {
		return err
	}
	return buf.WriteVarInt(p.ParticleStatus)
}

// C2SCookieResponseConfiguration represents "Cookie Response (configuration)".
//
// Response to a Cookie Request (configuration) from the server.
// The vanilla server only accepts responses of up to 5 kiB in size.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Cookie_Response_(Configuration)
type C2SCookieResponseConfiguration struct {
	// The identifier of the cookie.
	Key ns.Identifier
	// The data of the cookie.
	Payload ns.PrefixedOptional[ns.ByteArray]
}

func (p *C2SCookieResponseConfiguration) Read(buf *ns.PacketBuffer) error {
	var err error
	if p.Key, err = buf.ReadIdentifier(); err != nil {
		return err
	}
	return p.Payload.DecodeWith(buf, func(b *ns.PacketBuffer) (ns.ByteArray, error) {
		return b.ReadByteArray(5120)
	})
}

func (p *C2SCookieResponseConfiguration) Write(buf *ns.PacketBuffer) error {
	if err := buf.WriteIdentifier(p.Key); err != nil {
		return err
	}
	return p.Payload.EncodeWith(buf, func(b *ns.PacketBuffer, v ns.ByteArray) error {
		return b.WriteByteArray(v)
	})
}

// C2SCustomPayloadConfiguration represents "Serverbound Plugin Message (configuration)".
//
// Mods and plugins can use this to send their data.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Serverbound_Plugin_Message_(Configuration)
type C2SCustomPayloadConfiguration struct {
	// Name of the plugin channel used to send the data.
	Channel ns.Identifier
	// Any data, depending on the channel.
	Data ns.ByteArray
}

func (p *C2SCustomPayloadConfiguration) Read(buf *ns.PacketBuffer) error {
	var err error
	if p.Channel, err = buf.ReadIdentifier(); err != nil {
		return err
	}
	p.Data, err = buf.ReadRemaining()
	return err
}

func (p *C2SCustomPayloadConfiguration) Write(buf *ns.PacketBuffer) error {
	if err := buf.WriteIdentifier(p.Channel); err != nil {
		return err
	}
	_, err := buf.Write(p.Data)
	return err
}

// C2SFinishConfiguration represents "Acknowledge Finish Configuration".
//
// Sent by the client to notify the server that the configuration process has finished.
// This packet switches the connection state to play.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Acknowledge_Finish_Configuration
type C2SFinishConfiguration struct{}

func (p *C2SFinishConfiguration) Read(*ns.PacketBuffer) error  { return nil }
func (p *C2SFinishConfiguration) Write(*ns.PacketBuffer) error { return nil }

// C2SKeepAliveConfiguration represents "Serverbound Keep Alive (configuration)".
//
// The server will frequently send out a keep-alive, each containing a random ID.
// The client must respond with the same packet.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Serverbound_Keep_Alive_(Configuration)
type C2SKeepAliveConfiguration struct {
	KeepAliveId ns.Int64
}

func (p *C2SKeepAliveConfiguration) Read(buf *ns.PacketBuffer) error {
	var err error
	p.KeepAliveId, err = buf.ReadInt64()
	return err
}

func (p *C2SKeepAliveConfiguration) Write(buf *ns.PacketBuffer) error {
	return buf.WriteInt64(p.KeepAliveId)
}

// C2SPongConfiguration represents "Pong (configuration)".
//
// Response to the clientbound packet (Ping) with the same id.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Pong_(Configuration)
type C2SPongConfiguration struct {
	Id ns.Int32
}

func (p *C2SPongConfiguration) Read(buf *ns.PacketBuffer) error {
	var err error
	p.Id, err = buf.ReadInt32()
	return err
}

func (p *C2SPongConfiguration) Write(buf *ns.PacketBuffer) error {
	return buf.WriteInt32(p.Id)
}

// C2SResourcePackConfiguration represents "Resource Pack Response (configuration)".
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Resource_Pack_Response_(Configuration)
type C2SResourcePackConfiguration struct {
	// The unique identifier of the resource pack received in the Add Resource Pack request.
	Uuid ns.UUID
	// Result ID (see protocol docs).
	Result ns.VarInt
}

func (p *C2SResourcePackConfiguration) Read(buf *ns.PacketBuffer) error {
	var err error
	if p.Uuid, err = buf.ReadUUID(); err != nil {
		return err
	}
	p.Result, err = buf.ReadVarInt()
	return err
}

func (p *C2SResourcePackConfiguration) Write(buf *ns.PacketBuffer) error {
	if err := buf.WriteUUID(p.Uuid); err != nil {
		return err
	}
	return buf.WriteVarInt(p.Result)
}

// KnownPack represents a single known resource pack entry.
type KnownPack struct {
	Namespace ns.String
	Id        ns.String
	Version   ns.String
}

// C2SSelectKnownPacks represents "Serverbound Known Packs".
//
// Informs the server of which data packs are present on the client.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Serverbound_Known_Packs
type C2SSelectKnownPacks struct {
	KnownPacks []KnownPack
}

func (p *C2SSelectKnownPacks) Read(buf *ns.PacketBuffer) error {
	count, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	p.KnownPacks = make([]KnownPack, count)
	for i := range p.KnownPacks {
		if p.KnownPacks[i].Namespace, err = buf.ReadString(32767); err != nil {
			return err
		}
		if p.KnownPacks[i].Id, err = buf.ReadString(32767); err != nil {
			return err
		}
		if p.KnownPacks[i].Version, err = buf.ReadString(32767); err != nil {
			return err
		}
	}
	return nil
}

func (p *C2SSelectKnownPacks) Write(buf *ns.PacketBuffer) error {
	if err := buf.WriteVarInt(ns.VarInt(len(p.KnownPacks))); err != nil {
		return err
	}
	for _, pack := range p.KnownPacks {
		if err := buf.WriteString(pack.Namespace); err != nil {
			return err
		}
		if err := buf.WriteString(pack.Id); err != nil {
			return err
		}
		if err := buf.WriteString(pack.Version); err != nil {
			return err
		}
	}
	return nil
}

// C2SAcceptCodeOfConduct represents "Accept Code of Conduct".
//
// Sent by the client to accept the server's code of conduct.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Accept_Code_Of_Conduct
type C2SAcceptCodeOfConduct struct{}

func (p *C2SAcceptCodeOfConduct) Read(*ns.PacketBuffer) error  { return nil }
func (p *C2SAcceptCodeOfConduct) Write(*ns.PacketBuffer) error { return nil }

// C2SCustomClickActionConfiguration represents "Custom Click Action (configuration)".
//
// Sent when the client clicks a Text Component with the minecraft:custom click action.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Custom_Click_Action_(Configuration)
type C2SCustomClickActionConfiguration struct {
	// The identifier for the click action.
	Id ns.Identifier
	// The data to send with the click action. May be a TAG_END (0).
	Payload nbt.Tag
}

func (p *C2SCustomClickActionConfiguration) Read(buf *ns.PacketBuffer) error {
	var err error
	if p.Id, err = buf.ReadIdentifier(); err != nil {
		return err
	}
	// read remaining bytes as NBT
	remaining, err := buf.ReadByteArray(1048576)
	if err != nil {
		return err
	}
	p.Payload, err = nbt.DecodeNetwork(remaining)
	return err
}

func (p *C2SCustomClickActionConfiguration) Write(buf *ns.PacketBuffer) error {
	if err := buf.WriteIdentifier(p.Id); err != nil {
		return err
	}
	data, err := nbt.EncodeNetwork(p.Payload)
	if err != nil {
		return err
	}
	return buf.WriteFixedByteArray(data)
}
