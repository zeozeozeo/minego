package minego

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/zeozeozeo/minego/internal/data/packet_ids"
	"github.com/zeozeozeo/minego/internal/data/packets"
	pauth "github.com/zeozeozeo/minego/internal/protocol/auth"
	jp "github.com/zeozeozeo/minego/internal/protocol/java_protocol"
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
	"github.com/zeozeozeo/minego/internal/protocol/java_protocol/session_server"
	"github.com/zeozeozeo/minego/version"
)

func jpVarInt(v int32) ns.VarInt { return ns.VarInt(v) }

func (b *Bot) Connect(ctx context.Context) error {
	select {
	case <-b.done:
		return errors.New("minego: bot is closed")
	default:
	}
	b.mu.Lock()
	if b.connected {
		b.mu.Unlock()
		return ErrAlreadyConnected
	}
	b.mu.Unlock()
	if err := b.authenticate(ctx); err != nil {
		return fmt.Errorf("authenticate: %w", err)
	}
	host, port, address, err := resolveAddress(ctx, b.cfg.Address)
	if err != nil {
		return err
	}
	d := net.Dialer{Timeout: b.cfg.DialTimeout}
	conn, err := d.DialContext(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("connect %s: %w", address, err)
	}
	b.client = jp.NewTCPClient()
	b.client.SetConn(jp.NewConn(conn))
	b.client.SetState(jp.StateHandshake)
	b.ctx, b.cancel = context.WithCancel(context.Background())
	b.mu.Lock()
	b.connected = true
	b.mu.Unlock()
	portN, _ := strconv.Atoi(port)
	if err = b.client.WritePacket(&packets.C2SIntention{ProtocolVersion: ns.VarInt(b.pack.Protocol()), ServerAddress: ns.String(host), ServerPort: ns.Uint16(portN), Intent: 2}); err != nil {
		b.Close()
		return err
	}
	b.client.SetState(jp.StateLogin)
	uuid, _ := ns.UUIDFromString(b.uuid)
	if err = b.client.WritePacket(&packets.C2SHello{Name: ns.String(b.username), PlayerUuid: uuid}); err != nil {
		b.Close()
		return err
	}
	go b.readLoop()
	return nil
}

func (b *Bot) authenticate(ctx context.Context) error {
	if b.cfg.Auth.Mode == AuthOffline {
		b.username = b.cfg.Auth.Username
		b.uuid = offlineUUID(b.username).String()
		return nil
	}
	client := pauth.NewClient(pauth.AuthClientConfig{ClientID: b.cfg.Auth.ClientID, Username: b.cfg.Auth.Username, TokenStore: b.cfg.Auth.TokenStore, HTTPClient: b.cfg.HTTPClient, DeviceCode: b.cfg.Auth.DeviceCode})
	data, err := client.Login(ctx)
	if err != nil {
		return err
	}
	b.username, b.uuid, b.accessToken = data.Username, data.UUID, data.AccessToken
	b.chatCert, err = pauth.FetchMojangCertificate(b.accessToken)
	if err != nil {
		return fmt.Errorf("fetch chat certificate: %w", err)
	}
	return nil
}

func offlineUUID(name string) ns.UUID {
	sum := md5.Sum([]byte("OfflinePlayer:" + name))
	sum[6] = (sum[6] & 0x0f) | 0x30
	sum[8] = (sum[8] & 0x3f) | 0x80
	return ns.UUID(sum)
}

func resolveAddress(ctx context.Context, address string) (host, port, resolved string, err error) {
	host, port, err = net.SplitHostPort(address)
	if err == nil {
		return host, port, address, nil
	}
	host = address
	_, records, e := net.DefaultResolver.LookupSRV(ctx, "minecraft", "tcp", host)
	if e == nil && len(records) > 0 {
		port = strconv.Itoa(int(records[0].Port))
		return host, port, net.JoinHostPort(strings.TrimSuffix(records[0].Target, "."), port), nil
	}
	port = "25565"
	return host, port, net.JoinHostPort(host, port), nil
}

func (b *Bot) readLoop() {
	for {
		wire, err := b.client.ReadWirePacket()
		if err != nil {
			b.mu.Lock()
			b.connected = false
			b.mu.Unlock()
			b.finish(DisconnectEvent{Reason: "connection ended", Err: err})
			return
		}
		state := version.State(b.client.State())
		b.onPacket.emit(RawPacket{State: state, ID: int32(wire.PacketID), Data: append([]byte(nil), wire.Data...)})
		var hErr error
		switch b.client.State() {
		case jp.StateLogin:
			hErr = b.handleLogin(wire)
		case jp.StateConfiguration:
			hErr = b.handleConfiguration(wire)
		case jp.StatePlay:
			hErr = b.handlePlay(wire)
		}
		if hErr != nil {
			b.logError("packet handling failed", hErr)
			_ = b.client.Close()
			b.finish(DisconnectEvent{Reason: "protocol error", Err: hErr})
			return
		}
	}
}

func (b *Bot) handleLogin(w *jp.WirePacket) error {
	switch w.PacketID {
	case packet_ids.S2CLoginCompressionID:
		var p packets.S2CLoginCompression
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		b.client.SetCompressionThreshold(int(p.Threshold))
	case packet_ids.S2CHelloID:
		return b.handleEncryption(w)
	case packet_ids.S2CCustomQueryID:
		var p packets.S2CCustomQuery
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		// Unknown login plugin channels must receive an explicit unsuccessful
		// response or the server will keep the connection in login state.
		return b.client.WritePacket(&packets.C2SCustomQueryAnswer{MessageId: p.MessageId})
	case packet_ids.S2CLoginFinishedID:
		var p packets.S2CLoginFinished
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		b.username = string(p.Profile.Name)
		b.uuid = p.Profile.UUID.String()
		if err := b.client.WritePacket(&packets.C2SLoginAcknowledged{}); err != nil {
			return err
		}
		b.client.SetState(jp.StateConfiguration)
		return b.sendClientInfo()
	case packet_ids.S2CLoginDisconnectID:
		var p packets.S2CLoginDisconnectLogin
		_ = w.ReadInto(&p)
		return fmt.Errorf("login disconnected: %s", p.Reason.String())
	}
	return nil
}

func (b *Bot) handleEncryption(w *jp.WirePacket) error {
	var p packets.S2CHello
	if err := w.ReadInto(&p); err != nil {
		return err
	}
	if b.accessToken == "" && bool(p.ShouldAuthenticate) {
		return fmt.Errorf("server requires online authentication")
	}
	enc := b.client.Conn().Encryption()
	secret, err := enc.GenerateSharedSecret()
	if err != nil {
		return err
	}
	a, err := enc.EncryptWithPublicKey(p.PublicKey, secret)
	if err != nil {
		return err
	}
	token, err := enc.EncryptWithPublicKey(p.PublicKey, p.VerifyToken)
	if err != nil {
		return err
	}
	if b.accessToken != "" && bool(p.ShouldAuthenticate) {
		if err := session_server.NewSessionServerClient().Join(b.accessToken, b.uuid, string(p.ServerId), secret, p.PublicKey); err != nil {
			return err
		}
	}
	if err := b.client.WritePacket(&packets.C2SKey{SharedSecret: a, VerifyToken: token}); err != nil {
		return err
	}
	return enc.EnableEncryption()
}

func (b *Bot) sendClientInfo() error {
	buf := ns.NewWriter()
	if err := buf.WriteString(ns.String(b.cfg.Brand)); err != nil {
		return err
	}
	if err := b.client.WritePacket(&packets.C2SCustomPayloadConfiguration{Channel: "minecraft:brand", Data: buf.Bytes()}); err != nil {
		return err
	}
	return b.client.WritePacket(&packets.C2SClientInformationConfiguration{Locale: ns.String(b.cfg.Locale), ViewDistance: ns.Int8(b.cfg.ViewDistance), ChatMode: 0, ChatColors: true, DisplayedSkinParts: 0x7f, MainHand: 1, AllowServerListings: true, ParticleStatus: 2})
}

func (b *Bot) handleConfiguration(w *jp.WirePacket) error {
	switch w.PacketID {
	case packet_ids.S2CFinishConfigurationID:
		if err := b.client.WritePacket(&packets.C2SFinishConfiguration{}); err != nil {
			return err
		}
		b.client.SetState(jp.StatePlay)
		if b.chatCert != nil {
			return b.sendChatSession()
		}
	case packet_ids.S2CKeepAliveConfigurationID:
		var p packets.S2CKeepAliveConfiguration
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		return b.client.WritePacket(&packets.C2SKeepAliveConfiguration{KeepAliveId: p.KeepAliveId})
	case packet_ids.S2CPingConfigurationID:
		var p packets.S2CPingConfiguration
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		return b.client.WritePacket(&packets.C2SPongConfiguration{Id: p.Id})
	case packet_ids.S2CSelectKnownPacksID:
		return b.client.WritePacket(&packets.C2SSelectKnownPacks{})
	case packet_ids.S2CCustomPayloadConfigurationID:
		var p packets.S2CCustomPayloadConfiguration
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		if p.Channel == "minecraft:register" {
			return b.client.WritePacket(&packets.C2SCustomPayloadConfiguration{Channel: p.Channel, Data: p.Data})
		}
	case packet_ids.S2CResourcePackPushConfigurationID:
		var p packets.S2CResourcePackPushConfiguration
		if err := w.ReadInto(&p); err != nil {
			return err
		}
		result := ns.VarInt(1)
		if b.cfg.ResourcePackPolicy == ResourcePackAccept {
			result = 3
		}
		if err := b.client.WritePacket(&packets.C2SResourcePackConfiguration{Uuid: p.Uuid, Result: result}); err != nil {
			return err
		}
		if b.cfg.ResourcePackPolicy == ResourcePackAccept {
			return b.client.WritePacket(&packets.C2SResourcePackConfiguration{Uuid: p.Uuid, Result: 0})
		}
		return nil
	case packet_ids.S2CCodeOfConductID:
		return b.client.WritePacket(&packets.C2SAcceptCodeOfConduct{})
	case packet_ids.S2CDisconnectConfigurationID:
		var p packets.S2CDisconnectConfiguration
		_ = w.ReadInto(&p)
		return fmt.Errorf("configuration disconnected: %s", p.Reason.String())
	}
	return nil
}

func (b *Bot) send(ctx context.Context, p jp.Packet) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-b.done:
		return ErrNotConnected
	default:
	}
	b.mu.RLock()
	ok := b.connected
	b.mu.RUnlock()
	if !ok {
		return ErrNotConnected
	}
	return b.client.WritePacket(p)
}
func (b *Bot) sendWire(ctx context.Context, p *jp.WirePacket) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-b.done:
		return ErrNotConnected
	default:
	}
	return b.client.WriteWirePacket(p)
}

func (b *Bot) tickLoop() {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-b.ctx.Done():
			return
		case <-ticker.C:
			b.Navigator.tick()
		}
	}
}
