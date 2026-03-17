package packet

import (
	"testing"
	"time"
)

func TestPacket_Fields(t *testing.T) {
	srcPort := uint16(54321)
	dstPort := uint16(80)
	now := time.Now()

	pkt := Packet{
		Timestamp:   now,
		Protocol:    TCP,
		SrcAddr:     "192.168.0.1",
		DstAddr:     "8.8.8.8",
		SrcPort:     &srcPort,
		DstPort:     &dstPort,
		RawData:     []byte{0x01, 0x02},
		CaptureLen:  2,
		OriginalLen: 60,
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
		{"CaptureLen", pkt.CaptureLen, 2},
		{"OriginalLen", pkt.OriginalLen, 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("%s = %v; want %v", tt.name, tt.got, tt.want)
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
