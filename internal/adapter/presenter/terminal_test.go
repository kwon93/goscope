package presenter

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/kwon93/goscope/internal/domain/packet"
)

func TestTerminal_WritePacket_TCP(t *testing.T) {
	var buf bytes.Buffer
	term := NewTerminal(&buf)

	srcPort := uint16(54321)
	dstPort := uint16(80)

	pkt := packet.Packet{
		Timestamp: time.Date(2024, 1, 1, 14, 32, 1, 0, time.UTC),
		Protocol:  packet.TCP,
		SrcAddr:   "192.168.0.1",
		DstAddr:   "8.8.8.8",
		SrcPort:   &srcPort,
		DstPort:   &dstPort,
	}

	if err := term.WritePacket(context.Background(), pkt); err != nil {
		t.Fatalf("WritePacket: %v", err)
	}

	out := buf.String()
	tests := []struct {
		name    string
		contain string
	}{
		{"protocol", "tcp"},
		{"src IP", "192.168.0.1"},
		{"dst IP", "8.8.8.8"},
		{"src port", "54321"},
		{"dst port", "80"},
		{"timestamp", "14:32:01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(out, tt.contain) {
				t.Fatalf("output = %q; want to contain %q", out, tt.contain)
			}
		})
	}
}

func TestTerminal_WritePacket_UDP(t *testing.T) {
	var buf bytes.Buffer
	term := NewTerminal(&buf)

	srcPort := uint16(12345)
	dstPort := uint16(53)

	pkt := packet.Packet{
		Timestamp: time.Now(),
		Protocol:  packet.UDP,
		SrcAddr:   "192.168.0.1",
		DstAddr:   "8.8.8.8",
		SrcPort:   &srcPort,
		DstPort:   &dstPort,
	}

	if err := term.WritePacket(context.Background(), pkt); err != nil {
		t.Fatalf("WritePacket: %v", err)
	}

	out := buf.String()
	tests := []struct {
		name    string
		contain string
	}{
		{"protocol", "udp"},
		{"dst port", "53"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(out, tt.contain) {
				t.Fatalf("output = %q; want to contain %q", out, tt.contain)
			}
		})
	}
}

func TestTerminal_WritePacket_OutputFormat(t *testing.T) {
	// 출력 포맷: [HH:MM:SS] proto  src:srcPort -> dstPort → dst\n
	var buf bytes.Buffer
	term := NewTerminal(&buf)

	srcPort := uint16(8080)
	dstPort := uint16(443)

	pkt := packet.Packet{
		Timestamp: time.Date(2024, 6, 1, 9, 5, 3, 0, time.UTC),
		Protocol:  packet.TCP,
		SrcAddr:   "10.0.0.1",
		DstAddr:   "10.0.0.2",
		SrcPort:   &srcPort,
		DstPort:   &dstPort,
	}
	term.WritePacket(context.Background(), pkt) //nolint

	out := buf.String()
	// 타임스탬프 포맷 검증: 앞에 0이 붙어야 한다 (09:05:03)
	if !strings.Contains(out, "09:05:03") {
		t.Fatalf("output = %q; want timestamp 09:05:03", out)
	}
	// 한 줄로 끝나야 한다
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("output has %d lines; want 1", len(lines))
	}
}

func TestTerminal_NewTerminal_NilWriter(t *testing.T) {
	// NewTerminal(nil)은 패닉 없이 생성되어야 한다.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("NewTerminal(nil) panicked: %v", r)
		}
	}()
	NewTerminal(nil)
}

func TestTerminal_WritePacket_OTHER_NoPorts(t *testing.T) {
	var buf bytes.Buffer
	term := NewTerminal(&buf)

	pkt := packet.Packet{
		Timestamp: time.Now(),
		Protocol:  packet.OTHER,
		SrcAddr:   "10.0.0.1",
		DstAddr:   "10.0.0.255",
	}

	if err := term.WritePacket(context.Background(), pkt); err != nil {
		t.Fatalf("WritePacket: %v", err)
	}

	out := buf.String()
	tests := []struct {
		name    string
		contain string
	}{
		{"src addr", "10.0.0.1"},
		{"dst addr", "10.0.0.255"},
		{"protocol", "other"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(out, tt.contain) {
				t.Fatalf("output = %q; want to contain %q", out, tt.contain)
			}
		})
	}
	// 포트 번호가 없어야 함 (포트 형식 "숫자 -> 숫자" 미포함)
	if strings.Contains(out, " -> ") {
		t.Fatalf("output = %q; OTHER packet should not contain port info", out)
	}
}
