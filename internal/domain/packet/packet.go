package packet

import "time"

// Packet은 캡처된 네트워크 패킷의 도메인 표현이다.
// gopacket 타입을 직접 노출하지 않는다.
type Packet struct {
	Timestamp time.Time
	Protocol  Protocol
	SrcAddr   string
	DstAddr   string
	SrcPort   *uint16 // TCP/UDP만 존재; OTHER는 nil
	DstPort   *uint16

	// pcap 파일 쓰기를 위한 원본 데이터
	RawData     []byte
	CaptureLen  int
	OriginalLen int
}
