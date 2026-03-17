package packet

// Protocol은 전송 계층 프로토콜을 나타낸다.
type Protocol string

const (
	TCP   Protocol = "tcp"
	UDP   Protocol = "udp"
	OTHER Protocol = "other"
)
