package capture

import (
	"context"
	"fmt"
	"io"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
)

// Writer는 패킷을 pcap 파일 포맷으로 기록하는 Sink 구현체다.
type Writer struct {
	w *pcapgo.Writer
}

// NewWriter는 w에 pcap 파일 헤더를 쓰고 Writer를 반환한다.
func NewWriter(w io.Writer) (*Writer, error) {
	pw := pcapgo.NewWriter(w)
	if err := pw.WriteFileHeader(65536, layers.LinkTypeEthernet); err != nil {
		return nil, fmt.Errorf("write pcap file header: %w", err)
	}
	return &Writer{w: pw}, nil
}

// WritePacket은 Packet을 pcap 레코드로 기록한다.
func (wr *Writer) WritePacket(_ context.Context, pkt Packet) error {
	ci := gopacket.CaptureInfo{
		Timestamp:     pkt.Timestamp,
		CaptureLength: pkt.CaptureLen,
		Length:        pkt.OriginalLen,
	}
	if err := wr.w.WritePacket(ci, pkt.RawData); err != nil {
		return fmt.Errorf("write packet to pcap: %w", err)
	}
	return nil
}
