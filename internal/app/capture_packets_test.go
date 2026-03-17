package app

import (
	"context"
	"errors"
	"testing"
	"time"

	domcapture "github.com/kwon93/goscope/internal/domain/capture"
	"github.com/kwon93/goscope/internal/domain/packet"
)

// fakeSource는 미리 준비한 패킷을 채널로 보내는 fake PacketSource다.
type fakeSource struct {
	packets []packet.Packet
}

func (f *fakeSource) Capture(_ context.Context, _ domcapture.CaptureRequest) (<-chan packet.Packet, error) {
	ch := make(chan packet.Packet, len(f.packets))
	for _, p := range f.packets {
		ch <- p
	}
	close(ch)
	return ch, nil
}

// blockingSource는 ctx가 취소될 때까지 채널을 열어두는 fake PacketSource다.
type blockingSource struct{}

func (b *blockingSource) Capture(ctx context.Context, _ domcapture.CaptureRequest) (<-chan packet.Packet, error) {
	ch := make(chan packet.Packet)
	go func() {
		defer close(ch)
		<-ctx.Done()
	}()
	return ch, nil
}

// capturingSource는 전달받은 CaptureRequest를 기록하는 fake PacketSource다.
type capturingSource struct {
	lastReq domcapture.CaptureRequest
}

func (c *capturingSource) Capture(_ context.Context, req domcapture.CaptureRequest) (<-chan packet.Packet, error) {
	c.lastReq = req
	ch := make(chan packet.Packet)
	close(ch)
	return ch, nil
}

// fakeSink는 수신한 패킷을 기록하는 fake PacketSink다.
type fakeSink struct {
	received []packet.Packet
}

func (f *fakeSink) WritePacket(_ context.Context, pkt packet.Packet) error {
	f.received = append(f.received, pkt)
	return nil
}

// errorSource는 항상 에러를 반환하는 fake PacketSource다.
type errorSource struct{}

func (e *errorSource) Capture(_ context.Context, _ domcapture.CaptureRequest) (<-chan packet.Packet, error) {
	return nil, errors.New("source error")
}

// errorSink는 항상 에러를 반환하는 fake PacketSink다.
type errorSink struct{}

func (e *errorSink) WritePacket(_ context.Context, _ packet.Packet) error {
	return errors.New("sink error")
}

func TestCapturePackets_Run_DeliversToBothSinks(t *testing.T) {
	pkt := packet.Packet{
		Timestamp: time.Now(),
		Protocol:  packet.TCP,
		SrcAddr:   "1.2.3.4",
		DstAddr:   "5.6.7.8",
	}

	src := &fakeSource{packets: []packet.Packet{pkt}}
	sink1 := &fakeSink{}
	sink2 := &fakeSink{}

	uc := CapturePackets{
		Source: src,
		Sinks:  []domcapture.PacketSink{sink1, sink2},
	}

	if err := uc.Run(context.Background(), CapturePacketsInput{
		Interface: "eth0",
		Snaplen:   1600,
		Promisc:   true,
	}); err != nil {
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

func TestCapturePackets_Run_MultiplePackets(t *testing.T) {
	pkts := []packet.Packet{
		{Timestamp: time.Now(), Protocol: packet.TCP},
		{Timestamp: time.Now(), Protocol: packet.UDP},
		{Timestamp: time.Now(), Protocol: packet.OTHER},
	}

	src := &fakeSource{packets: pkts}
	sink := &fakeSink{}

	uc := CapturePackets{Source: src, Sinks: []domcapture.PacketSink{sink}}

	if err := uc.Run(context.Background(), CapturePacketsInput{}); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(sink.received) != 3 {
		t.Fatalf("sink received %d packets; want 3", len(sink.received))
	}
}

func TestCapturePackets_Run_SourceError(t *testing.T) {
	uc := CapturePackets{Source: &errorSource{}, Sinks: nil}

	err := uc.Run(context.Background(), CapturePacketsInput{Interface: "eth0"})
	if err == nil {
		t.Fatal("Run: got nil error; want error from source")
	}
}

func TestCapturePackets_Run_SinkError(t *testing.T) {
	pkt := packet.Packet{Timestamp: time.Now()}
	src := &fakeSource{packets: []packet.Packet{pkt}}

	uc := CapturePackets{
		Source: src,
		Sinks:  []domcapture.PacketSink{&errorSink{}},
	}

	err := uc.Run(context.Background(), CapturePacketsInput{Interface: "eth0"})
	if err == nil {
		t.Fatal("Run: got nil error; want error from sink")
	}
}

func TestCapturePackets_Run_EmptySource(t *testing.T) {
	// 패킷이 없으면 에러 없이 즉시 반환해야 한다.
	uc := CapturePackets{
		Source: &fakeSource{packets: nil},
		Sinks:  []domcapture.PacketSink{&fakeSink{}},
	}

	if err := uc.Run(context.Background(), CapturePacketsInput{}); err != nil {
		t.Fatalf("Run with empty source: got error %v; want nil", err)
	}
}

func TestCapturePackets_Run_NoSinks(t *testing.T) {
	// Sinks가 없어도 패닉 없이 정상 완료해야 한다.
	pkt := packet.Packet{Timestamp: time.Now(), Protocol: packet.TCP}
	uc := CapturePackets{
		Source: &fakeSource{packets: []packet.Packet{pkt}},
		Sinks:  nil,
	}

	if err := uc.Run(context.Background(), CapturePacketsInput{}); err != nil {
		t.Fatalf("Run with no sinks: got error %v; want nil", err)
	}
}

func TestCapturePackets_Run_PacketOrder(t *testing.T) {
	// 패킷이 소스에서 받은 순서대로 싱크에 전달되어야 한다.
	pkts := []packet.Packet{
		{SrcAddr: "1.1.1.1"},
		{SrcAddr: "2.2.2.2"},
		{SrcAddr: "3.3.3.3"},
	}

	sink := &fakeSink{}
	uc := CapturePackets{
		Source: &fakeSource{packets: pkts},
		Sinks:  []domcapture.PacketSink{sink},
	}

	if err := uc.Run(context.Background(), CapturePacketsInput{}); err != nil {
		t.Fatalf("Run: %v", err)
	}

	for i, want := range []string{"1.1.1.1", "2.2.2.2", "3.3.3.3"} {
		if got := sink.received[i].SrcAddr; got != want {
			t.Fatalf("packet[%d].SrcAddr = %v; want %v", i, got, want)
		}
	}
}

func TestCapturePackets_Run_ContextCancellation(t *testing.T) {
	// ctx가 취소되면 Run이 에러 없이 반환해야 한다.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	uc := CapturePackets{
		Source: &blockingSource{},
		Sinks:  []domcapture.PacketSink{&fakeSink{}},
	}

	if err := uc.Run(ctx, CapturePacketsInput{}); err != nil {
		t.Fatalf("Run after context cancellation: got error %v; want nil", err)
	}
}

func TestCapturePackets_Run_CaptureRequestPropagation(t *testing.T) {
	// Run에 전달한 Input이 CaptureRequest로 올바르게 변환되어 Source에 전달되어야 한다.
	src := &capturingSource{}
	uc := CapturePackets{Source: src, Sinks: nil}

	in := CapturePacketsInput{
		Interface: "eth0",
		Filter:    "tcp port 80",
		Snaplen:   512,
		Promisc:   false,
		ChBuffer:  10,
	}

	uc.Run(context.Background(), in) //nolint

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
				t.Fatalf("CaptureRequest.%s = %v; want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}
