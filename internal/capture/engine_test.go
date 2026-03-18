package capture

import (
	"context"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func TestEngine_Capture_InvalidInterface(t *testing.T) {
	e := NewEngine()
	_, err := e.Capture(context.Background(), Request{
		Interface: "nonexistent_iface_xyz",
		Snaplen:   1600,
		Promisc:   true,
	})
	if err == nil {
		t.Fatal("Capture() with invalid interface: got nil error; want error")
	}
}

func TestEngine_Capture_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := NewEngine().Capture(ctx, Request{
		Interface: "nonexistent_iface_xyz",
		Snaplen:   1600,
		Promisc:   true,
	})
	if err == nil {
		t.Fatal("Capture() with cancelled context and invalid interface: got nil error; want error")
	}
}

// newRawIPv4TCP는 테스트용 IPv4+TCP raw bytes를 반환한다.
// src: 192.168.0.1:54321 → dst: 8.8.8.8:80
func newRawIPv4TCP() []byte {
	return []byte{
		// Ethernet
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x08, 0x00,
		// IPv4: total=40, proto=TCP(6)
		0x45, 0x00, 0x00, 0x28,
		0x00, 0x01, 0x00, 0x00,
		0x40, 0x06, 0x00, 0x00,
		0xc0, 0xa8, 0x00, 0x01, // src: 192.168.0.1
		0x08, 0x08, 0x08, 0x08, // dst: 8.8.8.8
		// TCP: src=54321, dst=80
		0xd4, 0x31, 0x00, 0x50,
		0x00, 0x00, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x00,
		0x50, 0x02, 0x72, 0x10,
		0x00, 0x00, 0x00, 0x00,
	}
}

// newRawIPv4UDP는 테스트용 IPv4+UDP raw bytes를 반환한다.
// src: 192.168.0.1:12345 → dst: 8.8.8.8:53
func newRawIPv4UDP() []byte {
	return []byte{
		// Ethernet
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x08, 0x00,
		// IPv4: total=28, proto=UDP(17)
		0x45, 0x00, 0x00, 0x1c,
		0x00, 0x01, 0x00, 0x00,
		0x40, 0x11, 0x00, 0x00,
		0xc0, 0xa8, 0x00, 0x01, // src: 192.168.0.1
		0x08, 0x08, 0x08, 0x08, // dst: 8.8.8.8
		// UDP: src=12345, dst=53
		0x30, 0x39, 0x00, 0x35,
		0x00, 0x08, 0x00, 0x00,
	}
}

func TestToDomainPacket_TCP(t *testing.T) {
	raw := gopacket.NewPacket(newRawIPv4TCP(), layers.LayerTypeEthernet, gopacket.Default)
	pkt, ok := toDomainPacket(raw)

	if !ok {
		t.Fatal("toDomainPacket: got ok=false; want true")
	}

	tests := []struct {
		name string
		got  any
		want any
	}{
		{"Protocol", pkt.Protocol, TCP},
		{"SrcAddr", pkt.SrcAddr, "192.168.0.1"},
		{"DstAddr", pkt.DstAddr, "8.8.8.8"},
		{"SrcPort", *pkt.SrcPort, uint16(54321)},
		{"DstPort", *pkt.DstPort, uint16(80)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("%s = %v; want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestToDomainPacket_UDP(t *testing.T) {
	raw := gopacket.NewPacket(newRawIPv4UDP(), layers.LayerTypeEthernet, gopacket.Default)
	pkt, ok := toDomainPacket(raw)

	if !ok {
		t.Fatal("toDomainPacket: got ok=false; want true")
	}

	tests := []struct {
		name string
		got  any
		want any
	}{
		{"Protocol", pkt.Protocol, UDP},
		{"SrcAddr", pkt.SrcAddr, "192.168.0.1"},
		{"DstAddr", pkt.DstAddr, "8.8.8.8"},
		{"DstPort", *pkt.DstPort, uint16(53)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("%s = %v; want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestToDomainPacket_NoNetworkLayer(t *testing.T) {
	raw := gopacket.NewPacket([]byte{0x00, 0x01, 0x02}, layers.LayerTypeEthernet, gopacket.Default)
	_, ok := toDomainPacket(raw)

	if ok {
		t.Fatal("toDomainPacket: got ok=true for packet with no network layer; want false")
	}
}

func TestToDomainPacket_ICMP_OtherProtocol(t *testing.T) {
	raw := []byte{
		// Ethernet
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x08, 0x00,
		// IPv4: total=28, proto=ICMP(1)
		0x45, 0x00, 0x00, 0x1c,
		0x00, 0x01, 0x00, 0x00,
		0x40, 0x01, 0x00, 0x00,
		0xc0, 0xa8, 0x00, 0x01, // src: 192.168.0.1
		0x08, 0x08, 0x08, 0x08, // dst: 8.8.8.8
		// ICMP echo request
		0x08, 0x00, 0x00, 0x00,
		0x00, 0x01, 0x00, 0x01,
	}
	goPkt := gopacket.NewPacket(raw, layers.LayerTypeEthernet, gopacket.Default)
	pkt, ok := toDomainPacket(goPkt)

	if !ok {
		t.Fatal("toDomainPacket ICMP: got ok=false; want true")
	}

	tests := []struct {
		name string
		got  any
		want any
	}{
		{"Protocol", pkt.Protocol, OTHER},
		{"SrcAddr", pkt.SrcAddr, "192.168.0.1"},
		{"DstAddr", pkt.DstAddr, "8.8.8.8"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("%s = %v; want %v", tt.name, tt.got, tt.want)
			}
		})
	}

	if pkt.SrcPort != nil {
		t.Fatalf("ICMP SrcPort = %v; want nil", *pkt.SrcPort)
	}
	if pkt.DstPort != nil {
		t.Fatalf("ICMP DstPort = %v; want nil", *pkt.DstPort)
	}
}

func TestToDomainPacket_PreservesRawData(t *testing.T) {
	rawBytes := newRawIPv4TCP()
	goPkt := gopacket.NewPacket(rawBytes, layers.LayerTypeEthernet, gopacket.Default)
	pkt, ok := toDomainPacket(goPkt)

	if !ok {
		t.Fatal("toDomainPacket: got ok=false; want true")
	}
	if len(pkt.RawData) == 0 {
		t.Fatal("RawData: got empty; want non-empty")
	}
	if len(pkt.RawData) != len(rawBytes) {
		t.Fatalf("RawData len = %d; want %d", len(pkt.RawData), len(rawBytes))
	}
}

func TestToDomainPacket_CaptureMetadata(t *testing.T) {
	goPkt := gopacket.NewPacket(newRawIPv4TCP(), layers.LayerTypeEthernet, gopacket.Default)
	pkt, ok := toDomainPacket(goPkt)

	if !ok {
		t.Fatal("toDomainPacket: got ok=false; want true")
	}
	if pkt.CaptureLen < 0 {
		t.Fatalf("CaptureLen = %d; want >= 0", pkt.CaptureLen)
	}
	if pkt.OriginalLen < 0 {
		t.Fatalf("OriginalLen = %d; want >= 0", pkt.OriginalLen)
	}
}
