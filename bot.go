package minego

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"

	pauth "github.com/zeozeozeo/minego/internal/protocol/auth"
	jp "github.com/zeozeozeo/minego/internal/protocol/java_protocol"
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
	"github.com/zeozeozeo/minego/version"
	"github.com/zeozeozeo/minego/versions"
)

// Bot owns one Minecraft connection and its synchronized service state.
type Bot struct {
	cfg                         Config
	pack                        version.Pack
	client                      *jp.TCPClient
	mu                          sync.RWMutex
	connected                   bool
	ctx                         context.Context
	cancel                      context.CancelFunc
	done                        chan struct{}
	ready                       chan struct{}
	readyOnce                   sync.Once
	doneOnce                    sync.Once
	disconnectOnce              sync.Once
	username, uuid, accessToken string
	sequence                    int32
	chatCert                    *pauth.MojangCertificateData
	chatSession                 ns.UUID
	chatIndex                   atomic.Int32
	respawning                  atomic.Bool
	actions                     *actionCoordinator
	servicesMu                  sync.RWMutex
	services                    map[string]any

	World         *World
	Self          *Self
	Entities      *Entities
	Players       *Players
	Inventory     *Inventory
	Chat          *Chat
	Navigator     *Navigator
	Miner         *Miner
	Builder       *Builder
	Crafter       *Crafter
	Elytra        *Elytra
	Observability *Observability

	onDisconnect event[DisconnectEvent]
	onPacket     event[RawPacket]
}

type RawPacket struct {
	State version.State
	ID    int32
	Data  []byte
}

func New(cfg Config) (*Bot, error) {
	if cfg.Address == "" {
		return nil, errors.New("minego: address is required")
	}
	if cfg.Locale == "" {
		cfg.Locale = "en_us"
	}
	if cfg.ViewDistance == 0 {
		cfg.ViewDistance = 12
	}
	if cfg.Brand == "" {
		cfg.Brand = "minego"
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}
	if cfg.Auth.Username == "" {
		cfg.Auth.Username = "MineGo"
	}
	if len(cfg.Auth.Username) > 16 {
		return nil, errors.New("minego: username exceeds 16 characters")
	}
	if cfg.Auth.Mode == AuthMicrosoft && cfg.Auth.ClientID == "" {
		return nil, errors.New("minego: Microsoft client ID is required")
	}
	var pack version.Pack
	if cfg.Version != "" {
		var ok bool
		pack, ok = versions.ByName(cfg.Version)
		if !ok {
			return nil, &UnsupportedVersionError{Name: cfg.Version, Supported: SupportedVersions()}
		}
	}
	b := &Bot{cfg: cfg, pack: pack, client: jp.NewTCPClient(), done: make(chan struct{}), ready: make(chan struct{}), actions: newActionCoordinator(), services: make(map[string]any)}
	b.World = newWorld(b)
	b.Self = newSelf()
	b.Entities = newEntities()
	b.Players = newPlayers()
	b.Inventory = newInventory(b)
	b.Chat = newChat(b)
	b.Navigator = newNavigator(b)
	b.Miner = newMiner(b)
	b.Builder = newBuilder(b)
	b.Crafter = newCrafter(b)
	b.Elytra = newElytra(b)
	b.Observability = newObservability(b)
	for _, plugin := range cfg.Plugins {
		if plugin == nil {
			return nil, errors.New("minego: nil plugin")
		}
		if err := plugin.Register(&PluginContext{bot: b}); err != nil {
			return nil, fmt.Errorf("register plugin: %w", err)
		}
	}
	return b, nil
}

// SupportedVersions reports the packs compiled into this build.
func SupportedVersions() []version.Descriptor { return versions.Descriptors() }

// Version reports the selected version. It is zero before Connect when
// automatic detection was requested.
func (b *Bot) Version() version.Descriptor {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.pack == nil {
		return version.Descriptor{}
	}
	return b.pack.Descriptor()
}
func (b *Bot) Supports(feature version.Feature) bool { return b.Version().Supports(feature) }
func (b *Bot) Require(feature version.Feature) error {
	descriptor := b.Version()
	if descriptor.Supports(feature) {
		return nil
	}
	return &UnsupportedFeatureError{Feature: feature, Version: descriptor}
}
func (b *Bot) WaitReady(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-b.ready:
		return nil
	case <-b.done:
		return ErrNotConnected
	}
}
func (b *Bot) Done() <-chan struct{}                        { return b.done }
func (b *Bot) OnDisconnect(fn func(DisconnectEvent)) func() { return b.onDisconnect.subscribe(fn) }
func (b *Bot) OnPacket(fn func(RawPacket)) func()           { return b.onPacket.subscribe(fn) }
func (b *Bot) SendRaw(ctx context.Context, id int32, data []byte) error {
	return b.sendWire(ctx, &jp.WirePacket{PacketID: jpVarInt(id), Data: append([]byte(nil), data...)})
}
func (b *Bot) Close() error {
	b.mu.Lock()
	if b.cancel != nil {
		b.cancel()
	}
	b.connected = false
	b.mu.Unlock()
	err := b.client.Close()
	b.finish(DisconnectEvent{Reason: "closed"})
	return err
}
func (b *Bot) block(pos BlockPos, id int32) (Block, bool) {
	b.mu.RLock()
	pack := b.pack
	b.mu.RUnlock()
	if pack == nil {
		return Block{}, false
	}
	v, ok := pack.BlockByState(id)
	if !ok {
		return Block{}, false
	}
	boxes := make([]AABB, len(v.Collision))
	for i, x := range v.Collision {
		boxes[i] = AABB{MinX: x.MinX, MinY: x.MinY, MinZ: x.MinZ, MaxX: x.MaxX, MaxY: x.MaxY, MaxZ: x.MaxZ}
	}
	return Block{Position: pos, Name: v.Name, StateID: v.StateID, Properties: v.Properties, Hardness: v.Hardness, RequiresCorrectTool: v.RequiresCorrectTool, Collision: boxes}, true
}

// HasBlock reports whether the selected version knows a block name. It is
// false before automatic version detection completes.
func (b *Bot) HasBlock(name string) bool {
	b.mu.RLock()
	pack := b.pack
	b.mu.RUnlock()
	if pack == nil {
		return false
	}
	_, ok := pack.StateID(name, nil)
	return ok
}
func (b *Bot) finish(ev DisconnectEvent) {
	if b.cancel != nil {
		b.cancel()
	}
	b.doneOnce.Do(func() { close(b.done) })
	b.disconnectOnce.Do(func() { b.onDisconnect.emit(ev) })
}
func (b *Bot) logError(msg string, err error) { b.cfg.Logger.Error(msg, "error", err) }
func (b *Bot) String() string {
	return fmt.Sprintf("MineGo{%s %s/%d}", b.cfg.Address, b.pack.Name(), b.pack.Protocol())
}
