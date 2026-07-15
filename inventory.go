package minego

import (
	"context"
	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/items"
	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/packets"
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
	"hash/crc32"
	"sort"
	"sync"
)

type ItemStack struct {
	Name       string
	ID         int32
	Count      int32
	Components map[int32][]byte
}
type WindowSnapshot struct {
	ID      int32
	Type    int32
	Title   string
	StateID int32
	Slots   []ItemStack
	Carried ItemStack
}
type WindowChange struct{ Window WindowSnapshot }
type Inventory struct {
	bot      *Bot
	mu       sync.RWMutex
	slots    []ItemStack
	selected int
	window   WindowSnapshot
	onWindow event[WindowChange]
}

func newInventory(b *Bot) *Inventory { return &Inventory{bot: b, slots: make([]ItemStack, 46)} }
func (i *Inventory) Slots() []ItemStack {
	i.mu.RLock()
	defer i.mu.RUnlock()
	out := append([]ItemStack(nil), i.slots...)
	for index := range out {
		out[index] = cloneStack(out[index])
	}
	return out
}
func (i *Inventory) Window() WindowSnapshot {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return cloneWindow(i.window)
}
func (i *Inventory) OnWindowChange(fn func(WindowChange)) func() { return i.onWindow.subscribe(fn) }
func (i *Inventory) Selected() int                               { i.mu.RLock(); defer i.mu.RUnlock(); return i.selected }
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

func fromNetSlot(s ns.Slot) ItemStack {
	if s.Count <= 0 {
		return ItemStack{}
	}
	stack := ItemStack{Name: items.ItemName(int32(s.ItemID)), ID: int32(s.ItemID), Count: int32(s.Count)}
	if len(s.Components.Add) > 0 {
		stack.Components = make(map[int32][]byte, len(s.Components.Add))
		for _, c := range s.Components.Add {
			stack.Components[int32(c.ID)] = append([]byte(nil), c.Data...)
		}
	}
	return stack
}
func cloneStack(s ItemStack) ItemStack {
	if s.Components != nil {
		m := make(map[int32][]byte, len(s.Components))
		for k, v := range s.Components {
			m[k] = append([]byte(nil), v...)
		}
		s.Components = m
	}
	return s
}
func cloneWindow(w WindowSnapshot) WindowSnapshot {
	w.Slots = append([]ItemStack(nil), w.Slots...)
	for j := range w.Slots {
		w.Slots[j] = cloneStack(w.Slots[j])
	}
	w.Carried = cloneStack(w.Carried)
	return w
}
func (s ItemStack) hashed() ns.HashedSlot {
	if s.Count <= 0 {
		return ns.EmptyHashedSlot()
	}
	h := ns.NewHashedSlot(ns.VarInt(s.ID), ns.VarInt(s.Count))
	table := crc32.MakeTable(crc32.Castagnoli)
	ids := make([]int, 0, len(s.Components))
	for id := range s.Components {
		ids = append(ids, int(id))
	}
	sort.Ints(ids)
	for _, rawID := range ids {
		id := int32(rawID)
		data := s.Components[id]
		h.Components.Add = append(h.Components.Add, ns.HashedComponent{ID: ns.VarInt(id), Hash: ns.Int32(int32(crc32.Checksum(data, table)))})
	}
	return h
}
