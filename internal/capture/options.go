package capture

type config struct {
	iface    string //network interface 이름
	filter   string //BPF 필터
	snaplen  int32  //패킷 캡쳐 최대 바이트 수
	promisc  bool   //무차별 모드
	chBuffer int    //channel buffer size
}

type Option func(*config)

func WithInterface(iface string) Option {
	return func(c *config) {
		c.iface = iface
	}
}

func WithFilter(filter string) Option {
	return func(c *config) {
		c.filter = filter
	}
}

func WithSnaplen(snaplen int32) Option {
	return func(c *config) {
		c.snaplen = snaplen
	}
}

func WithPromisc(promisc bool) Option {
	return func(c *config) {
		c.promisc = promisc
	}
}

func WithChannelBuffer(chBuffer int) Option {
	return func(c *config) {
		c.chBuffer = chBuffer
	}
}

func defaultConfig() *config {
	return &config{
		iface:   "eth0",
		snaplen: 1600,
		promisc: true,
	}
}
