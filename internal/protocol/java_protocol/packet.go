// The `java_protocol` package contains the core structs and functions for working with the Java Edition protocol.
//
// > The Minecraft server accepts connections from TCP clients and communicates with them using packets.
// A packet is a sequence of bytes sent over the TCP connection (note: see `net_structures.ByteArray`).
// The meaning of a packet depends both on its packet ID and the current state of the connection
// (note: each state has its own packet ID counter, so packets in different states can have the same packet ID).
// The initial state of each connection is Handshaking, and state is switched using the packets 'Handshake' and 'Login Success'."
//
// Packet format:
//
// > Packets cannot be larger than (2^21) − 1 or 2 097 151 bytes (the maximum that can be sent in a 3-byte VarInt).
// Moreover, the length field must not be longer than 3 bytes, even if the encoded value is within the limit.
// Unnecessarily long encodings at 3 bytes or below are still allowed.
// For compressed packets, this applies to the Packet Length field, i. e. the compressed length.
//
// See https://minecraft.wiki/w/Java_Edition_protocol/Packets
package java_protocol

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"

	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

// Packet is the interface that all typed packet implementations must satisfy.
// Each packet knows its ID, protocol state, and direction.
type Packet interface {
	// ID returns the packet ID for this packet type.
	ID() ns.VarInt
	// State returns the protocol state this packet belongs to.
	State() State
	// Bound returns the direction of this packet (C2S or S2C).
	Bound() Bound
	// Read deserializes the packet data from the buffer.
	Read(buf *ns.PacketBuffer) error
	// Write serializes the packet data to the buffer.
	Write(buf *ns.PacketBuffer) error
}

// State is the phase that the packet is in (handshake, status, login, configuration, play).
// This is not sent over network (server and client automatically transition phases).
type State uint8

const (
	StateHandshake State = iota
	StateStatus
	StateLogin
	StateConfiguration
	StatePlay
)

// Bound is the direction that the packet is going.
//
// Serverbound: Client -> Server (C2S)
//
// Clientbound: Server -> Client (S2C)
type Bound uint8

const (
	// Client -> Server (C2S, serverbound)
	C2S Bound = iota
	// Server -> Client (S2C, clientbound)
	S2C
)

// WirePacket represents the raw packet as it appears on the wire.
// It contains only wire-level data without typed field information.
type WirePacket struct {
	// Length is the total length of (PacketID + Data) as read from the wire.
	Length ns.VarInt
	// PacketID is the packet identifier.
	PacketID ns.VarInt
	// Data is the raw payload bytes (without the packet ID).
	Data ns.ByteArray
}

// Clone returns a deep copy of the wire packet.
func (w *WirePacket) Clone() *WirePacket {
	data := make([]byte, len(w.Data))
	copy(data, w.Data)
	return &WirePacket{
		Length:   w.Length,
		PacketID: w.PacketID,
		Data:     data,
	}
}

// ReadWirePacketFrom reads a WirePacket from the given reader.
// Handles both compressed and uncompressed packet formats based on compressionThreshold.
// Use compressionThreshold < 0 to disable compression.
func ReadWirePacketFrom(r io.Reader, compressionThreshold int) (*WirePacket, error) {
	packetLength, err := ns.DecodeVarInt(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read packet length: %w", err)
	}

	data := make([]byte, packetLength)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, fmt.Errorf("failed to read packet data: %w", err)
	}

	reader := bytes.NewReader(data)

	if compressionThreshold >= 0 {
		return readCompressedPacket(reader, packetLength)
	}
	return readUncompressedPacket(reader, packetLength)
}

func readUncompressedPacket(reader *bytes.Reader, length ns.VarInt) (*WirePacket, error) {
	packetID, err := ns.DecodeVarInt(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read packet ID: %w", err)
	}

	remainingData, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read remaining data: %w", err)
	}

	return &WirePacket{
		Length:   length,
		PacketID: packetID,
		Data:     ns.ByteArray(remainingData),
	}, nil
}

func readCompressedPacket(reader *bytes.Reader, length ns.VarInt) (*WirePacket, error) {
	dataLength, err := ns.DecodeVarInt(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read data length: %w", err)
	}

	// dataLength == 0 means uncompressed despite compression being enabled
	if dataLength == 0 {
		return readUncompressedPacket(reader, length)
	}

	// uncompress
	compressedData, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read compressed data: %w", err)
	}
	uncompressedData, err := decompressZlib(compressedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress: %w", err)
	}

	// read uncompressed
	uncompressedReader := bytes.NewReader(uncompressedData)
	packetID, err := ns.DecodeVarInt(uncompressedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read packet ID: %w", err)
	}

	// remaining data is the packet data
	remainingData, err := io.ReadAll(uncompressedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read remaining data: %w", err)
	}
	return &WirePacket{
		Length:   length,
		PacketID: packetID,
		Data:     ns.ByteArray(remainingData),
	}, nil
}

// WriteTo writes the WirePacket to the given writer.
// Handles both compressed and uncompressed packet formats based on compressionThreshold.
// Use compressionThreshold < 0 to disable compression.
//
// Compression behavior (per Minecraft protocol):
//   - If size >= threshold: packet is zlib compressed
//   - If size < threshold: packet is sent uncompressed (with Data Length = 0)
//   - The vanilla server rejects compressed packets smaller than the threshold
//
// See https://minecraft.wiki/w/Java_Edition_protocol/Packets#Packet_format
func (w *WirePacket) WriteTo(writer io.Writer, compressionThreshold int) error {
	var data []byte
	var err error
	if compressionThreshold >= 0 {
		data, err = w.toBytesCompressed(compressionThreshold)
	} else {
		data, err = w.toBytesUncompressed()
	}
	if err != nil {
		return fmt.Errorf("failed to serialize packet: %w", err)
	}
	_, err = writer.Write(data)
	return err
}

// ReadInto deserializes the wire packet's raw data into a typed Packet.
// Returns an error if the packet ID doesn't match.
func (w *WirePacket) ReadInto(p Packet) error {
	if w == nil {
		return fmt.Errorf("nil wire packet")
	}
	if w.PacketID != p.ID() {
		return fmt.Errorf("packet ID mismatch: expected 0x%02X, got 0x%02X", p.ID(), w.PacketID)
	}
	buf := ns.NewReader(w.Data)
	return p.Read(buf)
}

// ReadPacket deserializes a WirePacket into a typed Packet using generics.
// This provides type-safe packet reading without manual type assertions.
//
// Example:
//
//	wire, _ := client.ReadWirePacket()
//	login, err := ReadPacket[LoginSuccessPacket](wire)
func ReadPacket[T any, PT interface {
	*T
	Packet
}](wire *WirePacket) (PT, error) {
	p := new(T)
	pt := PT(p)
	if err := wire.ReadInto(pt); err != nil {
		return nil, err
	}
	return pt, nil
}

// ToWire converts a typed Packet to a WirePacket by serializing its data.
// The resulting WirePacket can then be written to a connection via WriteTo()
// or converted to bytes via ToBytes().
func ToWire(p Packet) (*WirePacket, error) {
	buf := ns.NewWriter()
	if err := p.Write(buf); err != nil {
		return nil, fmt.Errorf("failed to serialize packet data: %w", err)
	}
	return &WirePacket{
		PacketID: p.ID(),
		Data:     buf.Bytes(),
	}, nil
}

// toBytesCompressed serializes with compression framing.
//
// Structure:
//
//	if (size >= networkCompressionThreshold)
//		packetLength: VarInt(Length of (Data Length) + length of compressed (Packet ID + Data)) +
//		dataLength: VarInt(Length of uncompressed (Packet ID + Data)) +
//		packetID: compressed(VarInt(Packet ID)) +
//		data: compressed(Data)
//	if (size < networkCompressionThreshold)
//		packetLength: VarInt(Length of (Data Length) + length of uncompressed (Packet ID + Data)) +
//		dataLength: VarInt(0) + // compressed data length is 0, which means no compression is used
//		packetID: VarInt(Packet ID) +
//		data: ByteArray(Data)
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#With_compression
func (w *WirePacket) toBytesCompressed(compressionThreshold int) ([]byte, error) {
	packetIDBytes, err := w.PacketID.ToBytes()
	if err != nil {
		return nil, err
	}
	uncompressedPayload := append(packetIDBytes, w.Data...)
	uncompressedLength := len(uncompressedPayload)

	if uncompressedLength >= compressionThreshold {
		// Compress the payload
		compressedPayload := compressZlib(uncompressedPayload)

		dataLengthBytes, err := ns.VarInt(uncompressedLength).ToBytes()
		if err != nil {
			return nil, err
		}
		packetContent := append(dataLengthBytes, compressedPayload...)
		packetLengthBytes, err := ns.VarInt(len(packetContent)).ToBytes()
		if err != nil {
			return nil, err
		}

		return append(packetLengthBytes, packetContent...), nil
	}

	// Uncompressed (below threshold)
	dataLengthBytes, err := ns.VarInt(0).ToBytes()
	if err != nil {
		return nil, err
	}
	packetContent := append(dataLengthBytes, uncompressedPayload...)
	packetLengthBytes, err := ns.VarInt(len(packetContent)).ToBytes()
	if err != nil {
		return nil, err
	}

	return append(packetLengthBytes, packetContent...), nil
}

// toBytesUncompressed serializes without compression.
//
// Structure:
//
//	packetLength: VarInt(Length of Packet ID + Data) +
//	packetID: VarInt(Packet ID) +
//	data: ByteArray(Data)
//
// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Without_compression
func (w *WirePacket) toBytesUncompressed() ([]byte, error) {
	packetIDBytes, err := w.PacketID.ToBytes()
	if err != nil {
		return nil, err
	}

	payload := append(packetIDBytes, w.Data...)
	packetLengthBytes, err := ns.VarInt(len(payload)).ToBytes()
	if err != nil {
		return nil, err
	}

	return append(packetLengthBytes, payload...), nil
}

func compressZlib(data []byte) []byte {
	compressedData := bytes.NewBuffer(nil)
	writer := zlib.NewWriter(compressedData)
	_, _ = writer.Write(data)
	_ = writer.Close()
	return compressedData.Bytes()
}

func decompressZlib(data []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}
