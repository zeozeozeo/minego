package minego

import "sync"

// WeatherState is the latest weather state announced by the server.
type WeatherState struct {
	Raining                 bool
	RainLevel, ThunderLevel float32
}

type Weather struct {
	mu       sync.RWMutex
	state    WeatherState
	onChange event[WeatherState]
}

func newWeather() *Weather                               { return &Weather{} }
func (w *Weather) State() WeatherState                   { w.mu.RLock(); defer w.mu.RUnlock(); return w.state }
func (w *Weather) OnChange(fn func(WeatherState)) func() { return w.onChange.subscribe(fn) }
func (w *Weather) update(fn func(*WeatherState)) {
	w.mu.Lock()
	fn(&w.state)
	v := w.state
	w.mu.Unlock()
	w.onChange.emit(v)
}

type TabListState struct{ Header, Footer string }
type TabList struct {
	mu       sync.RWMutex
	state    TabListState
	onChange event[TabListState]
}

func newTabList() *TabList                               { return &TabList{} }
func (t *TabList) State() TabListState                   { t.mu.RLock(); defer t.mu.RUnlock(); return t.state }
func (t *TabList) OnChange(fn func(TabListState)) func() { return t.onChange.subscribe(fn) }
func (t *TabList) update(v TabListState)                 { t.mu.Lock(); t.state = v; t.mu.Unlock(); t.onChange.emit(v) }

// Objective and Score preserve version-owned number-format bytes in Data.
type Objective struct {
	Name, DisplayName string
	RenderType        int32
	Data              []byte
}
type Score struct {
	Entity, Objective, DisplayName string
	Value                          int32
	Data                           []byte
}
type ScoreboardChange struct {
	Kind             string
	Objective        *Objective
	Score            *Score
	DisplaySlot      int32
	DisplayObjective string
}
type ScoreboardSnapshot struct {
	Objectives   map[string]Objective
	Scores       map[string]map[string]Score
	DisplaySlots map[int32]string
}
type Scoreboards struct {
	mu         sync.RWMutex
	objectives map[string]Objective
	scores     map[string]map[string]Score
	displays   map[int32]string
	onChange   event[ScoreboardChange]
}

func newScoreboards() *Scoreboards {
	return &Scoreboards{objectives: map[string]Objective{}, scores: map[string]map[string]Score{}, displays: map[int32]string{}}
}
func cloneObjective(v Objective) Objective { v.Data = append([]byte(nil), v.Data...); return v }
func cloneScore(v Score) Score             { v.Data = append([]byte(nil), v.Data...); return v }
func (s *Scoreboards) Snapshot() ScoreboardSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := ScoreboardSnapshot{Objectives: map[string]Objective{}, Scores: map[string]map[string]Score{}, DisplaySlots: map[int32]string{}}
	for k, v := range s.objectives {
		out.Objectives[k] = cloneObjective(v)
	}
	for o, entries := range s.scores {
		m := map[string]Score{}
		for e, v := range entries {
			m[e] = cloneScore(v)
		}
		out.Scores[o] = m
	}
	for k, v := range s.displays {
		out.DisplaySlots[k] = v
	}
	return out
}
func (s *Scoreboards) OnChange(fn func(ScoreboardChange)) func() { return s.onChange.subscribe(fn) }
func (s *Scoreboards) setObjective(v Objective) {
	v = cloneObjective(v)
	s.mu.Lock()
	_, ok := s.objectives[v.Name]
	s.objectives[v.Name] = v
	s.mu.Unlock()
	kind := "add"
	if ok {
		kind = "update"
	}
	c := cloneObjective(v)
	s.onChange.emit(ScoreboardChange{Kind: kind, Objective: &c})
}
func (s *Scoreboards) removeObjective(name string) {
	s.mu.Lock()
	v, ok := s.objectives[name]
	delete(s.objectives, name)
	delete(s.scores, name)
	for slot, n := range s.displays {
		if n == name {
			delete(s.displays, slot)
		}
	}
	s.mu.Unlock()
	if ok {
		c := cloneObjective(v)
		s.onChange.emit(ScoreboardChange{Kind: "remove", Objective: &c})
	}
}
func (s *Scoreboards) setDisplay(slot int32, name string) {
	s.mu.Lock()
	if name == "" {
		delete(s.displays, slot)
	} else {
		s.displays[slot] = name
	}
	s.mu.Unlock()
	s.onChange.emit(ScoreboardChange{Kind: "display", DisplaySlot: slot, DisplayObjective: name})
}
func (s *Scoreboards) setScore(v Score) {
	v = cloneScore(v)
	s.mu.Lock()
	if s.scores[v.Objective] == nil {
		s.scores[v.Objective] = map[string]Score{}
	}
	s.scores[v.Objective][v.Entity] = v
	s.mu.Unlock()
	c := cloneScore(v)
	s.onChange.emit(ScoreboardChange{Kind: "score", Score: &c})
}
func (s *Scoreboards) resetScore(entity, objective string) {
	s.mu.Lock()
	removed := []Score{}
	if objective != "" {
		if v, ok := s.scores[objective][entity]; ok {
			removed = append(removed, v)
			delete(s.scores[objective], entity)
		}
	} else {
		for _, m := range s.scores {
			if v, ok := m[entity]; ok {
				removed = append(removed, v)
				delete(m, entity)
			}
		}
	}
	s.mu.Unlock()
	for _, v := range removed {
		c := cloneScore(v)
		s.onChange.emit(ScoreboardChange{Kind: "reset", Score: &c})
	}
}

type Team struct {
	Name, DisplayName, NameTagVisibility, CollisionRule, Prefix, Suffix string
	FriendlyFlags                                                       uint8
	Color                                                               int32
	Entities                                                            []string
}
type TeamChange struct {
	Kind string
	Team Team
}
type Teams struct {
	mu       sync.RWMutex
	values   map[string]Team
	onChange event[TeamChange]
}

func newTeams() *Teams      { return &Teams{values: map[string]Team{}} }
func cloneTeam(v Team) Team { v.Entities = append([]string(nil), v.Entities...); return v }
func (t *Teams) All() []Team {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]Team, 0, len(t.values))
	for _, v := range t.values {
		out = append(out, cloneTeam(v))
	}
	return out
}
func (t *Teams) Get(name string) (Team, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	v, ok := t.values[name]
	return cloneTeam(v), ok
}
func (t *Teams) OnChange(fn func(TeamChange)) func() { return t.onChange.subscribe(fn) }
func (t *Teams) put(kind string, v Team) {
	v = cloneTeam(v)
	t.mu.Lock()
	t.values[v.Name] = v
	t.mu.Unlock()
	t.onChange.emit(TeamChange{kind, cloneTeam(v)})
}
func (t *Teams) remove(name string) {
	t.mu.Lock()
	v, ok := t.values[name]
	delete(t.values, name)
	t.mu.Unlock()
	if ok {
		t.onChange.emit(TeamChange{"remove", cloneTeam(v)})
	}
}

type BossBar struct {
	UUID, Title     string
	Health          float32
	Color, Division int32
	Flags           uint8
}
type BossBarChange struct {
	Kind string
	Bar  BossBar
}
type BossBars struct {
	mu       sync.RWMutex
	values   map[string]BossBar
	onChange event[BossBarChange]
}

func newBossBars() *BossBars { return &BossBars{values: map[string]BossBar{}} }
func (b *BossBars) All() []BossBar {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]BossBar, 0, len(b.values))
	for _, v := range b.values {
		out = append(out, v)
	}
	return out
}
func (b *BossBars) Get(id string) (BossBar, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	v, ok := b.values[id]
	return v, ok
}
func (b *BossBars) OnChange(fn func(BossBarChange)) func() { return b.onChange.subscribe(fn) }
func (b *BossBars) update(id, kind string, fn func(*BossBar)) {
	b.mu.Lock()
	v := b.values[id]
	v.UUID = id
	fn(&v)
	if kind == "remove" {
		delete(b.values, id)
	} else {
		b.values[id] = v
	}
	b.mu.Unlock()
	b.onChange.emit(BossBarChange{kind, v})
}

type SoundEvent struct {
	ID, Category, EntityID int32
	Name                   string
	Position               Vec3
	Volume, Pitch          float32
	Seed                   int64
	FixedRange             *float32
}
type ParticleEvent struct {
	ID, Count                   int32
	Position                    Vec3
	Offset                      Vec3
	MaxSpeed                    float32
	LongDistance, AlwaysVisible bool
	Data                        []byte
}
type ExplosionEvent struct {
	Position Vec3
	Data     []byte
}
type WorldEffects struct {
	onSound     event[SoundEvent]
	onParticle  event[ParticleEvent]
	onExplosion event[ExplosionEvent]
}

func newWorldEffects() *WorldEffects                             { return &WorldEffects{} }
func (e *WorldEffects) OnSound(fn func(SoundEvent)) func()       { return e.onSound.subscribe(fn) }
func (e *WorldEffects) OnParticle(fn func(ParticleEvent)) func() { return e.onParticle.subscribe(fn) }
func (e *WorldEffects) OnExplosion(fn func(ExplosionEvent)) func() {
	return e.onExplosion.subscribe(fn)
}

type ServerState struct {
	MOTD                                                                                                    string
	Icon                                                                                                    []byte
	Difficulty                                                                                              uint8
	DifficultyLocked                                                                                        bool
	ViewDistance, SimulationDistance, MaxPlayers                                                            int32
	Hardcore, ReducedDebugInfo, RespawnScreen, LimitedCrafting, Debug, Flat, OnlineMode, EnforcesSecureChat bool
	GameRules                                                                                               map[string]string
}
type Server struct {
	mu       sync.RWMutex
	state    ServerState
	onChange event[ServerState]
}

func newServer() *Server { return &Server{state: ServerState{GameRules: map[string]string{}}} }
func (s *Server) State() ServerState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v := s.state
	v.Icon = append([]byte(nil), v.Icon...)
	v.GameRules = make(map[string]string, len(s.state.GameRules))
	for k, x := range s.state.GameRules {
		v.GameRules[k] = x
	}
	return v
}
func (s *Server) OnChange(fn func(ServerState)) func() { return s.onChange.subscribe(fn) }
func (s *Server) update(fn func(*ServerState)) {
	s.mu.Lock()
	fn(&s.state)
	v := s.state
	v.Icon = append([]byte(nil), v.Icon...)
	v.GameRules = make(map[string]string, len(s.state.GameRules))
	for k, x := range s.state.GameRules {
		v.GameRules[k] = x
	}
	s.mu.Unlock()
	s.onChange.emit(v)
}
