package daemon

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	// DaemonEnvKey는 데몬 자식 프로세스를 식별하는 환경 변수다.
	DaemonEnvKey = "_GOSCOPE_DAEMON"
	// DaemonConfigKey는 데몬에게 직렬화된 설정을 전달하는 환경 변수다.
	DaemonConfigKey = "_GOSCOPE_CONFIG"
)

// runner는 OS별로 구현되는 데몬 동작 인터페이스다.
type runner interface {
	start(cfgJSON string) error
	isRunning(pid int) bool
	stop(pid int) error
	pidFile() string
}

// impl은 빌드 시 OS에 맞는 구현체로 초기화된다.
var impl runner

// IsDaemon은 현재 프로세스가 데몬 자식인지 반환한다.
func IsDaemon() bool {
	return os.Getenv(DaemonEnvKey) == "1"
}

// DaemonConfig는 부모가 전달한 직렬화 설정 문자열을 반환한다.
func DaemonConfig() string {
	return os.Getenv(DaemonConfigKey)
}

// PIDFile은 PID 파일 경로를 반환한다.
func PIDFile() string {
	return impl.pidFile()
}

// Daemonize는 현재 프로세스를 백그라운드 데몬으로 재시작한다.
// 자식 프로세스가 정상 시작되면 부모는 os.Exit(0)를 호출한다.
func Daemonize(cfgJSON string) error {
	if IsDaemon() {
		return nil
	}
	return impl.start(cfgJSON)
}

// WritePID는 현재 프로세스의 PID를 파일에 기록한다.
func WritePID() error {
	return os.WriteFile(impl.pidFile(), []byte(strconv.Itoa(os.Getpid())), 0644)
}

// CleanPID는 PID 파일을 삭제한다.
func CleanPID() error {
	return os.Remove(impl.pidFile())
}

// IsRunning은 PID 파일이 있고 해당 프로세스가 살아있으면 true와 PID를 반환한다.
func IsRunning() (bool, int, error) {
	data, err := os.ReadFile(impl.pidFile())
	if os.IsNotExist(err) {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, fmt.Errorf("read pid file: %w", err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return false, 0, fmt.Errorf("parse pid: %w", err)
	}

	return impl.isRunning(pid), pid, nil
}

// Stop은 PID 파일을 읽어 해당 프로세스를 종료한다.
func Stop() error {
	running, pid, err := IsRunning()
	if err != nil {
		return err
	}
	if !running {
		return fmt.Errorf("실행 중인 캡처 프로세스가 없습니다")
	}
	return impl.stop(pid)
}
