package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/kwon93/goscope/internal/capture"
)

// Run은 대화형 입력으로 설정을 받고 의존성을 조립한 뒤 패킷 캡처를 실행한다.
// 종료 코드를 반환한다.
func Run(ctx context.Context, in io.Reader, out, errOut io.Writer) int {
	cfg, err := ParseConfig(in, out)
	if err != nil {
		fmt.Fprintf(errOut, "설정 오류: %v\n", err)
		return 1
	}

	sinks, cleanup, err := buildSinks(cfg, out)
	if err != nil {
		fmt.Fprintf(errOut, "%v\n", err)
		return 1
	}
	defer cleanup()

	fmt.Fprintln(out, "패킷 캡처 시작 (Ctrl+C를 눌러 종료하세요)")
	if err := capture.Run(ctx, capture.NewEngine(), sinks, capture.Request{
		Interface: cfg.Interface,
		Filter:    cfg.Filter,
		Snaplen:   cfg.Snaplen,
		Promisc:   cfg.Promisc,
	}); err != nil {
		fmt.Fprintf(errOut, "오류: %v\n", err)
		return 1
	}
	return 0
}

// buildSinks는 Config를 기반으로 Sink 목록과 정리 함수를 반환한다.
func buildSinks(cfg Config, out io.Writer) ([]capture.Sink, func(), error) {
	sinks := []capture.Sink{capture.NewTerminal(out)}
	cleanup := func() {}

	if cfg.OutFile == "" {
		return sinks, cleanup, nil
	}

	f, err := os.Create(cfg.OutFile)
	if err != nil {
		return nil, nil, fmt.Errorf("파일 생성 실패: %w", err)
	}

	w, err := capture.NewWriter(f)
	if err != nil {
		f.Close()
		return nil, nil, fmt.Errorf("pcap writer 초기화 실패: %w", err)
	}

	sinks = append(sinks, w)
	cleanup = func() { f.Close() }
	return sinks, cleanup, nil
}
