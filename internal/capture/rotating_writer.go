package capture

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// RotatingWriter는 지정된 시간 간격마다 새 pcap 파일을 생성하는 Sink 구현체다.
type RotatingWriter struct {
	dir      string
	prefix   string
	interval time.Duration
	now      func() time.Time

	mu       sync.Mutex
	current  *os.File
	writer   *Writer
	rotateAt time.Time
}

// NewRotatingWriter는 dir 디렉터리에 prefix_<timestamp>.pcap 파일을 interval마다 교체한다.
func NewRotatingWriter(dir, prefix string, interval time.Duration) (*RotatingWriter, error) {
	return &RotatingWriter{
		dir:      dir,
		prefix:   prefix,
		interval: interval,
		now:      time.Now,
	}, nil
}

// WritePacket은 현재 시각이 rotateAt을 넘으면 새 파일을 열고, pcap 레코드를 기록한다.
func (r *RotatingWriter) WritePacket(ctx context.Context, pkt Packet) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.writer == nil || r.now().After(r.rotateAt) {
		if err := r.rotate(); err != nil {
			return err
		}
	}
	return r.writer.WritePacket(ctx, pkt)
}

// Close는 현재 열린 파일을 닫는다.
func (r *RotatingWriter) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.current != nil {
		return r.current.Close()
	}
	return nil
}

// rotate는 현재 파일을 닫고 새 파일을 연다. mu를 이미 보유한 상태에서 호출한다.
func (r *RotatingWriter) rotate() error {
	if r.current != nil {
		if err := r.current.Close(); err != nil {
			return fmt.Errorf("rotate: close previous file: %w", err)
		}
	}

	if err := os.MkdirAll(r.dir, 0700); err != nil {
		return fmt.Errorf("rotate: mkdir %q: %w", r.dir, err)
	}

	now := r.now()
	boundary := now.Truncate(r.interval)
	name := fmt.Sprintf("%s_%s.pcap", r.prefix, boundary.Format("2006-01-02T15-04-05"))
	path := filepath.Join(r.dir, name)

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("rotate: create %q: %w", path, err)
	}

	w, err := NewWriter(f)
	if err != nil {
		f.Close()
		return fmt.Errorf("rotate: init writer: %w", err)
	}

	r.current = f
	r.writer = w
	r.rotateAt = boundary.Add(r.interval)
	return nil
}
