# Go CLI Architecture

이 문서는 `goscope` 같은 Go CLI 애플리케이션의 구조 기준을 정의한다.

Go다운 구조의 핵심은 **패키지를 타입이 아니라 기능 단위로 나누는 것**이다.

## 1. 기본 원칙

### 패키지는 역할로 나눈다

Java처럼 레이어(domain/application/infrastructure)로 나누지 않는다.
Go에서는 **무엇을 하는가** 기준으로 패키지를 구성한다.

```text
Bad  : internal/domain/packet, internal/app, internal/adapter, internal/infrastructure
Good : internal/capture, internal/netif, internal/cli
```

### 인터페이스는 사용하는 쪽에 정의한다

Go의 인터페이스는 구현체가 아니라 **호출자** 쪽에서 선언한다.

```go
// capture 패키지가 자신에게 필요한 것을 직접 선언
package capture

type Source interface {
    Capture(ctx context.Context, req Request) (<-chan Packet, error)
}

type Sink interface {
    WritePacket(ctx context.Context, pkt Packet) error
}
```

구현체는 어디에 있든 이 인터페이스를 암묵적으로 만족하면 된다.

### 인터페이스는 작게 유지한다

Go 표준 라이브러리가 모범이다.

```go
io.Reader  → Read(p []byte) (n int, err error)    // 메서드 1개
io.Writer  → Write(p []byte) (n int, err error)   // 메서드 1개
```

인터페이스가 클수록 구현하기 어렵고, 테스트 fake 만들기도 어렵다.

### Accept interfaces, return structs

함수 인자는 인터페이스로 받고, 반환 타입은 구체 타입(struct)으로 돌려준다.

```go
// Good
func NewTerminal(w io.Writer) *Terminal { ... }

// Bad
func NewTerminal(w io.Writer) Sink { ... }
```

## 2. 권장 레이아웃

```text
cmd/goscope/
  main.go          # 진입점: signal 등록, cli.Run() 호출만

internal/
  cli/
    flags.go       # 플래그 파싱, 인터랙티브 입력
    runner.go      # 의존성 조립, 실행
  capture/
    capture.go     # Packet 타입, Source/Sink 인터페이스, 핵심 로직
    engine.go      # pcap 엔진 (Source 구현)
    writer.go      # .pcap 파일 저장 (Sink 구현)
    terminal.go    # 터미널 출력 (Sink 구현)
  netif/
    netif.go       # 네트워크 인터페이스 목록
```

### 왜 이 구조인가?

- `capture` 패키지 하나가 타입, 인터페이스, 구현체를 모두 소유한다
- 패키지 수가 적어 탐색이 쉽다
- `import cycle` 없이 자연스럽게 의존 관계가 정리된다

## 3. `cmd/main.go`

얇게 유지한다. signal 처리와 종료 코드만 담는다.

```go
func main() {
    os.Exit(run())
}

func run() int {
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
    defer cancel()
    return cli.Run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
}
```

`os.Exit(run())` 패턴: `defer`가 `os.Exit` 직전에 실행되도록 `run()`으로 감싼다.

## 4. `internal/capture`

이 프로젝트의 핵심 패키지. 타입과 인터페이스, 비즈니스 로직이 함께 있다.

```go
// capture.go — 타입과 인터페이스
type Packet struct { ... }
type Request struct { ... }

type Source interface {
    Capture(ctx context.Context, req Request) (<-chan Packet, error)
}
type Sink interface {
    WritePacket(ctx context.Context, pkt Packet) error
}

func Run(ctx context.Context, src Source, sinks []Sink, req Request) error { ... }

// engine.go — Source 구현
type Engine struct{}
func (e *Engine) Capture(...) (<-chan Packet, error) { ... }

// terminal.go — Sink 구현
type Terminal struct{ out io.Writer }
func (t *Terminal) WritePacket(...) error { ... }

// writer.go — Sink 구현
type Writer struct{ w *pcapgo.Writer }
func (w *Writer) WritePacket(...) error { ... }
```

## 5. 에러 처리

- 에러는 `fmt.Errorf("문맥: %w", err)` 로 래핑해 위로 전달한다
- 최종적으로 `cmd` 계층에서 사용자 메시지와 exit code로 변환한다
- 패닉은 쓰지 않는다

```go
// infrastructure
f, err := pcap.OpenLive(...)
if err != nil {
    return fmt.Errorf("pcap 열기 실패: %w", err)
}

// cmd
if err := run(...); err != nil {
    fmt.Fprintln(os.Stderr, err)
    return 1
}
```

## 6. 테스트 전략

인터페이스가 작으면 fake 구현이 단순해진다.

```go
type fakeSource struct {
    packets []capture.Packet
}

func (f *fakeSource) Capture(_ context.Context, _ capture.Request) (<-chan capture.Packet, error) {
    ch := make(chan capture.Packet, len(f.packets))
    for _, p := range f.packets {
        ch <- p
    }
    close(ch)
    return ch, nil
}
```

- `capture` 패키지: fake Source/Sink로 핵심 로직 단위 테스트
- `cli` 패키지: `bytes.Buffer`로 입출력 검증
- `netif` 패키지: 실제 OS 호출 → 통합 테스트로 최소화

## 7. 하지 않는 것

- 모든 struct에 인터페이스를 만들지 않는다
- 레이어 이름(domain, application, infrastructure)으로 패키지를 만들지 않는다
- 변경될 가능성 없는 코드를 위한 추상화를 미리 만들지 않는다

> "A little copying is better than a little dependency." — Go Proverbs

## References

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [Go Proverbs](https://go-proverbs.github.io/)
- [Organizing a Go module](https://go.dev/doc/modules/layout)
