package net_structures

import (
	"fmt"
)

// ProfileProperty represents a single property in a game profile.
type ProfileProperty struct {
	Name      String
	Value     String
	Signature PrefixedOptional[String]
}

// Decode reads a ProfileProperty from the buffer.
func (p *ProfileProperty) Decode(buf *PacketBuffer) error {
	var err error
	p.Name, err = buf.ReadString(64)
	if err != nil {
		return fmt.Errorf("failed to read property name: %w", err)
	}
	p.Value, err = buf.ReadString(32767)
	if err != nil {
		return fmt.Errorf("failed to read property value: %w", err)
	}
	if err := p.Signature.DecodeWith(buf, func(b *PacketBuffer) (String, error) {
		return b.ReadString(1024)
	}); err != nil {
		return fmt.Errorf("failed to read property signature: %w", err)
	}
	return nil
}

// Encode writes a ProfileProperty to the buffer.
func (p *ProfileProperty) Encode(buf *PacketBuffer) error {
	if err := buf.WriteString(p.Name); err != nil {
		return fmt.Errorf("failed to write property name: %w", err)
	}
	if err := buf.WriteString(p.Value); err != nil {
		return fmt.Errorf("failed to write property value: %w", err)
	}
	if err := p.Signature.EncodeWith(buf, func(b *PacketBuffer, v String) error {
		return b.WriteString(v)
	}); err != nil {
		return fmt.Errorf("failed to write property signature: %w", err)
	}
	return nil
}

// GameProfile represents a complete player profile with UUID, username, and properties.
//
// Wire format:
//
//	┌───────────────────┬─────────────────────────────────────────────────┐
//	│  UUID (16 bytes)  │  Username (String, max 16)                      │
//	├───────────────────┼─────────────────────────────────────────────────┤
//	│  Properties (VarInt length + array of ProfileProperty)              │
//	└─────────────────────────────────────────────────────────────────────┘
type GameProfile struct {
	UUID       UUID
	Username   String
	Properties PrefixedArray[ProfileProperty]
}

// Decode reads a GameProfile from the buffer.
func (p *GameProfile) Decode(buf *PacketBuffer) error {
	var err error
	p.UUID, err = buf.ReadUUID()
	if err != nil {
		return fmt.Errorf("failed to read profile uuid: %w", err)
	}
	p.Username, err = buf.ReadString(16)
	if err != nil {
		return fmt.Errorf("failed to read profile username: %w", err)
	}
	if err := p.Properties.DecodeWith(buf, func(b *PacketBuffer) (ProfileProperty, error) {
		var prop ProfileProperty
		err := prop.Decode(b)
		return prop, err
	}); err != nil {
		return fmt.Errorf("failed to read profile properties: %w", err)
	}
	return nil
}

// Encode writes a GameProfile to the buffer.
func (p *GameProfile) Encode(buf *PacketBuffer) error {
	if err := buf.WriteUUID(p.UUID); err != nil {
		return fmt.Errorf("failed to write profile uuid: %w", err)
	}
	if err := buf.WriteString(p.Username); err != nil {
		return fmt.Errorf("failed to write profile username: %w", err)
	}
	if err := p.Properties.EncodeWith(buf, func(b *PacketBuffer, v ProfileProperty) error {
		return v.Encode(b)
	}); err != nil {
		return fmt.Errorf("failed to write profile properties: %w", err)
	}
	return nil
}

// ReadGameProfile reads a GameProfile from the buffer.
func (pb *PacketBuffer) ReadGameProfile() (GameProfile, error) {
	var p GameProfile
	err := p.Decode(pb)
	return p, err
}

// WriteGameProfile writes a GameProfile to the buffer.
func (pb *PacketBuffer) WriteGameProfile(p GameProfile) error {
	return p.Encode(pb)
}

// ResolvableProfileKind indicates whether a profile is partial or complete.
type ResolvableProfileKind VarInt

const (
	ProfilePartial  ResolvableProfileKind = 0
	ProfileComplete ResolvableProfileKind = 1
)

// ResolvableProfile represents a player profile that can be either partial or complete.
//
// Wire format:
//
//	┌─────────────────────┬───────────────────────────────────────────────┐
//	│  Kind (VarInt Enum) │  Data (depends on Kind)                       │
//	└─────────────────────┴───────────────────────────────────────────────┘
//
// Partial (0): Optional username, optional UUID, optional properties, optional signature
// Complete (1): Full GameProfile + optional body/cape/elytra/model identifiers
type ResolvableProfile struct {
	Kind ResolvableProfileKind

	// partial profile fields (Kind = 0)
	PartialUsername   PrefixedOptional[String]
	PartialUUID       PrefixedOptional[UUID]
	PartialProperties PrefixedOptional[PrefixedArray[ProfileProperty]]
	PartialSignature  PrefixedOptional[String]

	// complete profile fields (Kind = 1)
	CompleteProfile GameProfile
	BodyModel       PrefixedOptional[Identifier]
	CapeModel       PrefixedOptional[Identifier]
	ElytraModel     PrefixedOptional[Identifier]
	SkinModel       PrefixedOptional[VarInt] // enum: 0=wide, 1=slim
}

// NewPartialProfile creates a partial resolvable profile.
func NewPartialProfile() *ResolvableProfile {
	return &ResolvableProfile{Kind: ProfilePartial}
}

// NewCompleteProfile creates a complete resolvable profile from a game profile.
func NewCompleteProfile(profile GameProfile) *ResolvableProfile {
	return &ResolvableProfile{
		Kind:            ProfileComplete,
		CompleteProfile: profile,
	}
}

// Decode reads a ResolvableProfile from the buffer.
func (p *ResolvableProfile) Decode(buf *PacketBuffer) error {
	kind, err := buf.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read resolvable profile kind: %w", err)
	}
	p.Kind = ResolvableProfileKind(kind)

	switch p.Kind {
	case ProfilePartial:
		if err := p.PartialUsername.DecodeWith(buf, func(b *PacketBuffer) (String, error) {
			return b.ReadString(16)
		}); err != nil {
			return fmt.Errorf("failed to read partial username: %w", err)
		}
		if err := p.PartialUUID.DecodeWith(buf, func(b *PacketBuffer) (UUID, error) {
			return b.ReadUUID()
		}); err != nil {
			return fmt.Errorf("failed to read partial uuid: %w", err)
		}
		if err := p.PartialProperties.DecodeWith(buf, func(b *PacketBuffer) (PrefixedArray[ProfileProperty], error) {
			var props PrefixedArray[ProfileProperty]
			err := props.DecodeWith(b, func(b2 *PacketBuffer) (ProfileProperty, error) {
				var prop ProfileProperty
				err := prop.Decode(b2)
				return prop, err
			})
			return props, err
		}); err != nil {
			return fmt.Errorf("failed to read partial properties: %w", err)
		}
		if err := p.PartialSignature.DecodeWith(buf, func(b *PacketBuffer) (String, error) {
			return b.ReadString(1024)
		}); err != nil {
			return fmt.Errorf("failed to read partial signature: %w", err)
		}

	case ProfileComplete:
		if err := p.CompleteProfile.Decode(buf); err != nil {
			return fmt.Errorf("failed to read complete profile: %w", err)
		}
		if err := p.BodyModel.DecodeWith(buf, func(b *PacketBuffer) (Identifier, error) {
			return b.ReadIdentifier()
		}); err != nil {
			return fmt.Errorf("failed to read body model: %w", err)
		}
		if err := p.CapeModel.DecodeWith(buf, func(b *PacketBuffer) (Identifier, error) {
			return b.ReadIdentifier()
		}); err != nil {
			return fmt.Errorf("failed to read cape model: %w", err)
		}
		if err := p.ElytraModel.DecodeWith(buf, func(b *PacketBuffer) (Identifier, error) {
			return b.ReadIdentifier()
		}); err != nil {
			return fmt.Errorf("failed to read elytra model: %w", err)
		}
		if err := p.SkinModel.DecodeWith(buf, func(b *PacketBuffer) (VarInt, error) {
			return b.ReadVarInt()
		}); err != nil {
			return fmt.Errorf("failed to read skin model: %w", err)
		}

	default:
		return fmt.Errorf("unknown resolvable profile kind: %d", p.Kind)
	}
	return nil
}

// Encode writes a ResolvableProfile to the buffer.
func (p *ResolvableProfile) Encode(buf *PacketBuffer) error {
	if err := buf.WriteVarInt(VarInt(p.Kind)); err != nil {
		return fmt.Errorf("failed to write resolvable profile kind: %w", err)
	}

	switch p.Kind {
	case ProfilePartial:
		if err := p.PartialUsername.EncodeWith(buf, func(b *PacketBuffer, v String) error {
			return b.WriteString(v)
		}); err != nil {
			return fmt.Errorf("failed to write partial username: %w", err)
		}
		if err := p.PartialUUID.EncodeWith(buf, func(b *PacketBuffer, v UUID) error {
			return b.WriteUUID(v)
		}); err != nil {
			return fmt.Errorf("failed to write partial uuid: %w", err)
		}
		if err := p.PartialProperties.EncodeWith(buf, func(b *PacketBuffer, v PrefixedArray[ProfileProperty]) error {
			return v.EncodeWith(b, func(b2 *PacketBuffer, prop ProfileProperty) error {
				return prop.Encode(b2)
			})
		}); err != nil {
			return fmt.Errorf("failed to write partial properties: %w", err)
		}
		if err := p.PartialSignature.EncodeWith(buf, func(b *PacketBuffer, v String) error {
			return b.WriteString(v)
		}); err != nil {
			return fmt.Errorf("failed to write partial signature: %w", err)
		}

	case ProfileComplete:
		if err := p.CompleteProfile.Encode(buf); err != nil {
			return fmt.Errorf("failed to write complete profile: %w", err)
		}
		if err := p.BodyModel.EncodeWith(buf, func(b *PacketBuffer, v Identifier) error {
			return b.WriteIdentifier(v)
		}); err != nil {
			return fmt.Errorf("failed to write body model: %w", err)
		}
		if err := p.CapeModel.EncodeWith(buf, func(b *PacketBuffer, v Identifier) error {
			return b.WriteIdentifier(v)
		}); err != nil {
			return fmt.Errorf("failed to write cape model: %w", err)
		}
		if err := p.ElytraModel.EncodeWith(buf, func(b *PacketBuffer, v Identifier) error {
			return b.WriteIdentifier(v)
		}); err != nil {
			return fmt.Errorf("failed to write elytra model: %w", err)
		}
		if err := p.SkinModel.EncodeWith(buf, func(b *PacketBuffer, v VarInt) error {
			return b.WriteVarInt(v)
		}); err != nil {
			return fmt.Errorf("failed to write skin model: %w", err)
		}

	default:
		return fmt.Errorf("unknown resolvable profile kind: %d", p.Kind)
	}
	return nil
}

// ReadResolvableProfile reads a ResolvableProfile from the buffer.
func (pb *PacketBuffer) ReadResolvableProfile() (ResolvableProfile, error) {
	var p ResolvableProfile
	err := p.Decode(pb)
	return p, err
}

// WriteResolvableProfile writes a ResolvableProfile to the buffer.
func (pb *PacketBuffer) WriteResolvableProfile(p ResolvableProfile) error {
	return p.Encode(pb)
}
