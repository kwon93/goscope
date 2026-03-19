package daemon

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

// withTempPIDFile은 테스트 중 PID 파일 경로를 임시 경로로 교체한다.
func withTempPIDFile(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir() + "/goscope.pid"
	restore := setTestPIDFile(func() string { return tmp })
	t.Cleanup(restore)
	return tmp
}

func TestIsDaemon_False_WhenEnvNotSet(t *testing.T) {
	os.Unsetenv(DaemonEnvKey)
	if IsDaemon() {
		t.Fatal("IsDaemon() = true; want false when env not set")
	}
}

func TestIsDaemon_True_WhenEnvSet(t *testing.T) {
	t.Setenv(DaemonEnvKey, "1")
	if !IsDaemon() {
		t.Fatal("IsDaemon() = false; want true when env set to 1")
	}
}

func TestDaemonConfig_ReturnsEnvValue(t *testing.T) {
	t.Setenv(DaemonConfigKey, `{"Interface":"eth0"}`)
	if got := DaemonConfig(); got != `{"Interface":"eth0"}` {
		t.Fatalf("DaemonConfig() = %q; want config JSON", got)
	}
}

func TestWritePID_CreatesFileWithCurrentPID(t *testing.T) {
	tmp := withTempPIDFile(t)

	if err := WritePID(); err != nil {
		t.Fatalf("WritePID: %v", err)
	}

	data, err := os.ReadFile(tmp)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		t.Fatalf("parse pid: %v", err)
	}
	if pid != os.Getpid() {
		t.Fatalf("PID in file = %d; want %d", pid, os.Getpid())
	}
}

func TestCleanPID_RemovesFile(t *testing.T) {
	tmp := withTempPIDFile(t)
	os.WriteFile(tmp, []byte("12345"), 0644) //nolint

	if err := CleanPID(); err != nil {
		t.Fatalf("CleanPID: %v", err)
	}
	if _, err := os.Stat(tmp); !os.IsNotExist(err) {
		t.Fatal("PID file still exists after CleanPID")
	}
}

func TestIsRunning_False_WhenNoPIDFile(t *testing.T) {
	withTempPIDFile(t) // 파일 없는 임시 경로로 교체

	running, _, err := IsRunning()
	if err != nil {
		t.Fatalf("IsRunning: %v", err)
	}
	if running {
		t.Fatal("IsRunning() = true; want false when no PID file")
	}
}

func TestIsRunning_False_WhenProcessGone(t *testing.T) {
	tmp := withTempPIDFile(t)
	os.WriteFile(tmp, []byte("999999999"), 0644) //nolint

	running, _, _ := IsRunning()
	if running {
		t.Fatal("IsRunning() = true; want false for non-existent PID")
	}
}

func TestPIDFile_ReturnsNonEmptyPath(t *testing.T) {
	if PIDFile() == "" {
		t.Fatal("PIDFile() = empty string; want a path")
	}
}
