package minego

import (
	"sync"

	"github.com/zeozeozeo/minego/internal/data/chunks"
	"github.com/zeozeozeo/minego/internal/protocol/nbt"
)

type chunkKey struct{ X, Z int32 }
type BlockEntity struct {
	Type int32
	Data nbt.Compound
}

// World is a concurrency-safe view of loaded server chunks.
type World struct {
	bot           *Bot
	mu            sync.RWMutex
	dimension     string
	chunks        map[chunkKey]*chunks.ChunkColumn
	blockEntities map[BlockPos]BlockEntity
	center        chunkKey
	viewDistance  int32
	onBlock       event[BlockChange]
	onChunk       event[chunkKey]
}

func newWorld(b *Bot) *World {
	return &World{bot: b, chunks: map[chunkKey]*chunks.ChunkColumn{}, blockEntities: map[BlockPos]BlockEntity{}}
}
func (w *World) Dimension() string { w.mu.RLock(); defer w.mu.RUnlock(); return w.dimension }
func (w *World) Block(pos BlockPos) (Block, bool) {
	w.mu.RLock()
	col := w.chunks[chunkKey{int32(pos.X >> 4), int32(pos.Z >> 4)}]
	var id int32
	if col != nil {
		id = col.GetBlockState(pos.X, pos.Y, pos.Z)
	}
	w.mu.RUnlock()
	if col == nil {
		return Block{}, false
	}
	return w.bot.block(pos, id)
}
func (w *World) IsLoaded(pos BlockPos) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.chunks[chunkKey{int32(pos.X >> 4), int32(pos.Z >> 4)}] != nil
}
func (w *World) LoadedChunks() [][2]int32 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	r := make([][2]int32, 0, len(w.chunks))
	for k := range w.chunks {
		r = append(r, [2]int32{k.X, k.Z})
	}
	return r
}
func (w *World) OnBlockChange(fn func(BlockChange)) func() { return w.onBlock.subscribe(fn) }
func (w *World) OnChunkLoad(fn func(x, z int32)) func() {
	return w.onChunk.subscribe(func(k chunkKey) { fn(k.X, k.Z) })
}
func (w *World) reset(dimension string) {
	w.mu.Lock()
	w.dimension = dimension
	w.chunks = map[chunkKey]*chunks.ChunkColumn{}
	w.blockEntities = map[BlockPos]BlockEntity{}
	w.mu.Unlock()
}
