package capture

import (
	"context"

	"github.com/kwon93/goscope/internal/domain/packet"
)

// CaptureRequest는 캡처 시작에 필요한 입력 파라미터다.
type CaptureRequest struct {
	Interface string
	Filter    string
	Snaplen   int32
	Promisc   bool
	ChBuffer  int
}

// PacketSource는 네트워크 패킷을 생산하는 포트다.
// 구현은 infrastructure/pcap 계층에서 제공한다.
type PacketSource interface {
	Capture(ctx context.Context, req CaptureRequest) (<-chan packet.Packet, error)
}

// PacketSink는 패킷을 소비(출력 또는 저장)하는 포트다.
type PacketSink interface {
	WritePacket(ctx context.Context, pkt packet.Packet) error
}
