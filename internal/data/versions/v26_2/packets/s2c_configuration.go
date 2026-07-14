package packets

import (
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
	"github.com/zeozeozeo/minego/internal/protocol/nbt"
)

// S2CCookieRequestConfiguration represents "Cookie Request (configuration)".
//
// Requests a cookie that was previously stored.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Cookie_Request_(Configuration)
type S2CCookieRequestConfiguration struct {
	// The identifier of the cookie.
	Key ns.Identifier
}

func (p *S2CCookieRequestConfiguration) Read(buf *ns.PacketBuffer) error {
	var err error
	p.Key, err = buf.ReadIdentifier()
	return err
}

func (p *S2CCookieRequestConfiguration) Write(buf *ns.PacketBuffer) error {
	return buf.WriteIdentifier(p.Key)
}

// S2CCustomPayloadConfiguration represents "Clientbound Plugin Message (configuration)".
//
// Mods and plugins can use this to send their data.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Clientbound_Plugin_Message_(Configuration)
type S2CCustomPayloadConfiguration struct {
	// Name of the plugin channel used to send the data.
	Channel ns.Identifier
	// Any data, depending on the channel.
	Data ns.ByteArray
}

func (p *S2CCustomPayloadConfiguration) Read(buf *ns.PacketBuffer) error {
	var err error
	if p.Channel, err = buf.ReadIdentifier(); err != nil {
		return err
	}
	p.Data, err = buf.ReadRemaining()
	return err
}

func (p *S2CCustomPayloadConfiguration) Write(buf *ns.PacketBuffer) error {
	if err := buf.WriteIdentifier(p.Channel); err != nil {
		return err
	}
	_, err := buf.Write(p.Data)
	return err
}

// S2CDisconnectConfiguration represents "Disconnect (configuration)".
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Disconnect_(Configuration)
type S2CDisconnectConfiguration struct {
	// The reason why the player was disconnected.
	Reason ns.TextComponent
}

func (p *S2CDisconnectConfiguration) Read(buf *ns.PacketBuffer) error {
	var err error
	p.Reason, err = buf.ReadTextComponent()
	return err
}

func (p *S2CDisconnectConfiguration) Write(buf *ns.PacketBuffer) error {
	return buf.WriteTextComponent(p.Reason)
}

// S2CFinishConfiguration represents "Finish Configuration".
//
// Sent by the server to notify the client that the configuration process has finished.
// This packet switches the connection state to play.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Finish_Configuration
type S2CFinishConfiguration struct{}

func (p *S2CFinishConfiguration) Read(*ns.PacketBuffer) error  { return nil }
func (p *S2CFinishConfiguration) Write(*ns.PacketBuffer) error { return nil }

// S2CKeepAliveConfiguration represents "Clientbound Keep Alive (configuration)".
//
// The server will frequently send out a keep-alive, each containing a random ID.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Clientbound_Keep_Alive_(Configuration)
type S2CKeepAliveConfiguration struct {
	KeepAliveId ns.Int64
}

func (p *S2CKeepAliveConfiguration) Read(buf *ns.PacketBuffer) error {
	var err error
	p.KeepAliveId, err = buf.ReadInt64()
	return err
}

func (p *S2CKeepAliveConfiguration) Write(buf *ns.PacketBuffer) error {
	return buf.WriteInt64(p.KeepAliveId)
}

// S2CPingConfiguration represents "Ping (configuration)".
//
// Packet is not used by the vanilla server.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Ping_(Configuration)
type S2CPingConfiguration struct {
	Id ns.Int32
}

func (p *S2CPingConfiguration) Read(buf *ns.PacketBuffer) error {
	var err error
	p.Id, err = buf.ReadInt32()
	return err
}

func (p *S2CPingConfiguration) Write(buf *ns.PacketBuffer) error {
	return buf.WriteInt32(p.Id)
}

// S2CResetChat represents "Reset Chat".
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Reset_Chat
type S2CResetChat struct{}

func (p *S2CResetChat) Read(*ns.PacketBuffer) error  { return nil }
func (p *S2CResetChat) Write(*ns.PacketBuffer) error { return nil }

// RegistryEntry represents a single registry entry.
type RegistryEntry struct {
	EntryId ns.Identifier
	HasData ns.Boolean
	Data    nbt.Tag // only present if HasData is true
}

// S2CRegistryData represents "Registry Data".
//
// Represents certain registries that are sent from the server.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Registry_Data
type S2CRegistryData struct {
	RegistryId ns.Identifier
	Entries    []RegistryEntry
}

func (p *S2CRegistryData) Read(buf *ns.PacketBuffer) error {
	var err error
	if p.RegistryId, err = buf.ReadIdentifier(); err != nil {
		return err
	}
	count, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	p.Entries = make([]RegistryEntry, count)
	for i := range p.Entries {
		if p.Entries[i].EntryId, err = buf.ReadIdentifier(); err != nil {
			return err
		}
		if p.Entries[i].HasData, err = buf.ReadBool(); err != nil {
			return err
		}
		if p.Entries[i].HasData {
			// read NBT data - we need to read remaining bytes for this entry,
			// as we don't know the exact NBT size upfront
			remaining, err := buf.ReadByteArray(1048576)
			if err != nil {
				return err
			}
			p.Entries[i].Data, err = nbt.DecodeNetwork(remaining)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *S2CRegistryData) Write(buf *ns.PacketBuffer) error {
	if err := buf.WriteIdentifier(p.RegistryId); err != nil {
		return err
	}
	if err := buf.WriteVarInt(ns.VarInt(len(p.Entries))); err != nil {
		return err
	}
	for _, entry := range p.Entries {
		if err := buf.WriteIdentifier(entry.EntryId); err != nil {
			return err
		}
		if err := buf.WriteBool(entry.HasData); err != nil {
			return err
		}
		if entry.HasData {
			data, err := nbt.EncodeNetwork(entry.Data)
			if err != nil {
				return err
			}
			if err := buf.WriteFixedByteArray(data); err != nil {
				return err
			}
		}
	}
	return nil
}

// S2CResourcePackPopConfiguration represents "Remove Resource Pack (configuration)".
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Remove_Resource_Pack_(Configuration)
type S2CResourcePackPopConfiguration struct {
	// The UUID of the resource pack to be removed. If not present, every resource pack will be removed.
	Uuid ns.PrefixedOptional[ns.UUID]
}

func (p *S2CResourcePackPopConfiguration) Read(buf *ns.PacketBuffer) error {
	return p.Uuid.DecodeWith(buf, func(b *ns.PacketBuffer) (ns.UUID, error) {
		return b.ReadUUID()
	})
}

func (p *S2CResourcePackPopConfiguration) Write(buf *ns.PacketBuffer) error {
	return p.Uuid.EncodeWith(buf, func(b *ns.PacketBuffer, v ns.UUID) error {
		return b.WriteUUID(v)
	})
}

// S2CResourcePackPushConfiguration represents "Add Resource Pack (configuration)".
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Add_Resource_Pack_(Configuration)
type S2CResourcePackPushConfiguration struct {
	// The unique identifier of the resource pack.
	Uuid ns.UUID
	// The URL to the resource pack.
	Url ns.String
	// A 40 character hexadecimal SHA-1 hash of the resource pack file.
	Hash ns.String
	// The vanilla client will be forced to use the resource pack from the server.
	Forced ns.Boolean
	// This is shown in the prompt making the client accept or decline the resource pack.
	PromptMessage ns.PrefixedOptional[ns.TextComponent]
}

func (p *S2CResourcePackPushConfiguration) Read(buf *ns.PacketBuffer) error {
	var err error
	if p.Uuid, err = buf.ReadUUID(); err != nil {
		return err
	}
	if p.Url, err = buf.ReadString(32767); err != nil {
		return err
	}
	if p.Hash, err = buf.ReadString(40); err != nil {
		return err
	}
	if p.Forced, err = buf.ReadBool(); err != nil {
		return err
	}
	return p.PromptMessage.DecodeWith(buf, func(b *ns.PacketBuffer) (ns.TextComponent, error) {
		return b.ReadTextComponent()
	})
}

func (p *S2CResourcePackPushConfiguration) Write(buf *ns.PacketBuffer) error {
	if err := buf.WriteUUID(p.Uuid); err != nil {
		return err
	}
	if err := buf.WriteString(p.Url); err != nil {
		return err
	}
	if err := buf.WriteString(p.Hash); err != nil {
		return err
	}
	if err := buf.WriteBool(p.Forced); err != nil {
		return err
	}
	return p.PromptMessage.EncodeWith(buf, func(b *ns.PacketBuffer, v ns.TextComponent) error {
		return b.WriteTextComponent(v)
	})
}

// S2CStoreCookieConfiguration represents "Store Cookie (configuration)".
//
// Stores some arbitrary data on the client, which persists between server transfers.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Store_Cookie_(Configuration)
type S2CStoreCookieConfiguration struct {
	// The identifier of the cookie.
	Key ns.Identifier
	// The data of the cookie.
	Payload ns.ByteArray
}

func (p *S2CStoreCookieConfiguration) Read(buf *ns.PacketBuffer) error {
	var err error
	if p.Key, err = buf.ReadIdentifier(); err != nil {
		return err
	}
	p.Payload, err = buf.ReadByteArray(5120)
	return err
}

func (p *S2CStoreCookieConfiguration) Write(buf *ns.PacketBuffer) error {
	if err := buf.WriteIdentifier(p.Key); err != nil {
		return err
	}
	return buf.WriteByteArray(p.Payload)
}

// S2CTransferConfiguration represents "Transfer (configuration)".
//
// Notifies the client that it should transfer to the given server.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Transfer_(Configuration)
type S2CTransferConfiguration struct {
	// The hostname or IP of the server.
	Host ns.String
	// The port of the server.
	Port ns.VarInt
}

func (p *S2CTransferConfiguration) Read(buf *ns.PacketBuffer) error {
	var err error
	if p.Host, err = buf.ReadString(32767); err != nil {
		return err
	}
	p.Port, err = buf.ReadVarInt()
	return err
}

func (p *S2CTransferConfiguration) Write(buf *ns.PacketBuffer) error {
	if err := buf.WriteString(p.Host); err != nil {
		return err
	}
	return buf.WriteVarInt(p.Port)
}

// S2CUpdateEnabledFeatures represents "Feature Flags".
//
// Used to enable and disable features on the client.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Feature_Flags
type S2CUpdateEnabledFeatures struct {
	FeatureFlags []ns.Identifier
}

func (p *S2CUpdateEnabledFeatures) Read(buf *ns.PacketBuffer) error {
	count, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	p.FeatureFlags = make([]ns.Identifier, count)
	for i := range p.FeatureFlags {
		if p.FeatureFlags[i], err = buf.ReadIdentifier(); err != nil {
			return err
		}
	}
	return nil
}

func (p *S2CUpdateEnabledFeatures) Write(buf *ns.PacketBuffer) error {
	if err := buf.WriteVarInt(ns.VarInt(len(p.FeatureFlags))); err != nil {
		return err
	}
	for _, flag := range p.FeatureFlags {
		if err := buf.WriteIdentifier(flag); err != nil {
			return err
		}
	}
	return nil
}

// Tag represents a single tag entry.
type Tag struct {
	TagName ns.Identifier
	Entries []ns.VarInt
}

// TagRegistry represents a registry of tags.
type TagRegistry struct {
	Registry ns.Identifier
	Tags     []Tag
}

// S2CUpdateTagsConfiguration represents "Update Tags (configuration)".
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Update_Tags_(Configuration)
type S2CUpdateTagsConfiguration struct {
	ArrayOfTags []TagRegistry
}

func (p *S2CUpdateTagsConfiguration) Read(buf *ns.PacketBuffer) error {
	registryCount, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	p.ArrayOfTags = make([]TagRegistry, registryCount)
	for i := range p.ArrayOfTags {
		if p.ArrayOfTags[i].Registry, err = buf.ReadIdentifier(); err != nil {
			return err
		}
		tagCount, err := buf.ReadVarInt()
		if err != nil {
			return err
		}
		p.ArrayOfTags[i].Tags = make([]Tag, tagCount)
		for j := range p.ArrayOfTags[i].Tags {
			if p.ArrayOfTags[i].Tags[j].TagName, err = buf.ReadIdentifier(); err != nil {
				return err
			}
			entryCount, err := buf.ReadVarInt()
			if err != nil {
				return err
			}
			p.ArrayOfTags[i].Tags[j].Entries = make([]ns.VarInt, entryCount)
			for k := range p.ArrayOfTags[i].Tags[j].Entries {
				if p.ArrayOfTags[i].Tags[j].Entries[k], err = buf.ReadVarInt(); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (p *S2CUpdateTagsConfiguration) Write(buf *ns.PacketBuffer) error {
	if err := buf.WriteVarInt(ns.VarInt(len(p.ArrayOfTags))); err != nil {
		return err
	}
	for _, registry := range p.ArrayOfTags {
		if err := buf.WriteIdentifier(registry.Registry); err != nil {
			return err
		}
		if err := buf.WriteVarInt(ns.VarInt(len(registry.Tags))); err != nil {
			return err
		}
		for _, tag := range registry.Tags {
			if err := buf.WriteIdentifier(tag.TagName); err != nil {
				return err
			}
			if err := buf.WriteVarInt(ns.VarInt(len(tag.Entries))); err != nil {
				return err
			}
			for _, entry := range tag.Entries {
				if err := buf.WriteVarInt(entry); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// S2CSelectKnownPacks represents "Clientbound Known Packs".
//
// Informs the client of which data packs are present on the server.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Clientbound_Known_Packs
type S2CSelectKnownPacks struct {
	KnownPacks []KnownPack
}

func (p *S2CSelectKnownPacks) Read(buf *ns.PacketBuffer) error {
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

func (p *S2CSelectKnownPacks) Write(buf *ns.PacketBuffer) error {
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

// CustomReportDetail represents a single report detail entry.
type CustomReportDetail struct {
	Title       ns.String
	Description ns.String
}

// S2CCustomReportDetailsConfiguration represents "Custom Report Details (configuration)".
//
// Contains a list of key-value text entries that are included in any crash or disconnection report.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Custom_Report_Details_(Configuration)
type S2CCustomReportDetailsConfiguration struct {
	Details []CustomReportDetail
}

func (p *S2CCustomReportDetailsConfiguration) Read(buf *ns.PacketBuffer) error {
	count, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	p.Details = make([]CustomReportDetail, count)
	for i := range p.Details {
		if p.Details[i].Title, err = buf.ReadString(128); err != nil {
			return err
		}
		if p.Details[i].Description, err = buf.ReadString(4096); err != nil {
			return err
		}
	}
	return nil
}

func (p *S2CCustomReportDetailsConfiguration) Write(buf *ns.PacketBuffer) error {
	if err := buf.WriteVarInt(ns.VarInt(len(p.Details))); err != nil {
		return err
	}
	for _, detail := range p.Details {
		if err := buf.WriteString(detail.Title); err != nil {
			return err
		}
		if err := buf.WriteString(detail.Description); err != nil {
			return err
		}
	}
	return nil
}

// ServerLink represents a server link entry.
type ServerLink struct {
	IsBuiltIn ns.Boolean
	// If IsBuiltIn is true, this is a VarInt for built-in label type.
	// If IsBuiltIn is false, this is a TextComponent for custom label.
	BuiltInLabel ns.VarInt
	CustomLabel  ns.TextComponent
	Url          ns.String
}

// S2CServerLinksConfiguration represents "Server Links (configuration)".
//
// Contains a list of links that the vanilla client will display in the pause menu.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Server_Links_(Configuration)
type S2CServerLinksConfiguration struct {
	Links []ServerLink
}

func (p *S2CServerLinksConfiguration) Read(buf *ns.PacketBuffer) error {
	count, err := buf.ReadVarInt()
	if err != nil {
		return err
	}
	p.Links = make([]ServerLink, count)
	for i := range p.Links {
		if p.Links[i].IsBuiltIn, err = buf.ReadBool(); err != nil {
			return err
		}
		if p.Links[i].IsBuiltIn {
			if p.Links[i].BuiltInLabel, err = buf.ReadVarInt(); err != nil {
				return err
			}
		} else {
			if p.Links[i].CustomLabel, err = buf.ReadTextComponent(); err != nil {
				return err
			}
		}
		if p.Links[i].Url, err = buf.ReadString(32767); err != nil {
			return err
		}
	}
	return nil
}

func (p *S2CServerLinksConfiguration) Write(buf *ns.PacketBuffer) error {
	if err := buf.WriteVarInt(ns.VarInt(len(p.Links))); err != nil {
		return err
	}
	for _, link := range p.Links {
		if err := buf.WriteBool(link.IsBuiltIn); err != nil {
			return err
		}
		if link.IsBuiltIn {
			if err := buf.WriteVarInt(link.BuiltInLabel); err != nil {
				return err
			}
		} else {
			if err := buf.WriteTextComponent(link.CustomLabel); err != nil {
				return err
			}
		}
		if err := buf.WriteString(link.Url); err != nil {
			return err
		}
	}
	return nil
}

// S2CClearDialogConfiguration represents "Clear Dialog (configuration)".
//
// Removes the current dialog screen and switches back to the previous one.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Clear_Dialog_(Configuration)
type S2CClearDialogConfiguration struct{}

func (p *S2CClearDialogConfiguration) Read(*ns.PacketBuffer) error  { return nil }
func (p *S2CClearDialogConfiguration) Write(*ns.PacketBuffer) error { return nil }

// S2CShowDialogConfiguration represents "Show Dialog (configuration)".
//
// Show a custom dialog screen to the client.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Show_Dialog_(Configuration)
type S2CShowDialogConfiguration struct {
	// Inline definition as described at Java Edition protocol/Registry data#Dialog.
	Dialog nbt.Tag
}

func (p *S2CShowDialogConfiguration) Read(buf *ns.PacketBuffer) error {
	// read remaining bytes as NBT
	remaining, err := buf.ReadByteArray(1048576)
	if err != nil {
		return err
	}
	p.Dialog, err = nbt.DecodeNetwork(remaining)
	return err
}

func (p *S2CShowDialogConfiguration) Write(buf *ns.PacketBuffer) error {
	data, err := nbt.EncodeNetwork(p.Dialog)
	if err != nil {
		return err
	}
	return buf.WriteFixedByteArray(data)
}

// S2CCodeOfConduct represents "Code of Conduct".
//
// Show the client the server Code of Conduct.
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Code_Of_Conduct
type S2CCodeOfConduct struct {
	// Code of Conduct of the server.
	Codeofconduct ns.String
}

func (p *S2CCodeOfConduct) Read(buf *ns.PacketBuffer) error {
	var err error
	p.Codeofconduct, err = buf.ReadString(32767)
	return err
}

func (p *S2CCodeOfConduct) Write(buf *ns.PacketBuffer) error {
	return buf.WriteString(p.Codeofconduct)
}
