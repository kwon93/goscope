package app

import (
	"context"
	"fmt"

	domcapture "github.com/kwon93/goscope/internal/domain/capture"
	"github.com/kwon93/goscope/internal/domain/packet"
)

// CapturePacketsInput은 CapturePackets 유스케이스의 입력 DTO다.
type CapturePacketsInput struct {
	Interface string
	Filter    string
	Snaplen   int32
	Promisc   bool
	ChBuffer  int
}

// CapturePackets는 패킷 캡처 유스케이스다.
type CapturePackets struct {
	Source domcapture.PacketSource
	Sinks  []domcapture.PacketSink
}

// Run은 캡처를 시작하고 ctx가 취소될 때까지 패킷을 모든 Sink에 전달한다.
func (uc CapturePackets) Run(ctx context.Context, in CapturePacketsInput) error {
	req := domcapture.CaptureRequest{
		Interface: in.Interface,
		Filter:    in.Filter,
		Snaplen:   in.Snaplen,
		Promisc:   in.Promisc,
		ChBuffer:  in.ChBuffer,
	}

	packets, err := uc.Source.Capture(ctx, req)
	if err != nil {
		return fmt.Errorf("start capture: %w", err)
	}

	for pkt := range packets {
		if err := uc.fanOut(ctx, pkt); err != nil {
			return err
		}
	}
	return nil
}

// fanOut은 하나의 패킷을 모든 Sink에 순서대로 전달한다.
func (uc CapturePackets) fanOut(ctx context.Context, pkt packet.Packet) error {
	for _, sink := range uc.Sinks {
		if err := sink.WritePacket(ctx, pkt); err != nil {
			return fmt.Errorf("sink write: %w", err)
		}
	}
	return nil
}
