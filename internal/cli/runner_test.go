package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/kwon93/goscope/internal/capture"
)

func TestBuildSinks_TerminalOnly(t *testing.T) {
	var out bytes.Buffer
	cfg := Config{}

	sinks, cleanup, err := buildSinks(cfg, &out)
	if err != nil {
		t.Fatalf("buildSinks: %v", err)
	}
	defer cleanup()

	if len(sinks) != 1 {
		t.Fatalf("expected 1 sink (terminal), got %d", len(sinks))
	}
}

func TestBuildSinks_OutFile_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "test.pcap")

	var out bytes.Buffer
	cfg := Config{OutFile: outFile}

	sinks, cleanup, err := buildSinks(cfg, &out)
	if err != nil {
		t.Fatalf("buildSinks: %v", err)
	}
	defer cleanup()

	if len(sinks) != 2 {
		t.Fatalf("expected 2 sinks (terminal + writer), got %d", len(sinks))
	}

	if _, err := os.Stat(outFile); os.IsNotExist(err) {
		t.Fatalf("expected output file %q to be created", outFile)
	}
}

func TestBuildSinks_OutFile_Permissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix file permission bits not enforced on Windows")
	}
	dir := t.TempDir()
	outFile := filepath.Join(dir, "test.pcap")

	var out bytes.Buffer
	cfg := Config{OutFile: outFile}

	_, cleanup, err := buildSinks(cfg, &out)
	if err != nil {
		t.Fatalf("buildSinks: %v", err)
	}
	defer cleanup()

	info, err := os.Stat(outFile)
	if err != nil {
		t.Fatalf("stat output file: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Fatalf("expected file permission 0600, got %04o", perm)
	}
}

func TestBuildSinks_RotateDir(t *testing.T) {
	dir := t.TempDir()

	var out bytes.Buffer
	cfg := Config{
		RotateDir:      dir,
		RotatePrefix:   "capture",
		RotateInterval: time.Hour,
	}

	sinks, cleanup, err := buildSinks(cfg, &out)
	if err != nil {
		t.Fatalf("buildSinks: %v", err)
	}
	defer cleanup()

	if len(sinks) != 2 {
		t.Fatalf("expected 2 sinks (terminal + rotating), got %d", len(sinks))
	}
}

func TestBuildSinks_OutFileAndRotate(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "single.pcap")

	var out bytes.Buffer
	cfg := Config{
		OutFile:        outFile,
		RotateDir:      dir,
		RotatePrefix:   "capture",
		RotateInterval: time.Hour,
	}

	sinks, cleanup, err := buildSinks(cfg, &out)
	if err != nil {
		t.Fatalf("buildSinks: %v", err)
	}
	defer cleanup()

	if len(sinks) != 3 {
		t.Fatalf("expected 3 sinks (terminal + writer + rotating), got %d", len(sinks))
	}
}

func TestBuildSinks_CleanupClosesFile(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "test.pcap")

	var out bytes.Buffer
	cfg := Config{OutFile: outFile}

	_, cleanup, err := buildSinks(cfg, &out)
	if err != nil {
		t.Fatalf("buildSinks: %v", err)
	}

	// cleanup이 패닉 없이 실행되어야 한다
	cleanup()
}

func TestHandleStop_NoRunningDaemon(t *testing.T) {
	var errOut bytes.Buffer
	code := handleStop(&errOut)
	if code != 1 {
		t.Fatalf("expected exit code 1 when no daemon running, got %d", code)
	}
	if errOut.Len() == 0 {
		t.Fatal("expected error message in errOut")
	}
}

func TestBuildSinks_SinksImplementInterface(t *testing.T) {
	var out bytes.Buffer
	cfg := Config{}

	sinks, cleanup, err := buildSinks(cfg, &out)
	if err != nil {
		t.Fatalf("buildSinks: %v", err)
	}
	defer cleanup()

	for i, s := range sinks {
		if s == nil {
			t.Fatalf("sink[%d] is nil", i)
		}
		// capture.Sink 인터페이스를 구현하는지 컴파일 타임 검증
		var _ capture.Sink = s
	}
}
