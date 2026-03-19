package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func init() {
	impl = &unixRunner{}
}

type unixRunner struct {
	pidFileFn func() string // 테스트에서 교체 가능
}

func (u *unixRunner) pidFile() string {
	if u.pidFileFn != nil {
		return u.pidFileFn()
	}
	return "/tmp/goscope.pid"
}

func (u *unixRunner) start(cfgJSON string) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("executable path: %w", err)
	}

	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Env = append(os.Environ(),
		DaemonEnvKey+"=1",
		DaemonConfigKey+"="+cfgJSON,
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start daemon: %w", err)
	}

	fmt.Printf("백그라운드 캡처 시작. PID: %d\n중지하려면: goscope --stop\n", cmd.Process.Pid)
	os.Exit(0)
	return nil
}

func (u *unixRunner) isRunning(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}

// setTestPIDFile은 테스트에서 PID 파일 경로를 교체하고 복원 함수를 반환한다.
func setTestPIDFile(fn func() string) func() {
	r := impl.(*unixRunner)
	orig := r.pidFileFn
	r.pidFileFn = fn
	return func() { r.pidFileFn = orig }
}

func (u *unixRunner) stop(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process %d: %w", pid, err)
	}
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("send SIGTERM to %d: %w", pid, err)
	}
	return nil
}
