package minego

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/zeozeozeo/minego/internal/data/versions/v26_2/packets"
	jp "github.com/zeozeozeo/minego/internal/protocol/java_protocol"
)

func TestVersionSelectionAndCapabilities(t *testing.T) {
	versions := SupportedVersions()
	if len(versions) != 3 || versions[0].Name != "1.21.11" || versions[0].Protocol != 774 || versions[1].Name != "26.1" || versions[1].Protocol != 775 || versions[2].Name != "26.2" || versions[2].Protocol != 776 {
		t.Fatalf("supported versions = %#v", versions)
	}
	versions[0].Name = "mutated"
	if SupportedVersions()[0].Name != "1.21.11" {
		t.Fatal("SupportedVersions returned shared state")
	}
	bot, err := New(Config{Address: "localhost", Version: "26.2"})
	if err != nil {
		t.Fatal(err)
	}
	if !bot.Supports(FeatureNavigation) || bot.Version().Protocol != 776 {
		t.Fatalf("selected version = %#v", bot.Version())
	}
	if err := bot.Require(FeatureNavigation); err != nil {
		t.Fatal(err)
	}
	bot261, err := New(Config{Address: "localhost", Version: "26.1"})
	if err != nil || bot261.Version().Protocol != 775 {
		t.Fatalf("26.1 selection = %#v, %v", bot261, err)
	}
	if bot, err := New(Config{Address: "localhost", Version: "1.21.11"}); err != nil || bot.Version().Protocol != 774 {
		t.Fatalf("1.21.11 selection = %#v, %v", bot, err)
	}
	if err := bot.Require(Feature("future")); !errors.Is(err, ErrUnsupportedFeature) {
		t.Fatalf("unsupported feature error = %v", err)
	}
	_, err = New(Config{Address: "localhost", Version: "1.20"})
	var unsupported *UnsupportedVersionError
	if !errors.As(err, &unsupported) || !errors.Is(err, ErrUnsupportedVersion) {
		t.Fatalf("unsupported error = %v", err)
	}
}

type testPlugin struct{ registered bool }

func (p *testPlugin) Register(ctx *PluginContext) error {
	p.registered = true
	return ctx.AddService("test/value", 42)
}

func TestPluginRegistersBeforeConnection(t *testing.T) {
	plugin := &testPlugin{}
	bot, err := New(Config{Address: "localhost", Version: "26.2", Plugins: []Plugin{plugin}})
	if err != nil {
		t.Fatal(err)
	}
	value, ok := bot.Service("test/value")
	if !plugin.registered || !ok || value != 42 {
		t.Fatalf("registered=%v service=%v,%v", plugin.registered, value, ok)
	}
}

func TestSubscriptionUnsubscribeIsIdempotent(t *testing.T) {
	var messages event[int]
	called := 0
	unsubscribe := messages.subscribe(func(value int) { called += value })
	messages.emit(2)
	unsubscribe()
	unsubscribe()
	messages.emit(2)
	if called != 2 {
		t.Fatalf("handler called total = %d", called)
	}
}

func TestActionCoordinatorPreemptionAndCancellation(t *testing.T) {
	c := newActionCoordinator()
	background, err := c.acquire(context.Background(), controlMovement|controlView, priorityBackground)
	if err != nil {
		t.Fatal(err)
	}
	backgroundCtx := background.Context(context.Background())
	explicit, err := c.acquire(context.Background(), controlMovement, priorityExplicit)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-backgroundCtx.Done():
	case <-time.After(time.Second):
		t.Fatal("background action was not preempted")
	}
	explicit.Release()
	explicit.Release()

	held, err := c.acquire(context.Background(), controlHands, priorityExplicit)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.acquire(ctx, controlHands, priorityAutomation); !errors.Is(err, context.Canceled) {
		t.Fatalf("waiting acquire error = %v", err)
	}
	held.Release()
}

func TestObservabilitySnapshot(t *testing.T) {
	bot, err := New(Config{Address: "localhost", Version: "26.2"})
	if err != nil {
		t.Fatal(err)
	}
	bot.Observability.received(12)
	bot.Observability.sent(7)
	snapshot := bot.Observability.Snapshot()
	if snapshot.PacketsReceived != 1 || snapshot.BytesReceived != 12 || snapshot.PacketsSent != 1 || snapshot.BytesSent != 7 || snapshot.LastPacketAt.IsZero() {
		t.Fatalf("telemetry = %#v", snapshot)
	}
}

func TestProbeProtocol(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	serverErr := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			serverErr <- err
			return
		}
		defer conn.Close()
		protocolConn := jp.NewConn(conn)
		if _, err = jp.ReadWirePacketFrom(protocolConn, -1); err != nil {
			serverErr <- err
			return
		}
		if _, err = jp.ReadWirePacketFrom(protocolConn, -1); err != nil {
			serverErr <- err
			return
		}
		wire, err := jp.ToWire(&packets.S2CStatusResponse{JsonResponse: `{"version":{"name":"26.2","protocol":776},"players":{"max":20,"online":0},"description":"test"}`})
		if err == nil {
			err = wire.WriteTo(protocolConn, -1)
		}
		serverErr <- err
	}()
	host, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	protocol, err := probeProtocol(context.Background(), host, port, listener.Addr().String(), time.Second)
	if err != nil || protocol != 776 {
		t.Fatalf("probe = %d, %v", protocol, err)
	}
	if err := <-serverErr; err != nil {
		t.Fatal(err)
	}
}
