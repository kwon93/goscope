package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/kwon93/goscope/internal/adapter/presenter"
	"github.com/kwon93/goscope/internal/app"
	domcapture "github.com/kwon93/goscope/internal/domain/capture"
	infpcap "github.com/kwon93/goscope/internal/infrastructure/pcap"
)

// Run은 CLI 인자를 파싱하고 의존성을 조립한 뒤 유스케이스를 실행한다.
// 종료 코드를 반환한다.
func Run(ctx context.Context, args []string, in io.Reader, out, errOut io.Writer) int {
	cfg, err := ParseConfig(args, in, out)
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

	uc := app.CapturePackets{
		Source: infpcap.NewEngine(),
		Sinks:  sinks,
	}

	fmt.Fprintln(out, "패킷 캡처 시작 (Ctrl+C를 눌러 종료하세요)")
	if err := uc.Run(ctx, app.CapturePacketsInput{
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

// buildSinks는 Config를 기반으로 PacketSink 목록과 정리 함수를 반환한다.
func buildSinks(cfg Config, out io.Writer) ([]domcapture.PacketSink, func(), error) {
	sinks := []domcapture.PacketSink{presenter.NewTerminal(out)}
	cleanup := func() {}

	if cfg.OutFile == "" {
		return sinks, cleanup, nil
	}

	f, err := os.Create(cfg.OutFile)
	if err != nil {
		return nil, nil, fmt.Errorf("파일 생성 실패: %w", err)
	}

	w, err := infpcap.NewWriter(f)
	if err != nil {
		f.Close()
		return nil, nil, fmt.Errorf("pcap writer 초기화 실패: %w", err)
	}

	sinks = append(sinks, w)
	cleanup = func() { f.Close() }
	return sinks, cleanup, nil
}
