package capture

import (
	"github.com/google/gopacket/layers"
	gpcap "github.com/google/gopacket/pcap"
)

// ValidateBPFFilter는 BPF 표현식을 컴파일해 문법을 검증한다.
func ValidateBPFFilter(filter string) error {
	if filter == "" {
		return nil
	}
	_, err := gpcap.CompileBPFFilter(layers.LinkTypeEthernet, 1600, filter)
	return err
}
