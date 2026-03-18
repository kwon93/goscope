package capture

import (
	"bytes"
	"context"
	"testing"
	"time"
)

func TestWriter_WritePacket(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}

	pkt := Packet{
		Timestamp:   time.Now(),
		RawData:     []byte{0x00, 0x01, 0x02, 0x03},
		CaptureLen:  4,
		OriginalLen: 4,
	}

	if err := w.WritePacket(context.Background(), pkt); err != nil {
		t.Fatalf("WritePacket: %v", err)
	}

	// pcap global header(24 bytes) + record header(16 bytes) + data(4 bytes) = 44 bytes
	if buf.Len() < 44 {
		t.Fatalf("buf.Len() = %d; want >= 44", buf.Len())
	}
}

func TestWriter_WriteMultiplePackets(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}

	for i := 0; i < 3; i++ {
		pkt := Packet{
			Timestamp:   time.Now(),
			RawData:     []byte{0x00, 0x01},
			CaptureLen:  2,
			OriginalLen: 2,
		}
		if err := w.WritePacket(context.Background(), pkt); err != nil {
			t.Fatalf("WritePacket[%d]: %v", i, err)
		}
	}

	// 24 + 3*(16+2) = 24 + 48 = 72 bytes
	if buf.Len() < 72 {
		t.Fatalf("buf.Len() = %d; want >= 72", buf.Len())
	}
}

func TestNewWriter_FileHeader_MagicNumber(t *testing.T) {
	var buf bytes.Buffer
	if _, err := NewWriter(&buf); err != nil {
		t.Fatalf("NewWriter: %v", err)
	}

	b := buf.Bytes()
	if len(b) < 4 {
		t.Fatalf("header too short: got %d bytes; want >= 4", len(b))
	}

	leMagic := []byte{0xd4, 0xc3, 0xb2, 0xa1}
	beMagic := []byte{0xa1, 0xb2, 0xc3, 0xd4}
	if !bytes.Equal(b[:4], leMagic) && !bytes.Equal(b[:4], beMagic) {
		t.Fatalf("pcap magic = %x; want %x or %x", b[:4], leMagic, beMagic)
	}
}

func TestWriter_WritePacket_EmptyRawData(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(&buf)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}

	pkt := Packet{
		Timestamp:   time.Now(),
		RawData:     []byte{},
		CaptureLen:  0,
		OriginalLen: 0,
	}
	if err := w.WritePacket(context.Background(), pkt); err != nil {
		t.Fatalf("WritePacket with empty data: %v", err)
	}
}
