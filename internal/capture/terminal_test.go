package capture

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestTerminal_WritePacket_TCP(t *testing.T) {
	var buf bytes.Buffer
	term := NewTerminal(&buf)

	srcPort := uint16(54321)
	dstPort := uint16(80)

	pkt := Packet{
		Timestamp: time.Date(2024, 1, 1, 14, 32, 1, 0, time.UTC),
		Protocol:  TCP,
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

	pkt := Packet{
		Timestamp: time.Now(),
		Protocol:  UDP,
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
	var buf bytes.Buffer
	term := NewTerminal(&buf)

	srcPort := uint16(8080)
	dstPort := uint16(443)

	pkt := Packet{
		Timestamp: time.Date(2024, 6, 1, 9, 5, 3, 0, time.UTC),
		Protocol:  TCP,
		SrcAddr:   "10.0.0.1",
		DstAddr:   "10.0.0.2",
		SrcPort:   &srcPort,
		DstPort:   &dstPort,
	}
	term.WritePacket(context.Background(), pkt) //nolint

	out := buf.String()
	if !strings.Contains(out, "09:05:03") {
		t.Fatalf("output = %q; want timestamp 09:05:03", out)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("output has %d lines; want 1", len(lines))
	}
}

func TestTerminal_NewTerminal_NilWriter(t *testing.T) {
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

	pkt := Packet{
		Timestamp: time.Now(),
		Protocol:  OTHER,
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

	if strings.Contains(out, " -> ") {
		t.Fatalf("output = %q; OTHER packet should not contain port info", out)
	}
}
