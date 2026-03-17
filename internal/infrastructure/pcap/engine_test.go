package pcap

import (
	"context"
	"testing"

	domcapture "github.com/kwon93/goscope/internal/domain/capture"
)

func TestEngine_Capture_InvalidInterface(t *testing.T) {
	e := NewEngine()
	req := domcapture.CaptureRequest{
		Interface: "nonexistent_iface_xyz",
		Snaplen:   1600,
		Promisc:   true,
	}
	_, err := e.Capture(context.Background(), req)
	if err == nil {
		t.Fatal("Capture() with invalid interface: got nil error; want error")
	}
}

func TestEngine_Capture_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	e := NewEngine()
	req := domcapture.CaptureRequest{
		Interface: "nonexistent_iface_xyz",
		Snaplen:   1600,
		Promisc:   true,
	}
	_, err := e.Capture(ctx, req)
	if err == nil {
		t.Fatal("Capture() with cancelled context and invalid interface: got nil error; want error")
	}
}
