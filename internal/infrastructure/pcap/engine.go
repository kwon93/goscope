package pcap

import (
	"context"
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	gpcap "github.com/google/gopacket/pcap"

	domcapture "github.com/kwon93/goscope/internal/domain/capture"
	"github.com/kwon93/goscope/internal/domain/packet"
)

// Engine은 gopacket/pcap 기반 PacketSource 구현체다.
type Engine struct{}

// NewEngine은 Engine을 생성한다.
func NewEngine() *Engine {
	return &Engine{}
}

// Capture는 지정된 인터페이스에서 패킷을 캡처해 채널로 반환한다.
// NetworkLayer가 없는 패킷은 채널에 전달하지 않는다.
func (e *Engine) Capture(ctx context.Context, req domcapture.CaptureRequest) (<-chan packet.Packet, error) {
	handle, err := gpcap.OpenLive(req.Interface, req.Snaplen, req.Promisc, gpcap.BlockForever)
	if err != nil {
		return nil, fmt.Errorf("open interface %q: %w", req.Interface, err)
	}

	if req.Filter != "" {
		if err := handle.SetBPFFilter(req.Filter); err != nil {
			handle.Close()
			return nil, fmt.Errorf("set BPF filter %q: %w", req.Filter, err)
		}
	}

	out := make(chan packet.Packet, req.ChBuffer)

	go func() {
		defer close(out)
		defer handle.Close()

		src := gopacket.NewPacketSource(handle, handle.LinkType())

		for {
			select {
			case <-ctx.Done():
				return
			case raw, ok := <-src.Packets():
				if !ok {
					return
				}
				pkt, ok := toDomainPacket(raw)
				if !ok {
					continue
				}
				select {
				case out <- pkt:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out, nil
}

// toDomainPacket은 gopacket.Packet을 domain.Packet으로 변환한다.
// NetworkLayer가 없으면 (zero, false)를 반환한다.
func toDomainPacket(raw gopacket.Packet) (packet.Packet, bool) {
	net := raw.NetworkLayer()
	if net == nil {
		return packet.Packet{}, false
	}

	src, dst := net.NetworkFlow().Endpoints()
	ci := raw.Metadata().CaptureInfo

	pkt := packet.Packet{
		Timestamp:   ci.Timestamp,
		Protocol:    packet.OTHER,
		SrcAddr:     src.String(),
		DstAddr:     dst.String(),
		RawData:     raw.Data(),
		CaptureLen:  ci.CaptureLength,
		OriginalLen: ci.Length,
	}

	if tcp, ok := raw.Layer(layers.LayerTypeTCP).(*layers.TCP); ok && tcp != nil {
		pkt.Protocol = packet.TCP
		srcPort := uint16(tcp.SrcPort)
		dstPort := uint16(tcp.DstPort)
		pkt.SrcPort = &srcPort
		pkt.DstPort = &dstPort
	} else if udp, ok := raw.Layer(layers.LayerTypeUDP).(*layers.UDP); ok && udp != nil {
		pkt.Protocol = packet.UDP
		srcPort := uint16(udp.SrcPort)
		dstPort := uint16(udp.DstPort)
		pkt.SrcPort = &srcPort
		pkt.DstPort = &dstPort
	}

	return pkt, true
}
