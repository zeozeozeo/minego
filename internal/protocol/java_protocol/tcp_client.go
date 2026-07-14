package java_protocol

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

// TCPClient is a Minecraft protocol client connection.
type TCPClient struct {
	writeMu              sync.Mutex
	conn                 *Conn
	state                State
	compressionThreshold int

	debug  bool
	logger *log.Logger
}

// NewTCPClient creates a new TCP client.
func NewTCPClient() *TCPClient {
	return &TCPClient{
		conn:                 nil,
		state:                StateHandshake,
		compressionThreshold: -1,
		debug:                false,
		logger:               log.New(os.Stdout, "[TCPClient] ", log.LstdFlags),
	}
}

// Connect connects to a Minecraft server.
// The address can be "host", "host:port", or will use SRV records if no port specified.
// Returns the resolved host and port.
func (c *TCPClient) Connect(address string) (host string, port string, err error) {
	resolvedAddr, err := resolveMinecraftAddress(address)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve address: %w", err)
	}

	netConn, err := net.Dial("tcp", resolvedAddr)
	if err != nil {
		return "", "", fmt.Errorf("failed to connect to %s: %w", resolvedAddr, err)
	}

	c.conn = NewConn(netConn)
	return net.SplitHostPort(resolvedAddr)
}

// SetConn sets the underlying connection (for testing or server-accepted connections).
func (c *TCPClient) SetConn(conn *Conn) {
	c.conn = conn
}

// Close closes the connection.
func (c *TCPClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Conn returns the underlying Conn.
func (c *TCPClient) Conn() *Conn {
	return c.conn
}

// SetState sets the protocol state.
func (c *TCPClient) SetState(state State) {
	c.state = state
}

// State returns the current protocol state.
func (c *TCPClient) State() State {
	return c.state
}

// SetCompressionThreshold enables compression with the given threshold.
// Use -1 to disable compression.
func (c *TCPClient) SetCompressionThreshold(threshold int) {
	c.compressionThreshold = threshold
}

// CompressionThreshold returns the current compression threshold.
func (c *TCPClient) CompressionThreshold() int {
	return c.compressionThreshold
}

// EnableDebug enables or disables debug logging.
func (c *TCPClient) EnableDebug(enabled bool) {
	c.debug = enabled
}

// SetLogger sets a custom logger.
func (c *TCPClient) SetLogger(l *log.Logger) {
	c.logger = l
}

// WritePacket writes a typed Packet to the connection.
// Safe for concurrent use from multiple goroutines.
func (c *TCPClient) WritePacket(p Packet) error {
	if c.conn == nil {
		return fmt.Errorf("connection is nil")
	}

	// serialize outside the lock (CPU work, no I/O)
	wire, err := ToWire(p)
	if err != nil {
		return fmt.Errorf("failed to serialize packet: %w", err)
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	c.debugf("-> send: state=%v bound=%v id=0x%02X", p.State(), p.Bound(), int(p.ID()))

	if err := wire.WriteTo(c.conn, c.compressionThreshold); err != nil {
		return fmt.Errorf("failed to write packet: %w", err)
	}

	return nil
}

// WriteWirePacket writes a raw WirePacket to the connection.
// Safe for concurrent use from multiple goroutines.
func (c *TCPClient) WriteWirePacket(pkt *WirePacket) error {
	if c.conn == nil {
		return fmt.Errorf("connection is nil")
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	c.debugf("-> send (wire): id=0x%02X data_len=%d", int(pkt.PacketID), len(pkt.Data))

	if err := pkt.WriteTo(c.conn, c.compressionThreshold); err != nil {
		return fmt.Errorf("failed to write wire packet: %w", err)
	}

	return nil
}

// ReadWirePacket reads a raw WirePacket from the connection.
func (c *TCPClient) ReadWirePacket() (*WirePacket, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("connection is nil")
	}

	c.debugf("<- recv: waiting for packet")
	wire, err := ReadWirePacketFrom(c.conn, c.compressionThreshold)
	if err != nil {
		return nil, err
	}

	c.debugf("<- recv: id=0x%02X len=%d data_len=%d", int(wire.PacketID), int(wire.Length), len(wire.Data))
	return wire, nil
}

func (c *TCPClient) debugf(format string, args ...any) {
	if c.debug && c.logger != nil {
		c.logger.Printf(format, args...)
	}
}

// resolveMinecraftAddress resolves a Minecraft server address using SRV records if available,
// falling back to the default port 25565 if no port is specified.
func resolveMinecraftAddress(address string) (string, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		host = address
		port = ""
	}

	// if port is explicitly specified, use it directly
	if port != "" {
		return net.JoinHostPort(host, port), nil
	}

	// ...otherwise, lookup SRV _minecraft._tcp.<host>
	_, srvRecords, err := net.LookupSRV("minecraft", "tcp", host)
	if err == nil && len(srvRecords) > 0 {
		srv := srvRecords[0]
		target := strings.TrimSuffix(srv.Target, ".")
		return net.JoinHostPort(target, strconv.Itoa(int(srv.Port))), nil
	}

	// no SRV record found, use default port
	return net.JoinHostPort(host, "25565"), nil
}
