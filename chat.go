package minego

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"strings"
	"time"

	"github.com/zeozeozeo/minego/internal/data/packets"
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

type ChatKind uint8

const (
	ChatPlayer ChatKind = iota
	ChatSystem
	ChatDisguised
)

type ChatMessage struct {
	Kind      ChatKind
	Sender    string
	Text      string
	Timestamp time.Time
	Verified  bool
}
type Chat struct {
	bot       *Bot
	onMessage event[ChatMessage]
}

func newChat(b *Bot) *Chat                            { return &Chat{bot: b} }
func (c *Chat) OnMessage(fn func(ChatMessage)) func() { return c.onMessage.subscribe(fn) }
func (c *Chat) Send(ctx context.Context, message string) error {
	if strings.HasPrefix(message, "/") {
		return c.Command(ctx, strings.TrimPrefix(message, "/"))
	}
	var salt [8]byte
	_, _ = rand.Read(salt[:])
	bits := ns.NewFixedBitSet(20)
	now := time.Now()
	saltN := int64(binary.BigEndian.Uint64(salt[:]))
	signature := ns.PrefixedOptional[ns.ByteArray]{}
	if c.bot.chatCert != nil {
		sig, err := c.bot.signChat(message, now, saltN)
		if err != nil {
			return err
		}
		signature = ns.PrefixedOptional[ns.ByteArray]{Present: true, Value: sig}
	}
	return c.bot.send(ctx, &packets.C2SChat{Message: ns.String(message), Timestamp: ns.Int64(now.UnixMilli()), Salt: ns.Int64(saltN), Signature: signature, Acknowledged: bits})
}

func (b *Bot) sendChatSession() error {
	if _, err := rand.Read(b.chatSession[:]); err != nil {
		return err
	}
	return b.client.WritePacket(&packets.C2SChatSessionUpdate{SessionId: b.chatSession, ExpiresAt: ns.Int64(b.chatCert.ExpiryTime.UnixMilli()), PublicKey: b.chatCert.PublicKeyBytes, KeySignature: b.chatCert.SignatureBytes})
}

func (b *Bot) signChat(message string, timestamp time.Time, salt int64) (ns.ByteArray, error) {
	uuid, err := ns.UUIDFromString(b.uuid)
	if err != nil {
		return nil, err
	}
	h := sha256.New()
	_ = binary.Write(h, binary.BigEndian, int32(1))
	h.Write(uuid[:])
	h.Write(b.chatSession[:])
	_ = binary.Write(h, binary.BigEndian, b.chatIndex.Add(1)-1)
	_ = binary.Write(h, binary.BigEndian, salt)
	_ = binary.Write(h, binary.BigEndian, timestamp.Unix())
	data := []byte(message)
	_ = binary.Write(h, binary.BigEndian, int32(len(data)))
	h.Write(data)
	_ = binary.Write(h, binary.BigEndian, int32(0))
	sig, err := rsa.SignPKCS1v15(rand.Reader, b.chatCert.PrivateKey, crypto.SHA256, h.Sum(nil))
	return sig, err
}
func (c *Chat) Command(ctx context.Context, command string) error {
	return c.bot.send(ctx, &packets.C2SChatCommand{Command: ns.String(strings.TrimPrefix(command, "/"))})
}
