package cli

import (
	"bytes"
	"strings"
	"testing"
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
