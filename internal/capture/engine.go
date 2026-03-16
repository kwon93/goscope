package capture

import (
	"context"
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

type Engine struct {
	config *config
}

func NewEngine(opts ...Option) *Engine {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return &Engine{
		config: cfg,
	}
}

func (e *Engine) Start(ctx context.Context) (<-chan gopacket.Packet, error) {
	// 1. pcap.OpenLive() 로 인터페이스 열기
	handle, err := pcap.OpenLive(e.config.iface, e.config.snaplen, e.config.promisc, pcap.BlockForever)
	if err != nil {
		return nil, fmt.Errorf("인터페이스 열기 실패: %w", err)
	}

	// 2. BPF 필터 적용
	if e.config.filter != "" {
		if err := handle.SetBPFFilter(e.config.filter); err != nil {
			return nil, fmt.Errorf("BPF 필터 적용 실패: %w", err)
		}
	}

	// 3. 채널만들기
	packets := make(chan gopacket.Packet, e.config.chBuffer)

	// 4. 고루틴 시작
	go func() {
		defer close(packets)
		defer handle.Close()

		src := gopacket.NewPacketSource(handle, handle.LinkType())

		for {
			select {
			case <-ctx.Done():
				return
			case packet := <-src.Packets():
				packets <- packet
			}
		}
	}()

	// 5. 채널 반환
	return packets, nil
}
