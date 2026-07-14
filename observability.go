package minego

import (
	"sync/atomic"
	"time"
)

// Telemetry is an immutable point-in-time view of connection and action
// activity. Counters are process-local and reset for each Bot.
type Telemetry struct {
	PacketsReceived uint64
	PacketsSent     uint64
	BytesReceived   uint64
	BytesSent       uint64
	ActiveActions   int
	LastPacketAt    time.Time
}

type Observability struct {
	bot             *Bot
	packetsReceived atomic.Uint64
	packetsSent     atomic.Uint64
	bytesReceived   atomic.Uint64
	bytesSent       atomic.Uint64
	lastPacketUnix  atomic.Int64
}

func newObservability(bot *Bot) *Observability { return &Observability{bot: bot} }

func (o *Observability) Snapshot() Telemetry {
	last := o.lastPacketUnix.Load()
	var lastPacket time.Time
	if last != 0 {
		lastPacket = time.Unix(0, last)
	}
	return Telemetry{
		PacketsReceived: o.packetsReceived.Load(),
		PacketsSent:     o.packetsSent.Load(),
		BytesReceived:   o.bytesReceived.Load(),
		BytesSent:       o.bytesSent.Load(),
		ActiveActions:   o.bot.actions.count(),
		LastPacketAt:    lastPacket,
	}
}

func (o *Observability) received(bytes int) {
	o.packetsReceived.Add(1)
	o.bytesReceived.Add(uint64(bytes))
	o.lastPacketUnix.Store(time.Now().UnixNano())
}

func (o *Observability) sent(bytes int) {
	o.packetsSent.Add(1)
	o.bytesSent.Add(uint64(bytes))
}
