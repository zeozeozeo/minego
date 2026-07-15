package minego

import (
	"context"
	"fmt"
	"time"

	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/packets"
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

// Container is a handle to one server-authoritative window. Methods reject a
// stale handle after the server closes or replaces that window.
type Container struct {
	service *Containers
	ID      int32
}

type Containers struct{ bot *Bot }

func newContainers(bot *Bot) *Containers { return &Containers{bot: bot} }

func (c *Containers) Current() (*Container, bool) {
	w := c.bot.Inventory.Window()
	if w.ID == 0 {
		return nil, false
	}
	return &Container{service: c, ID: w.ID}, true
}

// Open activates a block and waits for the corresponding window lifecycle.
func (c *Containers) Open(ctx context.Context, pos BlockPos) (*Container, error) {
	return c.openAfter(ctx, func(ctx context.Context) error {
		return c.bot.Interaction.ActivateBlock(ctx, BlockInteraction{Position: pos, Face: 1, Cursor: Vec3{.5, .5, .5}, Hand: MainHand})
	})
}

// OpenEntity activates an entity-backed window, notably villager trading and
// mount inventories.
func (c *Containers) OpenEntity(ctx context.Context, entityID int32) (*Container, error) {
	return c.openAfter(ctx, func(ctx context.Context) error {
		return c.bot.Interaction.ActivateEntity(ctx, EntityInteraction{EntityID: entityID, Hand: MainHand, Reach: 4.5})
	})
}

func (c *Containers) openAfter(ctx context.Context, activate func(context.Context) error) (*Container, error) {
	lease, err := c.bot.actions.acquire(ctx, controlView|controlHands|controlWindows, priorityExplicit)
	if err != nil {
		return nil, err
	}
	defer lease.Release()
	ctx = lease.Context(ctx)
	ch := make(chan WindowSnapshot, 1)
	unsub := c.bot.Inventory.OnWindowChange(func(e WindowChange) {
		if e.Window.ID != 0 {
			select {
			case ch <- e.Window:
			default:
			}
		}
	})
	defer unsub()
	if err := activate(ctx); err != nil {
		return nil, err
	}
	t := time.NewTimer(5 * time.Second)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.bot.done:
		return nil, ErrNotConnected
	case <-t.C:
		return nil, fmt.Errorf("minego: server did not open a container")
	case w := <-ch:
		return &Container{service: c, ID: w.ID}, nil
	}
}

func (c *Container) Snapshot() (WindowSnapshot, error) {
	w := c.service.bot.Inventory.Window()
	if w.ID != c.ID {
		return WindowSnapshot{}, ErrWindowClosed
	}
	return w, nil
}

func (c *Container) Click(ctx context.Context, slot int, button int8, mode ClickMode) (WindowSnapshot, error) {
	if _, err := c.Snapshot(); err != nil {
		return WindowSnapshot{}, err
	}
	return c.service.bot.Inventory.Click(ctx, ClickOptions{WindowID: c.ID, Slot: slot, Button: button, Mode: mode})
}

func (c *Container) Transfer(ctx context.Context, source, destination, count int) error {
	if _, err := c.Snapshot(); err != nil {
		return err
	}
	return c.service.bot.Inventory.Transfer(ctx, c.ID, source, destination, count)
}

func (c *Container) QuickMove(ctx context.Context, slot int) error {
	_, err := c.Click(ctx, slot, 0, ClickQuickMove)
	return err
}

func (c *Container) Close(ctx context.Context) error {
	if _, err := c.Snapshot(); err != nil {
		return err
	}
	return c.service.bot.Inventory.Close(ctx)
}

func (c *Container) Property(id int16) (int16, bool) {
	w, err := c.Snapshot()
	if err != nil {
		return 0, false
	}
	v, ok := w.Properties[id]
	return v, ok
}

// FurnaceState normalizes the four vanilla furnace properties.
type FurnaceState struct{ BurnTime, FuelTime, CookTime, CookTimeTotal int16 }

func (c *Container) Furnace() FurnaceState {
	w, err := c.Snapshot()
	if err != nil {
		return FurnaceState{}
	}
	return FurnaceState{BurnTime: w.Properties[0], FuelTime: w.Properties[1], CookTime: w.Properties[2], CookTimeTotal: w.Properties[3]}
}

func (c *Container) PutFurnaceInput(ctx context.Context, sourceSlot, count int) error {
	return c.Transfer(ctx, sourceSlot, 0, count)
}
func (c *Container) PutFurnaceFuel(ctx context.Context, sourceSlot, count int) error {
	return c.Transfer(ctx, sourceSlot, 1, count)
}
func (c *Container) TakeFurnaceOutput(ctx context.Context) error { return c.QuickMove(ctx, 2) }

// ChooseEnchantment clicks one of the three enchantment choices.
func (c *Container) ChooseEnchantment(ctx context.Context, choice int) error {
	if choice < 0 || choice > 2 {
		return fmt.Errorf("minego: enchantment choice must be between 0 and 2")
	}
	return c.service.bot.send(ctx, &packets.C2SContainerButtonClick{WindowId: ns.VarInt(c.ID), ButtonId: ns.VarInt(choice)})
}

func (c *Container) Rename(ctx context.Context, name string) error {
	if len(name) > 50 {
		return fmt.Errorf("minego: anvil name exceeds 50 characters")
	}
	return c.service.bot.send(ctx, &packets.C2SRenameItem{ItemName: ns.String(name)})
}

// SelectTrade selects a villager offer. Inputs may then be filled with
// Transfer and the result taken from slot 2.
func (c *Container) SelectTrade(ctx context.Context, index int) error {
	if index < 0 {
		return fmt.Errorf("minego: trade index cannot be negative")
	}
	return c.service.bot.send(ctx, &packets.C2SSelectTrade{SelectedSlot: ns.VarInt(index)})
}

func (c *Container) MerchantOffers() []byte {
	w, err := c.Snapshot()
	if err != nil {
		return nil
	}
	return append([]byte(nil), w.Offers...)
}
