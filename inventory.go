package minego

import (
	"context"
	"fmt"
	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/items"
	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/packets"
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
	"hash/crc32"
	"sort"
	"sync"
	"time"
)

type ItemStack struct {
	Name       string
	ID         int32
	Count      int32
	Components map[int32][]byte
}
type WindowSnapshot struct {
	ID         int32
	Type       int32
	Title      string
	StateID    int32
	Slots      []ItemStack
	Carried    ItemStack
	Properties map[int16]int16
	Offers     []byte
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
	w.Offers = append([]byte(nil), w.Offers...)
	if w.Properties != nil {
		properties := make(map[int16]int16, len(w.Properties))
		for k, v := range w.Properties {
			properties[k] = v
		}
		w.Properties = properties
	}
	return w
}

// ClickMode is the vanilla container click mode.
type ClickMode int32

const (
	ClickPickup ClickMode = iota
	ClickQuickMove
	ClickSwap
	ClickClone
	ClickThrow
	ClickQuickCraft
	ClickPickupAll
)

type ClickOptions struct {
	WindowID int32
	Slot     int
	Button   int8
	Mode     ClickMode
	Timeout  time.Duration
}

// Click sends a state-ID-protected click and waits for an authoritative
// update. WindowID zero targets the player inventory.
func (i *Inventory) Click(ctx context.Context, opt ClickOptions) (WindowSnapshot, error) {
	if opt.Slot < -999 || opt.Slot > 32767 {
		return WindowSnapshot{}, ErrInvalidSlot
	}
	if opt.Mode < ClickPickup || opt.Mode > ClickPickupAll {
		return WindowSnapshot{}, fmt.Errorf("minego: invalid click mode %d", opt.Mode)
	}
	if opt.Timeout <= 0 {
		opt.Timeout = 5 * time.Second
	}
	lease, err := i.bot.actions.acquire(ctx, controlInventory|controlWindows, priorityExplicit)
	if err != nil {
		return WindowSnapshot{}, err
	}
	defer lease.Release()
	ctx = lease.Context(ctx)
	w := i.windowByID(opt.WindowID)
	if opt.WindowID != 0 && w.ID != opt.WindowID {
		return WindowSnapshot{}, ErrNoWindow
	}
	p := &packets.C2SContainerClick{WindowId: ns.VarInt(opt.WindowID), StateId: ns.VarInt(w.StateID), Slot: ns.Int16(opt.Slot), Button: ns.Int8(opt.Button), Mode: ns.VarInt(opt.Mode), CarriedItem: w.Carried.hashed()}
	if err := i.bot.send(ctx, p); err != nil {
		return WindowSnapshot{}, err
	}
	return i.waitWindowState(ctx, opt.WindowID, w.StateID, opt.Timeout)
}

func (i *Inventory) QuickMove(ctx context.Context, windowID int32, slot int) (WindowSnapshot, error) {
	return i.Click(ctx, ClickOptions{WindowID: windowID, Slot: slot, Mode: ClickQuickMove})
}

// Transfer moves count items using authoritative pickup clicks. A nonpositive
// count moves the entire source stack.
func (i *Inventory) Transfer(ctx context.Context, windowID int32, source, destination, count int) error {
	w := i.windowByID(windowID)
	if source < 0 || source >= len(w.Slots) || destination < 0 || destination >= len(w.Slots) {
		return ErrInvalidSlot
	}
	available := int(w.Slots[source].Count)
	if available == 0 {
		return fmt.Errorf("minego: source slot is empty")
	}
	if count <= 0 || count > available {
		count = available
	}
	if _, err := i.Click(ctx, ClickOptions{WindowID: windowID, Slot: source, Mode: ClickPickup}); err != nil {
		return err
	}
	if count == available {
		if _, err := i.Click(ctx, ClickOptions{WindowID: windowID, Slot: destination, Mode: ClickPickup}); err != nil {
			return err
		}
	} else {
		for n := 0; n < count; n++ {
			if _, err := i.Click(ctx, ClickOptions{WindowID: windowID, Slot: destination, Button: 1, Mode: ClickPickup}); err != nil {
				return err
			}
		}
	}
	if i.windowByID(windowID).Carried.Count > 0 {
		_, err := i.Click(ctx, ClickOptions{WindowID: windowID, Slot: source, Mode: ClickPickup})
		return err
	}
	return nil
}

type EquipmentSlot int

const (
	EquipHelmet  EquipmentSlot = 5
	EquipChest   EquipmentSlot = 6
	EquipLegs    EquipmentSlot = 7
	EquipBoots   EquipmentSlot = 8
	EquipOffhand EquipmentSlot = 45
)

// Equip moves an inventory item into a player equipment slot.
func (i *Inventory) Equip(ctx context.Context, inventorySlot int, equipment EquipmentSlot) error {
	if equipment != EquipHelmet && equipment != EquipChest && equipment != EquipLegs && equipment != EquipBoots && equipment != EquipOffhand {
		return ErrInvalidSlot
	}
	return i.Transfer(ctx, 0, inventorySlot, int(equipment), 0)
}

// Toss drops one item or an entire stack from a player inventory slot.
func (i *Inventory) Toss(ctx context.Context, slot int, entireStack bool) error {
	button := int8(0)
	if entireStack {
		button = 1
	}
	_, err := i.Click(ctx, ClickOptions{WindowID: 0, Slot: slot, Button: button, Mode: ClickThrow})
	return err
}

func (i *Inventory) Close(ctx context.Context) error {
	w := i.Window()
	if w.ID == 0 {
		return nil
	}
	return i.bot.send(ctx, &packets.C2SContainerClose{WindowId: ns.VarInt(w.ID)})
}

func (i *Inventory) CreativeSet(ctx context.Context, slot int, item ItemStack) error {
	if i.bot.Self.State().GameMode != 1 {
		return ErrInvalidGameMode
	}
	if slot < 0 || slot > 45 {
		return ErrInvalidSlot
	}
	return i.bot.send(ctx, &packets.C2SSetCreativeModeSlot{Slot: ns.Int16(slot), ClickedItem: item.netSlot()})
}

func (i *Inventory) windowByID(id int32) WindowSnapshot {
	i.mu.RLock()
	defer i.mu.RUnlock()
	w := cloneWindow(i.window)
	if id == 0 && len(w.Slots) == 0 {
		w = WindowSnapshot{ID: 0, Slots: append([]ItemStack(nil), i.slots...)}
	}
	return w
}

func (i *Inventory) waitWindowState(ctx context.Context, id, previous int32, timeout time.Duration) (WindowSnapshot, error) {
	if w := i.windowByID(id); w.StateID != previous {
		return w, nil
	}
	ch := make(chan WindowSnapshot, 1)
	unsub := i.OnWindowChange(func(e WindowChange) {
		if e.Window.ID == id && e.Window.StateID != previous {
			select {
			case ch <- e.Window:
			default:
			}
		}
	})
	defer unsub()
	t := time.NewTimer(timeout)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return WindowSnapshot{}, ctx.Err()
	case <-i.bot.done:
		return WindowSnapshot{}, ErrNotConnected
	case <-t.C:
		return WindowSnapshot{}, fmt.Errorf("minego: server did not acknowledge container click")
	case w := <-ch:
		return w, nil
	}
}

func (s ItemStack) netSlot() ns.Slot {
	if s.Count <= 0 {
		return ns.EmptySlot()
	}
	r := ns.NewSlot(ns.VarInt(s.ID), ns.VarInt(s.Count))
	ids := make([]int, 0, len(s.Components))
	for id := range s.Components {
		ids = append(ids, int(id))
	}
	sort.Ints(ids)
	for _, raw := range ids {
		id := int32(raw)
		r.Components.Add = append(r.Components.Add, ns.RawSlotComponent{ID: ns.VarInt(id), Data: append([]byte(nil), s.Components[id]...)})
	}
	return r
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
