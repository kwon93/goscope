package capture

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakeSource는 미리 준비한 패킷을 채널로 보내는 fake Source다.
type fakeSource struct {
	packets []Packet
}

func (f *fakeSource) Capture(_ context.Context, _ Request) (<-chan Packet, error) {
	ch := make(chan Packet, len(f.packets))
	for _, p := range f.packets {
		ch <- p
	}
	close(ch)
	return ch, nil
}

// blockingSource는 ctx가 취소될 때까지 채널을 열어두는 fake Source다.
type blockingSource struct{}

func (b *blockingSource) Capture(ctx context.Context, _ Request) (<-chan Packet, error) {
	ch := make(chan Packet)
	go func() {
		defer close(ch)
		<-ctx.Done()
	}()
	return ch, nil
}

// capturingSource는 전달받은 Request를 기록하는 fake Source다.
type capturingSource struct {
	lastReq Request
}

func (c *capturingSource) Capture(_ context.Context, req Request) (<-chan Packet, error) {
	c.lastReq = req
	ch := make(chan Packet)
	close(ch)
	return ch, nil
}

// fakeSink는 수신한 패킷을 기록하는 fake Sink다.
type fakeSink struct {
	received []Packet
}

func (f *fakeSink) WritePacket(_ context.Context, pkt Packet) error {
	f.received = append(f.received, pkt)
	return nil
}

// errorSource는 항상 에러를 반환하는 fake Source다.
type errorSource struct{}

func (e *errorSource) Capture(_ context.Context, _ Request) (<-chan Packet, error) {
	return nil, errors.New("source error")
}

// errorSink는 항상 에러를 반환하는 fake Sink다.
type errorSink struct{}

func (e *errorSink) WritePacket(_ context.Context, _ Packet) error {
	return errors.New("sink error")
}

func TestRun_DeliversToBothSinks(t *testing.T) {
	pkt := Packet{
		Timestamp: time.Now(),
		Protocol:  TCP,
		SrcAddr:   "1.2.3.4",
		DstAddr:   "5.6.7.8",
	}

	src := &fakeSource{packets: []Packet{pkt}}
	sink1 := &fakeSink{}
	sink2 := &fakeSink{}

	if err := Run(context.Background(), src, []Sink{sink1, sink2}, Request{Interface: "eth0", Snaplen: 1600, Promisc: true}); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(sink1.received) != 1 {
		t.Fatalf("sink1 received %d packets; want 1", len(sink1.received))
	}
	if len(sink2.received) != 1 {
		t.Fatalf("sink2 received %d packets; want 1", len(sink2.received))
	}
	if sink1.received[0].SrcAddr != "1.2.3.4" {
		t.Fatalf("sink1 SrcAddr = %v; want 1.2.3.4", sink1.received[0].SrcAddr)
	}
}

func TestRun_MultiplePackets(t *testing.T) {
	pkts := []Packet{
		{Timestamp: time.Now(), Protocol: TCP},
		{Timestamp: time.Now(), Protocol: UDP},
		{Timestamp: time.Now(), Protocol: OTHER},
	}

	sink := &fakeSink{}
	if err := Run(context.Background(), &fakeSource{packets: pkts}, []Sink{sink}, Request{}); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(sink.received) != 3 {
		t.Fatalf("sink received %d packets; want 3", len(sink.received))
	}
}

func TestRun_SourceError(t *testing.T) {
	err := Run(context.Background(), &errorSource{}, nil, Request{Interface: "eth0"})
	if err == nil {
		t.Fatal("Run: got nil error; want error from source")
	}
}

func TestRun_SinkError(t *testing.T) {
	pkt := Packet{Timestamp: time.Now()}
	src := &fakeSource{packets: []Packet{pkt}}

	err := Run(context.Background(), src, []Sink{&errorSink{}}, Request{Interface: "eth0"})
	if err == nil {
		t.Fatal("Run: got nil error; want error from sink")
	}
}

func TestRun_EmptySource(t *testing.T) {
	if err := Run(context.Background(), &fakeSource{}, []Sink{&fakeSink{}}, Request{}); err != nil {
		t.Fatalf("Run with empty source: got error %v; want nil", err)
	}
}

func TestRun_NoSinks(t *testing.T) {
	pkt := Packet{Timestamp: time.Now(), Protocol: TCP}
	if err := Run(context.Background(), &fakeSource{packets: []Packet{pkt}}, nil, Request{}); err != nil {
		t.Fatalf("Run with no sinks: got error %v; want nil", err)
	}
}

func TestRun_PacketOrder(t *testing.T) {
	pkts := []Packet{
		{SrcAddr: "1.1.1.1"},
		{SrcAddr: "2.2.2.2"},
		{SrcAddr: "3.3.3.3"},
	}

	sink := &fakeSink{}
	if err := Run(context.Background(), &fakeSource{packets: pkts}, []Sink{sink}, Request{}); err != nil {
		t.Fatalf("Run: %v", err)
	}

	for i, want := range []string{"1.1.1.1", "2.2.2.2", "3.3.3.3"} {
		if got := sink.received[i].SrcAddr; got != want {
			t.Fatalf("packet[%d].SrcAddr = %v; want %v", i, got, want)
		}
	}
}

func TestRun_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	if err := Run(ctx, &blockingSource{}, []Sink{&fakeSink{}}, Request{}); err != nil {
		t.Fatalf("Run after context cancellation: got error %v; want nil", err)
	}
}

func TestRun_RequestPropagation(t *testing.T) {
	src := &capturingSource{}

	req := Request{
		Interface: "eth0",
		Filter:    "tcp port 80",
		Snaplen:   512,
		Promisc:   false,
		ChBuffer:  10,
	}

	Run(context.Background(), src, nil, req) //nolint

	tests := []struct {
		name string
		got  any
		want any
	}{
		{"Interface", src.lastReq.Interface, "eth0"},
		{"Filter", src.lastReq.Filter, "tcp port 80"},
		{"Snaplen", src.lastReq.Snaplen, int32(512)},
		{"Promisc", src.lastReq.Promisc, false},
		{"ChBuffer", src.lastReq.ChBuffer, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("Request.%s = %v; want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestProtocolConstants(t *testing.T) {
	tests := []struct {
		name string
		got  Protocol
		want Protocol
	}{
		{"TCP", TCP, "tcp"},
		{"UDP", UDP, "udp"},
		{"OTHER", OTHER, "other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("protocol = %v; want %v", tt.got, tt.want)
			}
		})
	}
}

func TestPacket_OtherProtocol_NilPorts(t *testing.T) {
	pkt := Packet{Protocol: OTHER}
	if pkt.SrcPort != nil {
		t.Fatal("OTHER packet SrcPort: got non-nil; want nil")
	}
	if pkt.DstPort != nil {
		t.Fatal("OTHER packet DstPort: got non-nil; want nil")
	}
}
