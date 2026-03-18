package cli

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/kwon93/goscope/internal/netif"
)

// Config는 CLI 플래그와 인터랙티브 입력으로 결정되는 실행 설정이다.
type Config struct {
	Interface string
	Filter    string
	OutFile   string
	Snaplen   int32
	Promisc   bool
}

// ParseConfig는 args를 파싱하고, 필요한 경우 in/out으로 사용자 입력을 받아 Config를 반환한다.
func ParseConfig(args []string, in io.Reader, out io.Writer) (Config, error) {
	fs := flag.NewFlagSet("goscope", flag.ContinueOnError)
	fs.SetOutput(out)

	iface := fs.String("i", "", "네트워크 인터페이스")
	filter := fs.String("f", "", "BPF 필터 (ex: tcp port 80)")
	outFile := fs.String("w", "", "저장할 pcap 파일명 (ex: capture.pcap)")
	snaplen := fs.Int("snaplen", 1600, "캡처 최대 바이트 수")
	promisc := fs.Bool("promisc", true, "무차별 모드")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	cfg := Config{
		Interface: *iface,
		Filter:    *filter,
		OutFile:   *outFile,
		Snaplen:   int32(*snaplen),
		Promisc:   *promisc,
	}

	if cfg.OutFile == "" {
		name, err := promptOutFile(in, out)
		if err != nil {
			return Config{}, err
		}
		cfg.OutFile = name
	}

	if cfg.Interface == "" {
		selected, err := promptInterface(in, out)
		if err != nil {
			return Config{}, err
		}
		cfg.Interface = selected
	}

	return cfg, nil
}

func promptOutFile(in io.Reader, out io.Writer) (string, error) {
	fmt.Fprint(out, "저장할 파일명을 입력하세요 (생략시 Enter): ")
	name, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && err.Error() != "EOF" {
		return "", err
	}
	name = strings.TrimSpace(name)
	if name != "" && !strings.HasSuffix(name, ".pcap") {
		name += ".pcap"
	}
	return name, nil
}

func promptInterface(in io.Reader, out io.Writer) (string, error) {
	ifaces, err := netif.List()
	if err != nil {
		return "", err
	}
	for i, iface := range ifaces {
		fmt.Fprintf(out, "[%d] %s (%s)\n", i, iface.Description, iface.Name)
	}
	var num int
	fmt.Fscan(in, &num)
	return ifaces[num].Name, nil
}
