package capture

import (
	"context"
	"fmt"
	"io"
	"os"
)

// Terminal은 패킷을 사람이 읽을 수 있는 형태로 출력하는 Sink 구현체다.
type Terminal struct {
	out io.Writer
}

// NewTerminal은 Terminal을 생성한다. w가 nil이면 os.Stdout을 사용한다.
func NewTerminal(w io.Writer) *Terminal {
	if w == nil {
		w = os.Stdout
	}
	return &Terminal{out: w}
}

// WritePacket은 패킷 정보를 터미널에 한 줄로 출력한다.
func (t *Terminal) WritePacket(_ context.Context, pkt Packet) error {
	ts := pkt.Timestamp.Format("15:04:05")

	src := pkt.SrcAddr
	dst := pkt.DstAddr
	if pkt.SrcPort != nil {
		src = fmt.Sprintf("%s:%d", pkt.SrcAddr, *pkt.SrcPort)
	}
	if pkt.DstPort != nil {
		dst = fmt.Sprintf("%s:%d", pkt.DstAddr, *pkt.DstPort)
	}

	fmt.Fprintf(t.out, "[%s] %s  %s → %s\n", ts, pkt.Protocol, src, dst)
	return nil
}
