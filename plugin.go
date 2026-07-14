package minego

import (
	"context"
	"fmt"
)

// Plugin is configured while a Bot is constructed, before any network
// connection is made. Register must not perform blocking network work.
type Plugin interface {
	Register(*PluginContext) error
}

// PluginContext exposes normalized services and an explicitly named raw wire
// escape hatch. Raw packets remain version-specific and should be capability
// gated by the plugin.
type PluginContext struct{ bot *Bot }

func (p *PluginContext) Version() VersionProvider      { return VersionProvider{bot: p.bot} }
func (p *PluginContext) World() *World                 { return p.bot.World }
func (p *PluginContext) Self() *Self                   { return p.bot.Self }
func (p *PluginContext) Entities() *Entities           { return p.bot.Entities }
func (p *PluginContext) Inventory() *Inventory         { return p.bot.Inventory }
func (p *PluginContext) Chat() *Chat                   { return p.bot.Chat }
func (p *PluginContext) Observability() *Observability { return p.bot.Observability }
func (p *PluginContext) OnDisconnect(fn func(DisconnectEvent)) func() {
	return p.bot.OnDisconnect(fn)
}
func (p *PluginContext) OnRawPacket(fn func(RawPacket)) func() { return p.bot.OnPacket(fn) }
func (p *PluginContext) SendRaw(ctx context.Context, id int32, data []byte) error {
	return p.bot.SendRaw(ctx, id, data)
}

// AddService publishes a plugin-owned service under a stable, namespaced key.
func (p *PluginContext) AddService(name string, service any) error {
	if name == "" || service == nil {
		return fmt.Errorf("minego: plugin service requires a name and value")
	}
	p.bot.servicesMu.Lock()
	defer p.bot.servicesMu.Unlock()
	if _, exists := p.bot.services[name]; exists {
		return fmt.Errorf("minego: plugin service %q already registered", name)
	}
	p.bot.services[name] = service
	return nil
}

// Service returns a plugin-published service. Callers type assert the result to
// the plugin's documented interface.
func (b *Bot) Service(name string) (any, bool) {
	b.servicesMu.RLock()
	defer b.servicesMu.RUnlock()
	service, ok := b.services[name]
	return service, ok
}

// VersionProvider lets plugins query version metadata without accessing the
// underlying pack or generated registries.
type VersionProvider struct{ bot *Bot }

func (v VersionProvider) Selected() bool                { return v.bot.Version().Name != "" }
func (v VersionProvider) Descriptor() Version           { return v.bot.Version() }
func (v VersionProvider) Supports(feature Feature) bool { return v.bot.Supports(feature) }
