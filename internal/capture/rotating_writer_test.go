package capture

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newTestRotatingWriter(t *testing.T, interval time.Duration, now func() time.Time) *RotatingWriter {
	t.Helper()
	rw, err := NewRotatingWriter(t.TempDir(), "cap", interval)
	if err != nil {
		t.Fatalf("NewRotatingWriter: %v", err)
	}
	rw.now = now
	return rw
}

func TestRotatingWriter_FirstWrite_CreatesFile(t *testing.T) {
	fixed := time.Date(2026, 3, 19, 14, 0, 0, 0, time.UTC)
	rw := newTestRotatingWriter(t, time.Hour, func() time.Time { return fixed })
	defer rw.Close()

	if err := rw.WritePacket(context.Background(), Packet{}); err != nil {
		t.Fatalf("WritePacket: %v", err)
	}
	if rw.current == nil {
		t.Fatal("current file: got nil; want non-nil")
	}
}

func TestRotatingWriter_FileName_ContainsTimestamp(t *testing.T) {
	dir := t.TempDir()
	fixed := time.Date(2026, 3, 19, 14, 30, 0, 0, time.UTC)

	rw, err := NewRotatingWriter(dir, "cap", time.Hour)
	if err != nil {
		t.Fatalf("NewRotatingWriter: %v", err)
	}
	rw.now = func() time.Time { return fixed }
	defer rw.Close()

	if err := rw.WritePacket(context.Background(), Packet{}); err != nil {
		t.Fatalf("WritePacket: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("files in dir = %d; want 1", len(entries))
	}
	// boundary = Truncate(1h) = 14:00:00
	if !strings.Contains(entries[0].Name(), "2026-03-19T14-00-00") {
		t.Fatalf("filename = %q; want timestamp 2026-03-19T14-00-00", entries[0].Name())
	}
}

func TestRotatingWriter_NoRotation_SameFile(t *testing.T) {
	fixed := time.Date(2026, 3, 19, 14, 0, 0, 0, time.UTC)
	rw := newTestRotatingWriter(t, time.Hour, func() time.Time { return fixed })
	defer rw.Close()

	if err := rw.WritePacket(context.Background(), Packet{}); err != nil {
		t.Fatalf("first WritePacket: %v", err)
	}
	first := rw.current

	if err := rw.WritePacket(context.Background(), Packet{}); err != nil {
		t.Fatalf("second WritePacket: %v", err)
	}
	if rw.current != first {
		t.Fatal("current file changed; want same file within interval")
	}
}

func TestRotatingWriter_Rotation_CreatesNewFile(t *testing.T) {
	dir := t.TempDir()
	tick := time.Date(2026, 3, 19, 14, 0, 0, 0, time.UTC)

	rw, err := NewRotatingWriter(dir, "cap", time.Hour)
	if err != nil {
		t.Fatalf("NewRotatingWriter: %v", err)
	}
	rw.now = func() time.Time { return tick }
	defer rw.Close()

	if err := rw.WritePacket(context.Background(), Packet{}); err != nil {
		t.Fatalf("first WritePacket: %v", err)
	}
	first := rw.current

	tick = tick.Add(2 * time.Hour)

	if err := rw.WritePacket(context.Background(), Packet{}); err != nil {
		t.Fatalf("second WritePacket: %v", err)
	}
	if rw.current == first {
		t.Fatal("current file did not rotate; want new file after interval")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("files in dir = %d; want 2 after rotation", len(entries))
	}
}

func TestRotatingWriter_Close_NilSafe(t *testing.T) {
	rw, err := NewRotatingWriter(t.TempDir(), "cap", time.Hour)
	if err != nil {
		t.Fatalf("NewRotatingWriter: %v", err)
	}
	if err := rw.Close(); err != nil {
		t.Fatalf("Close on empty writer: %v", err)
	}
}

func TestRotatingWriter_NewFile_HasPcapHeader(t *testing.T) {
	dir := t.TempDir()
	fixed := time.Date(2026, 3, 19, 14, 0, 0, 0, time.UTC)

	rw, err := NewRotatingWriter(dir, "cap", time.Hour)
	if err != nil {
		t.Fatalf("NewRotatingWriter: %v", err)
	}
	rw.now = func() time.Time { return fixed }

	if err := rw.WritePacket(context.Background(), Packet{}); err != nil {
		t.Fatalf("WritePacket: %v", err)
	}
	rw.Close() //nolint

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, entries[0].Name()))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(data) < 4 {
		t.Fatal("pcap file too small; want at least pcap global header")
	}
	// pcap magic: 0xa1b2c3d4 (big-endian) or 0xd4c3b2a1 (little-endian on disk)
	if data[0] != 0xd4 && data[0] != 0xa1 {
		t.Fatalf("pcap magic first byte = %#x; want 0xd4 or 0xa1", data[0])
	}
}
