package capture

import (
	"context"
	"fmt"
	"time"
)

// Protocol은 전송 계층 프로토콜을 나타낸다.
type Protocol string

const (
	TCP   Protocol = "tcp"
	UDP   Protocol = "udp"
	OTHER Protocol = "other"
)

// Packet은 캡처된 네트워크 패킷이다.
// gopacket 타입을 직접 노출하지 않는다.
type Packet struct {
	Timestamp   time.Time
	Protocol    Protocol
	SrcAddr     string
	DstAddr     string
	SrcPort     *uint16 // TCP/UDP만 존재; OTHER는 nil
	DstPort     *uint16
	RawData     []byte
	CaptureLen  int
	OriginalLen int
}

// Request는 캡처 시작에 필요한 파라미터다.
type Request struct {
	Interface string
	Filter    string
	Snaplen   int32
	Promisc   bool
	ChBuffer  int
}

// Source는 패킷을 생산하는 인터페이스다.
type Source interface {
	Capture(ctx context.Context, req Request) (<-chan Packet, error)
}

// Sink는 패킷을 소비(출력 또는 저장)하는 인터페이스다.
type Sink interface {
	WritePacket(ctx context.Context, pkt Packet) error
}

// Run은 src에서 패킷을 받아 ctx가 취소될 때까지 모든 sink에 전달한다.
func Run(ctx context.Context, src Source, sinks []Sink, req Request) error {
	packets, err := src.Capture(ctx, req)
	if err != nil {
		return fmt.Errorf("start capture: %w", err)
	}

	for pkt := range packets {
		for _, sink := range sinks {
			if err := sink.WritePacket(ctx, pkt); err != nil {
				return fmt.Errorf("sink write: %w", err)
			}
		}
	}
	return nil
}
