package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/kwon93/goscope/internal/capture"
	"github.com/kwon93/goscope/internal/daemon"
)

// Run은 대화형 입력으로 설정을 받고 의존성을 조립한 뒤 패킷 캡처를 실행한다.
// 종료 코드를 반환한다.
func Run(ctx context.Context, in io.Reader, out, errOut io.Writer) int {
	if len(os.Args) > 1 && os.Args[1] == "--stop" {
		return handleStop(errOut)
	}

	cfg, err := ParseConfig(in, out)
	if err != nil {
		fmt.Fprintf(errOut, "설정 오류: %v\n", err)
		return 1
	}

	if cfg.Background && !daemon.IsDaemon() {
		running, _, _ := daemon.IsRunning()
		if running {
			fmt.Fprintln(errOut, "이미 백그라운드 캡처가 실행 중입니다. 중지하려면: goscope --stop")
			return 1
		}
		cfgJSON, err := json.Marshal(cfg)
		if err != nil {
			fmt.Fprintf(errOut, "설정 직렬화 실패: %v\n", err)
			return 1
		}
		if err := daemon.Daemonize(string(cfgJSON)); err != nil {
			fmt.Fprintf(errOut, "daemon 시작 실패: %v\n", err)
			return 1
		}
		// 부모 프로세스는 Daemonize 내부에서 os.Exit(0). 여기 도달하면 자식.
	}

	if daemon.IsDaemon() {
		if err := daemon.WritePID(); err != nil {
			fmt.Fprintf(errOut, "PID 파일 기록 실패: %v\n", err)
			return 1
		}
		defer daemon.CleanPID() //nolint
	}

	sinks, cleanup, err := buildSinks(cfg, out)
	if err != nil {
		fmt.Fprintf(errOut, "%v\n", err)
		return 1
	}
	defer cleanup()

	if !cfg.TimerStart.IsZero() {
		wait := time.Until(cfg.TimerStart)
		if wait > 0 {
			fmt.Fprintf(out, "캡처 시작까지 대기 중... (%s 시작)\n", cfg.TimerStart.Format("15:04"))
			select {
			case <-time.After(wait):
			case <-ctx.Done():
				return 0
			}
		}
	}

	captureCtx := ctx
	if !cfg.TimerEnd.IsZero() {
		var cancel context.CancelFunc
		captureCtx, cancel = context.WithDeadline(ctx, cfg.TimerEnd)
		defer cancel()
	}

	if !daemon.IsDaemon() {
		if !cfg.TimerEnd.IsZero() {
			fmt.Fprintf(out, "패킷 캡처 시작 (종료 시각: %s)\n", cfg.TimerEnd.Format("15:04"))
		} else {
			fmt.Fprintln(out, "패킷 캡처 시작 (Ctrl+C를 눌러 종료하세요)")
		}
	}
	if err := capture.Run(captureCtx, capture.NewEngine(), sinks, capture.Request{
		Interface: cfg.Interface,
		Filter:    cfg.Filter,
		Snaplen:   cfg.Snaplen,
		Promisc:   cfg.Promisc,
	}); err != nil {
		fmt.Fprintf(errOut, "오류: %v\n", err)
		return 1
	}

	if ctx.Err() != nil && !daemon.IsDaemon() {
		fmt.Fprintln(out, "\n캡처가 종료되었습니다. Enter를 눌러 닫으세요...")
		bufio.NewReader(os.Stdin).ReadString('\n') //nolint
	}
	return 0
}

func handleStop(errOut io.Writer) int {
	if err := daemon.Stop(); err != nil {
		fmt.Fprintf(errOut, "중지 실패: %v\n", err)
		return 1
	}
	fmt.Fprintln(errOut, "캡처가 중지되었습니다.")
	return 0
}

func buildSinks(cfg Config, out io.Writer) ([]capture.Sink, func(), error) {
	var sinks []capture.Sink
	var closers []func() error

	if !daemon.IsDaemon() {
		sinks = append(sinks, capture.NewTerminal(out))
	}

	if cfg.OutFile != "" {
		f, err := os.OpenFile(cfg.OutFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return nil, nil, fmt.Errorf("파일 생성 실패: %w", err)
		}
		w, err := capture.NewWriter(f)
		if err != nil {
			f.Close()
			return nil, nil, fmt.Errorf("pcap writer 초기화 실패: %w", err)
		}
		sinks = append(sinks, w)
		closers = append(closers, f.Close)
	}

	if cfg.RotateDir != "" {
		rw, err := capture.NewRotatingWriter(cfg.RotateDir, cfg.RotatePrefix, cfg.RotateInterval)
		if err != nil {
			return nil, nil, fmt.Errorf("rotating writer 초기화 실패: %w", err)
		}
		sinks = append(sinks, rw)
		closers = append(closers, rw.Close)
	}

	cleanup := func() {
		for _, c := range closers {
			c() //nolint
		}
	}
	return sinks, cleanup, nil
}
