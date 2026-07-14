package java_protocol

import (
	"net"

	"github.com/zeozeozeo/minego/internal/protocol/crypto"
)

// Conn wraps a net.Conn with optional encryption.
// It implements io.Reader and io.Writer, transparently handling
// encryption/decryption when enabled.
type Conn struct {
	conn       net.Conn
	encryption *crypto.Encryption
}

// NewConn creates a new Conn wrapping the given net.Conn.
func NewConn(conn net.Conn) *Conn {
	return &Conn{
		conn:       conn,
		encryption: crypto.NewEncryption(),
	}
}

// Read implements io.Reader. If encryption is enabled, data is decrypted.
func (c *Conn) Read(p []byte) (int, error) {
	n, err := c.conn.Read(p)
	if err != nil {
		return n, err
	}

	if c.encryption.IsEnabled() {
		decrypted := c.encryption.Decrypt(p[:n])
		copy(p[:n], decrypted)
	}

	return n, nil
}

// Write implements io.Writer. If encryption is enabled, data is encrypted.
func (c *Conn) Write(p []byte) (int, error) {
	data := p
	if c.encryption.IsEnabled() {
		data = c.encryption.Encrypt(p)
	}

	return c.conn.Write(data)
}

// Close closes the underlying connection.
func (c *Conn) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// NetConn returns the underlying net.Conn.
func (c *Conn) NetConn() net.Conn {
	return c.conn
}

// Encryption returns the encryption instance for configuration.
func (c *Conn) Encryption() *crypto.Encryption {
	return c.encryption
}

// LocalAddr returns the local network address.
func (c *Conn) LocalAddr() net.Addr {
	if c.conn != nil {
		return c.conn.LocalAddr()
	}
	return nil
}

// RemoteAddr returns the remote network address.
func (c *Conn) RemoteAddr() net.Addr {
	if c.conn != nil {
		return c.conn.RemoteAddr()
	}
	return nil
}
