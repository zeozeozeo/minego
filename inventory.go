package minego

import (
	"context"
	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/items"
	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/packets"
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
	"sync"
)

type ItemStack struct {
	Name  string
	ID    int32
	Count int32
}
type Inventory struct {
	bot      *Bot
	mu       sync.RWMutex
	slots    []ItemStack
	selected int
}

func newInventory(b *Bot) *Inventory { return &Inventory{bot: b, slots: make([]ItemStack, 46)} }
func (i *Inventory) Slots() []ItemStack {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return append([]ItemStack(nil), i.slots...)
}
func (i *Inventory) Selected() int { i.mu.RLock(); defer i.mu.RUnlock(); return i.selected }
func (i *Inventory) Select(ctx context.Context, hotbar int) error {
	if hotbar < 0 || hotbar > 8 {
		return ErrInvalidSlot
	}
	lease, err := i.bot.actions.acquire(ctx, controlInventory, priorityExplicit)
	if err != nil {
		return err
	}
	defer lease.Release()
	ctx = lease.Context(ctx)
	if err := i.bot.send(ctx, &packets.C2SSetCarriedItem{Slot: ns.Int16(hotbar)}); err != nil {
		return err
	}
	i.mu.Lock()
	i.selected = hotbar
	i.mu.Unlock()
	return nil
}
func stack(id, count int32) ItemStack {
	if count <= 0 {
		return ItemStack{}
	}
	return ItemStack{Name: items.ItemName(id), ID: id, Count: count}
}
