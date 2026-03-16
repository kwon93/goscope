package output

import (
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/kwon93/goscope/internal/types"
)

func Print(pkt gopacket.Packet) {
	// 1. 시간
	ts := pkt.Metadata().Timestamp.Format("15:04:05")

	// 2. IP Layer
	net := pkt.NetworkLayer()
	if net == nil {
		return
	}
	src, dst := net.NetworkFlow().Endpoints()

	// 3. TCP, UPD Check
	proto := types.OTHER
	var port string

	if tcp := pkt.Layer(layers.LayerTypeTCP); tcp != nil {
		t := tcp.(*layers.TCP)
		proto = types.TCP
		port = fmt.Sprintf(":%d -> %d", t.SrcPort, t.DstPort)
	} else if udp := pkt.Layer(layers.LayerTypeUDP); udp != nil {
		u := udp.(*layers.UDP)
		proto = types.UDP
		port = fmt.Sprintf(":%d -> %d", u.SrcPort, u.DstPort)
	}

	// 4. 출력
	fmt.Printf("[%s] %s  %s%s → %s\n", ts, proto, src, port, dst)
}
