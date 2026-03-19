package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func init() {
	impl = &winRunner{}
}

type winRunner struct {
	pidFileFn func() string // 테스트에서 교체 가능
}

func (w *winRunner) pidFile() string {
	if w.pidFileFn != nil {
		return w.pidFileFn()
	}
	return filepath.Join(os.TempDir(), "goscope.pid")
}

func (w *winRunner) start(cfgJSON string) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("executable path: %w", err)
	}

	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Env = append(os.Environ(),
		DaemonEnvKey+"=1",
		DaemonConfigKey+"="+cfgJSON,
	)
	// DETACHED_PROCESS: 터미널(콘솔)로부터 분리된 새 프로세스 생성
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | 0x00000008,
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start daemon: %w", err)
	}

	fmt.Printf("백그라운드 캡처 시작. PID: %d\n중지하려면: goscope --stop\n", cmd.Process.Pid)
	os.Exit(0)
	return nil
}

// setTestPIDFile은 테스트에서 PID 파일 경로를 교체하고 복원 함수를 반환한다.
func setTestPIDFile(fn func() string) func() {
	r := impl.(*winRunner)
	orig := r.pidFileFn
	r.pidFileFn = fn
	return func() { r.pidFileFn = orig }
}

func (w *winRunner) isRunning(pid int) bool {
	// tasklist로 PID 존재 여부 확인
	out, err := exec.Command(
		"tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/NH",
	).Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), strconv.Itoa(pid))
}

func (w *winRunner) stop(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process %d: %w", pid, err)
	}
	// Windows는 SIGTERM 미지원 — Kill()로 프로세스 종료
	if err := proc.Kill(); err != nil {
		return fmt.Errorf("kill process %d: %w", pid, err)
	}
	return nil
}
