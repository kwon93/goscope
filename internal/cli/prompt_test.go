package cli

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
	"time"
)

// --- promptOutFile ---

func TestPromptOutFile_Empty(t *testing.T) {
	in := strings.NewReader("\n")
	out := &bytes.Buffer{}
	name, err := promptOutFile(in, out)
	if err != nil {
		t.Fatalf("promptOutFile: %v", err)
	}
	if name != "" {
		t.Fatalf("got %q; want empty string", name)
	}
}

func TestPromptOutFile_WithoutExtension(t *testing.T) {
	in := strings.NewReader("capture\n")
	out := &bytes.Buffer{}
	name, err := promptOutFile(in, out)
	if err != nil {
		t.Fatalf("promptOutFile: %v", err)
	}
	if name != "capture.pcap" {
		t.Fatalf("got %q; want capture.pcap", name)
	}
}

func TestPromptOutFile_WithExtension(t *testing.T) {
	in := strings.NewReader("capture.pcap\n")
	out := &bytes.Buffer{}
	name, err := promptOutFile(in, out)
	if err != nil {
		t.Fatalf("promptOutFile: %v", err)
	}
	if name != "capture.pcap" {
		t.Fatalf("got %q; want capture.pcap", name)
	}
}

func TestPromptOutFile_NullByte(t *testing.T) {
	in := strings.NewReader("cap\x00ture\n")
	out := &bytes.Buffer{}
	_, err := promptOutFile(in, out)
	if err == nil {
		t.Fatal("got nil error; want error for null byte in filename")
	}
}

// --- promptFilter ---

func TestPromptFilter_Empty(t *testing.T) {
	in := strings.NewReader("\n")
	out := &bytes.Buffer{}
	filter, err := promptFilter(in, out)
	if err != nil {
		t.Fatalf("promptFilter: %v", err)
	}
	if filter != "" {
		t.Fatalf("got %q; want empty string", filter)
	}
}

func TestPromptFilter_Valid(t *testing.T) {
	in := strings.NewReader("tcp port 80\n")
	out := &bytes.Buffer{}
	filter, err := promptFilter(in, out)
	if err != nil {
		t.Fatalf("promptFilter: %v", err)
	}
	if filter != "tcp port 80" {
		t.Fatalf("got %q; want tcp port 80", filter)
	}
}

func TestPromptFilter_InvalidThenValid(t *testing.T) {
	// 첫 줄: 잘못된 BPF → 재입력 안내 후 두 번째 줄 수락
	in := strings.NewReader("tcp 80\ntcp port 80\n")
	out := &bytes.Buffer{}
	filter, err := promptFilter(in, out)
	if err != nil {
		t.Fatalf("promptFilter: %v", err)
	}
	if filter != "tcp port 80" {
		t.Fatalf("got %q; want tcp port 80", filter)
	}
	if !strings.Contains(out.String(), "잘못된 BPF 필터") {
		t.Fatal("expected error message in output; got none")
	}
}

func TestPromptFilter_MultipleInvalidThenValid(t *testing.T) {
	in := strings.NewReader("tcp 80\nport\nudp port 53\n")
	out := &bytes.Buffer{}
	filter, err := promptFilter(in, out)
	if err != nil {
		t.Fatalf("promptFilter: %v", err)
	}
	if filter != "udp port 53" {
		t.Fatalf("got %q; want udp port 53", filter)
	}
}

// --- promptTimer ---

func TestPromptTimer_Skip(t *testing.T) {
	in := strings.NewReader("N\n")
	out := &bytes.Buffer{}
	start, end, err := promptTimer(in, out)
	if err != nil {
		t.Fatalf("promptTimer: %v", err)
	}
	if !start.IsZero() {
		t.Fatalf("expected zero start time, got %v", start)
	}
	if !end.IsZero() {
		t.Fatalf("expected zero end time, got %v", end)
	}
}

func TestPromptTimer_EmptyInputSkips(t *testing.T) {
	// 빈 입력(Enter)도 N과 동일하게 타이머 미설정
	in := strings.NewReader("\n")
	out := &bytes.Buffer{}
	start, end, err := promptTimer(in, out)
	if err != nil {
		t.Fatalf("promptTimer: %v", err)
	}
	if !start.IsZero() || !end.IsZero() {
		t.Fatalf("expected zero times, got start=%v end=%v", start, end)
	}
}

func TestPromptTimer_StartOptional(t *testing.T) {
	// 시작 시각 생략(optional=true) + 종료 시각만 설정
	in := strings.NewReader("y\n\n23:00\n")
	out := &bytes.Buffer{}
	start, end, err := promptTimer(in, out)
	if err != nil {
		t.Fatalf("promptTimer: %v", err)
	}
	if !start.IsZero() {
		t.Fatalf("expected zero start time when skipped, got %v", start)
	}
	if end.IsZero() {
		t.Fatal("expected non-zero end time")
	}
	if end.Hour() != 23 || end.Minute() != 0 {
		t.Fatalf("expected end 23:00, got %02d:%02d", end.Hour(), end.Minute())
	}
}

func TestPromptTimer_BothTimes(t *testing.T) {
	// 시작 < 종료: 둘 다 정상 설정
	in := strings.NewReader("y\n14:00\n15:00\n")
	out := &bytes.Buffer{}
	start, end, err := promptTimer(in, out)
	if err != nil {
		t.Fatalf("promptTimer: %v", err)
	}
	if start.IsZero() || end.IsZero() {
		t.Fatalf("expected both times set, got start=%v end=%v", start, end)
	}
	if start.Hour() != 14 || start.Minute() != 0 {
		t.Fatalf("expected start 14:00, got %02d:%02d", start.Hour(), start.Minute())
	}
	if end.Hour() != 15 || end.Minute() != 0 {
		t.Fatalf("expected end 15:00, got %02d:%02d", end.Hour(), end.Minute())
	}
	if !end.After(start) {
		t.Fatalf("expected end after start, got start=%v end=%v", start, end)
	}
}

func TestPromptTimer_MidnightCrossing(t *testing.T) {
	// 종료 < 시작(일자 경계): end에 +24h 조정 후 end.After(start) 보장
	in := strings.NewReader("y\n23:00\n01:00\n")
	out := &bytes.Buffer{}
	start, end, err := promptTimer(in, out)
	if err != nil {
		t.Fatalf("promptTimer: %v", err)
	}
	if !end.After(start) {
		t.Fatalf("expected end after start for midnight crossing, got start=%v end=%v", start, end)
	}
	// 두 시각의 차이가 2시간(01:00~23:00 역방향 = 2h 순방향)
	diff := end.Sub(start)
	if diff != 2*time.Hour {
		t.Fatalf("expected 2h gap for 23:00→01:00, got %v", diff)
	}
}

// --- promptTimeInput ---

func TestPromptTimeInput_InvalidFormatThenValid(t *testing.T) {
	in := bufio.NewReader(strings.NewReader("badtime\n14:30\n"))
	out := &bytes.Buffer{}
	result, err := promptTimeInput(in, out, "시각: ", false)
	if err != nil {
		t.Fatalf("promptTimeInput: %v", err)
	}
	if result.Hour() != 14 || result.Minute() != 30 {
		t.Fatalf("expected 14:30, got %02d:%02d", result.Hour(), result.Minute())
	}
	if !strings.Contains(out.String(), "HH:MM") {
		t.Fatal("expected retry message in output")
	}
}

func TestPromptTimeInput_PastTime_NextDay(t *testing.T) {
	// 00:00은 자정으로 현재 시각보다 항상 과거 → 내일 00:00으로 조정
	in := bufio.NewReader(strings.NewReader("00:00\n"))
	out := &bytes.Buffer{}
	before := time.Now()
	result, err := promptTimeInput(in, out, "시각: ", false)
	if err != nil {
		t.Fatalf("promptTimeInput: %v", err)
	}
	if !result.After(before) {
		t.Fatalf("expected past time to be adjusted to future, got %v", result)
	}
	if result.Hour() != 0 || result.Minute() != 0 {
		t.Fatalf("expected 00:00, got %02d:%02d", result.Hour(), result.Minute())
	}
}

func TestPromptTimeInput_Optional_EmptyReturnsZero(t *testing.T) {
	in := bufio.NewReader(strings.NewReader("\n"))
	out := &bytes.Buffer{}
	result, err := promptTimeInput(in, out, "시각: ", true)
	if err != nil {
		t.Fatalf("promptTimeInput: %v", err)
	}
	if !result.IsZero() {
		t.Fatalf("expected zero time for optional empty input, got %v", result)
	}
}

func TestPromptTimeInput_NonOptional_EmptyRetries(t *testing.T) {
	// optional=false일 때 빈 입력은 재시도, 두 번째 유효 입력 수락
	in := bufio.NewReader(strings.NewReader("\n10:00\n"))
	out := &bytes.Buffer{}
	result, err := promptTimeInput(in, out, "시각: ", false)
	if err != nil {
		t.Fatalf("promptTimeInput: %v", err)
	}
	if result.Hour() != 10 || result.Minute() != 0 {
		t.Fatalf("expected 10:00, got %02d:%02d", result.Hour(), result.Minute())
	}
}
