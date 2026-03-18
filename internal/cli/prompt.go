package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/kwon93/goscope/internal/capture"
	"github.com/kwon93/goscope/internal/netif"
)

// Config는 대화형 입력으로 결정되는 실행 설정이다.
type Config struct {
	Interface string
	Filter    string
	OutFile   string
	Snaplen   int32
	Promisc   bool
}

// ParseConfig는 in/out으로 사용자 입력을 받아 Config를 반환한다.
func ParseConfig(in io.Reader, out io.Writer) (Config, error) {
	cfg := Config{
		Snaplen: 1600,
		Promisc: true,
	}

	name, err := promptOutFile(in, out)
	if err != nil {
		return Config{}, err
	}
	cfg.OutFile = name

	iface, err := promptInterface(in, out)
	if err != nil {
		return Config{}, err
	}
	cfg.Interface = iface

	filter, err := promptFilter(in, out)
	if err != nil {
		return Config{}, err
	}
	cfg.Filter = filter

	return cfg, nil
}

func promptOutFile(in io.Reader, out io.Writer) (string, error) {
	fmt.Fprint(out, "저장할 파일명을 입력하세요 (생략시 Enter): ")
	name, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("파일명 입력 오류: %w", err)
	}
	name = strings.TrimSpace(name)
	if strings.ContainsAny(name, "\x00") {
		return "", fmt.Errorf("파일명에 사용할 수 없는 문자가 포함되어 있습니다")
	}
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
	if len(ifaces) == 0 {
		return "", fmt.Errorf("사용 가능한 네트워크 인터페이스가 없습니다")
	}

	fmt.Fprintln(out, "네트워크 인터페이스를 선택하세요:")
	for i, iface := range ifaces {
		fmt.Fprintf(out, "  [%d] %s (%s)\n", i, iface.Description, iface.Name)
	}

	reader := bufio.NewReader(in)
	for {
		fmt.Fprintf(out, "선택 (0~%d, 기본값 0): ", len(ifaces)-1)
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return "", fmt.Errorf("인터페이스 입력 오류: %w", err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			return ifaces[0].Name, nil
		}
		var num int
		if _, err := fmt.Sscan(line, &num); err != nil || num < 0 || num >= len(ifaces) {
			fmt.Fprintf(out, "올바른 번호를 입력하세요 (0~%d)\n", len(ifaces)-1)
			continue
		}
		return ifaces[num].Name, nil
	}
}

func promptFilter(in io.Reader, out io.Writer) (string, error) {
	reader := bufio.NewReader(in)
	for {
		fmt.Fprint(out, "BPF 필터를 입력하세요 (ex: tcp port 80, 생략시 Enter): ")
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return "", fmt.Errorf("필터 입력 오류: %w", err)
		}
		filter := strings.TrimSpace(line)
		if validateErr := capture.ValidateBPFFilter(filter); validateErr != nil {
			fmt.Fprintf(out, "잘못된 BPF 필터 입력입니다. (잘모른다면 생략하시는것도...ㅎㅎ): %v\n", validateErr)
			continue
		}
		return filter, nil
	}
}
