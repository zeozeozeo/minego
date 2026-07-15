package minego

import (
	"fmt"
	"strings"
	"sync"

	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

// Player is a tab-list entry linked to its currently spawned entity when one
// is known. EntityID is zero while the player is not in tracking range.
type Player struct {
	UUID     string
	Name     string
	EntityID int32
	Latency  int32
	GameMode int32
}

type PlayerChange struct {
	Kind   string
	Player Player
}

type Players struct {
	mu       sync.RWMutex
	values   map[string]Player
	onChange event[PlayerChange]
}

func newPlayers() *Players { return &Players{values: make(map[string]Player)} }
func (p *Players) All() []Player {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]Player, 0, len(p.values))
	for _, value := range p.values {
		out = append(out, value)
	}
	return out
}
func (p *Players) Get(uuid string) (Player, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	v, ok := p.values[uuid]
	return v, ok
}
func (p *Players) ByName(name string) (Player, bool)     { return p.find(name, "") }
func (p *Players) OnChange(fn func(PlayerChange)) func() { return p.onChange.subscribe(fn) }
func (p *Players) find(name, uuid string) (Player, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if uuid != "" {
		v, ok := p.values[uuid]
		return v, ok
	}
	for _, v := range p.values {
		if strings.EqualFold(v.Name, name) {
			return v, true
		}
	}
	return Player{}, false
}
func (p *Players) linkEntity(entity Entity) {
	p.mu.Lock()
	v, ok := p.values[entity.UUID]
	if ok {
		v.EntityID = entity.ID
		p.values[entity.UUID] = v
	}
	p.mu.Unlock()
	if ok {
		p.onChange.emit(PlayerChange{Kind: "update", Player: v})
	}
}

func (p *Players) upsert(v Player) {
	p.mu.Lock()
	old, ok := p.values[v.UUID]
	if v.Name == "" {
		v.Name = old.Name
	}
	if v.EntityID == 0 {
		v.EntityID = old.EntityID
	}
	p.values[v.UUID] = v
	p.mu.Unlock()
	kind := "add"
	if ok {
		kind = "update"
	}
	p.onChange.emit(PlayerChange{Kind: kind, Player: v})
}
func (p *Players) remove(uuid string) {
	p.mu.Lock()
	v, ok := p.values[uuid]
	delete(p.values, uuid)
	p.mu.Unlock()
	if ok {
		p.onChange.emit(PlayerChange{Kind: "remove", Player: v})
	}
}

func (p *Players) decodeUpdate(data []byte) error {
	b := ns.NewReader(data)
	actionsRaw, err := b.ReadInt8()
	if err != nil {
		return err
	}
	actions := uint8(actionsRaw)
	count, err := b.ReadVarInt()
	if err != nil {
		return err
	}
	pending := make([]Player, 0, int(count))
	for range int(count) {
		uuid, err := b.ReadUUID()
		if err != nil {
			return err
		}
		id := uuid.String()
		v, _ := p.Get(id)
		v.UUID = id
		if actions&1 != 0 {
			name, err := b.ReadString(16)
			if err != nil {
				return err
			}
			v.Name = string(name)
			properties, err := b.ReadVarInt()
			if err != nil {
				return err
			}
			for range int(properties) {
				if _, err = b.ReadString(32767); err != nil {
					return err
				}
				if _, err = b.ReadString(32767); err != nil {
					return err
				}
				signed, err := b.ReadBool()
				if err != nil {
					return err
				}
				if signed {
					if _, err = b.ReadString(32767); err != nil {
						return err
					}
				}
			}
		}
		if actions&2 != 0 {
			present, err := b.ReadBool()
			if err != nil {
				return err
			}
			if present {
				if _, err = b.ReadUUID(); err != nil {
					return err
				}
				if _, err = b.ReadInt64(); err != nil {
					return err
				}
				if _, err = b.ReadByteArray(1 << 20); err != nil {
					return err
				}
				if _, err = b.ReadByteArray(1 << 20); err != nil {
					return err
				}
			}
		}
		if actions&4 != 0 {
			gm, err := b.ReadVarInt()
			if err != nil {
				return err
			}
			v.GameMode = int32(gm)
		}
		if actions&8 != 0 {
			if _, err = b.ReadBool(); err != nil {
				return err
			}
		}
		if actions&16 != 0 {
			latency, err := b.ReadVarInt()
			if err != nil {
				return err
			}
			v.Latency = int32(latency)
		}
		if actions&32 != 0 {
			present, err := b.ReadBool()
			if err != nil {
				return err
			}
			if present {
				if _, err = b.ReadTextComponent(); err != nil {
					return err
				}
			}
		}
		if actions&64 != 0 {
			if _, err = b.ReadVarInt(); err != nil {
				return err
			}
		}
		if actions&128 != 0 {
			if _, err = b.ReadBool(); err != nil {
				return err
			}
		}
		pending = append(pending, v)
	}
	for _, v := range pending {
		p.upsert(v)
	}
	return nil
}
func (p *Players) decodeRemove(data []byte) error {
	b := ns.NewReader(data)
	count, err := b.ReadVarInt()
	if err != nil {
		return err
	}
	if count < 0 || count > 10000 {
		return fmt.Errorf("invalid player removal count %d", count)
	}
	for range int(count) {
		uuid, err := b.ReadUUID()
		if err != nil {
			return err
		}
		p.remove(uuid.String())
	}
	return nil
}
