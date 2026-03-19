package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/kwon93/goscope/internal/capture"
	"github.com/kwon93/goscope/internal/daemon"
	"github.com/kwon93/goscope/internal/netif"
)

// Config는 대화형 입력으로 결정되는 실행 설정이다.
type Config struct {
	Interface      string
	Filter         string
	OutFile        string
	RotateDir      string
	RotateInterval time.Duration
	RotatePrefix   string
	Background     bool
	Snaplen        int32
	Promisc        bool
}

// ParseConfig는 in/out으로 사용자 입력을 받아 Config를 반환한다.
// 데몬 자식 프로세스인 경우 환경 변수에서 직렬화된 Config를 읽는다.
func ParseConfig(in io.Reader, out io.Writer) (Config, error) {
	if daemon.IsDaemon() {
		return parseConfigFromEnv()
	}

	cfg := Config{
		Snaplen: 1600,
		Promisc: true,
	}

	if err := promptOutput(in, out, &cfg); err != nil {
		return Config{}, err
	}

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

	if cfg.RotateDir != "" {
		bg, err := promptBackground(in, out)
		if err != nil {
			return Config{}, err
		}
		cfg.Background = bg
	}

	return cfg, nil
}

func parseConfigFromEnv() (Config, error) {
	raw := daemon.DaemonConfig()
	if raw == "" {
		return Config{}, fmt.Errorf("데몬 설정 환경 변수가 없습니다")
	}
	var cfg Config
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return Config{}, fmt.Errorf("데몬 설정 파싱 실패: %w", err)
	}
	return cfg, nil
}

func promptOutput(in io.Reader, out io.Writer, cfg *Config) error {
	fmt.Fprintln(out, "저장 방식을 선택하세요:")
	fmt.Fprintln(out, "  [1] 저장 안 함 (터미널 출력만)")
	fmt.Fprintln(out, "  [2] 단일 파일로 저장")
	fmt.Fprintln(out, "  [3] 시간대별 파일 분할 저장")

	reader := bufio.NewReader(in)
	for {
		fmt.Fprint(out, "선택 (기본값 1): ")
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("저장 방식 입력 오류: %w", err)
		}
		switch strings.TrimSpace(line) {
		case "", "1":
			return nil
		case "2":
			// reader는 io.Reader를 구현하므로 promptOutFile에 전달 가능
			name, err := promptOutFile(reader, out)
			if err != nil {
				return err
			}
			cfg.OutFile = name
			return nil
		case "3":
			dir, err := promptRotateDir(reader, out)
			if err != nil {
				return err
			}
			interval, err := promptRotateInterval(reader, out)
			if err != nil {
				return err
			}
			cfg.RotateDir = dir
			cfg.RotateInterval = interval
			cfg.RotatePrefix = "capture"
			return nil
		default:
			fmt.Fprintln(out, "1, 2, 3 중 하나를 입력하세요.")
		}
	}
}

func promptOutFile(in io.Reader, out io.Writer) (string, error) {
	fmt.Fprint(out, "저장할 파일명을 입력하세요: ")
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

func promptRotateDir(reader *bufio.Reader, out io.Writer) (string, error) {
	fmt.Fprint(out, "저장 디렉터리 (기본값 현재 디렉터리 .): ")
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("디렉터리 입력 오류: %w", err)
	}
	dir := strings.TrimSpace(line)
	if dir == "" {
		dir = "."
	}
	return dir, nil
}

func promptRotateInterval(reader *bufio.Reader, out io.Writer) (time.Duration, error) {
	for {
		fmt.Fprint(out, "파일 분할 주기 (예: 1h, 30m, 기본값 1h): ")
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return 0, fmt.Errorf("주기 입력 오류: %w", err)
		}
		s := strings.TrimSpace(line)
		if s == "" {
			return time.Hour, nil
		}
		d, err := time.ParseDuration(s)
		if err != nil || d < time.Minute {
			fmt.Fprintln(out, "1분 이상의 유효한 주기를 입력하세요 (예: 30m, 1h, 6h).")
			continue
		}
		return d, nil
	}
}

func promptBackground(in io.Reader, out io.Writer) (bool, error) {
	fmt.Fprint(out, "백그라운드로 실행하시겠습니까? (터미널을 닫아도 계속 실행됩니다) [y/N]: ")
	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("백그라운드 입력 오류: %w", err)
	}
	return strings.EqualFold(strings.TrimSpace(line), "y"), nil
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
