package netif

import (
	"fmt"

	"github.com/google/gopacket/pcap"
)

// Interface는 네트워크 인터페이스의 이름과 설명을 담는다.
type Interface struct {
	Name        string
	Description string
}

// List는 시스템의 모든 네트워크 인터페이스를 반환한다.
func List() ([]Interface, error) {
	devs, err := pcap.FindAllDevs()
	if err != nil {
		return nil, fmt.Errorf("list network interfaces: %w", err)
	}

	ifaces := make([]Interface, len(devs))
	for i, d := range devs {
		ifaces[i] = Interface{Name: d.Name, Description: d.Description}
	}
	return ifaces, nil
}
