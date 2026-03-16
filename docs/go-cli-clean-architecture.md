# Go CLI Clean Architecture

이 문서는 `goscope` 같은 Go CLI 애플리케이션에서 따를 클린 아키텍처 기준을 정의한다. 목적은 다음 세 가지다.

- CLI 프레임워크와 핵심 로직을 분리한다.
- 패킷 캡처 같은 외부 의존성을 안쪽 계층으로 침투시키지 않는다.
- 기능이 늘어나도 파일이 아니라 계층과 책임 기준으로 확장한다.

## 1. Core Rule

의존성은 항상 바깥에서 안쪽으로만 향한다.

```text
cmd -> app -> domain <- adapter <- infrastructure
```

의미:

- `cmd`: 플래그 파싱, 프로세스 종료 코드, 시그널 처리
- `app`: 유스케이스 조합, 실행 흐름 orchestration
- `domain`: 순수 비즈니스 규칙, 엔티티, 정책, 포트
- `adapter`: 외부 입출력을 도메인 포트에 맞게 변환
- `infrastructure`: `pcap`, 파일 시스템, stdout, OS signal 같은 실제 구현

`domain`은 어떤 외부 라이브러리도 몰라야 한다.

## 2. Recommended Layout

`goscope` 기준 권장 구조:

```text
cmd/goscope/main.go

internal/app/capture_packets.go
internal/app/list_interfaces.go

internal/domain/packet/packet.go
internal/domain/packet/protocol.go
internal/domain/packet/port.go

internal/domain/capture/service.go
internal/domain/capture/ports.go

internal/adapter/cli/flags.go
internal/adapter/cli/run.go
internal/adapter/presenter/terminal.go

internal/infrastructure/pcap/engine.go
internal/infrastructure/netif/list.go
internal/infrastructure/signal/shutdown.go
```

## 3. Layer Responsibilities

### 3.1 `cmd`

- 프로그램 시작점만 둔다.
- 플래그를 파싱한다.
- 종료 코드를 결정한다.
- 로깅 설정, signal wiring, context 생성까지 처리한다.
- 핵심 로직은 직접 구현하지 않는다.

`main.go`는 얇아야 한다.

권장:

```go
func main() {
	os.Exit(run())
}
```

### 3.2 `app`

- 한 개의 사용자 행동을 한 개의 유스케이스로 모델링한다.
- 여러 포트를 조합해 실행 순서를 정의한다.
- 도메인 규칙은 호출하지만, 세부 인프라 구현은 모른다.
- 입력 DTO와 출력 DTO를 명시한다.

예시 유스케이스:

- `CapturePackets`
- `ListInterfaces`
- `ReadPcapFile`
- `WritePcapFile`

### 3.3 `domain`

- 패킷 분석 도구의 핵심 규칙을 담는다.
- 포트 인터페이스를 정의한다.
- 외부 라이브러리 타입을 직접 노출하지 않는다.

예시:

```go
type Packet struct {
	Timestamp time.Time
	Protocol  Protocol
	SrcAddr   string
	DstAddr   string
	Length    int
}
```

도메인에서는 `gopacket.Packet`을 직접 쓰지 않는다.

### 3.4 `adapter`

- CLI 입력을 유스케이스 입력으로 변환한다.
- 유스케이스 출력을 사람이 읽을 문자열이나 JSON으로 변환한다.
- 도메인 포트와 인프라 구현 사이의 번역 계층 역할을 한다.

예시:

- `adapter/cli`: flags -> use case input
- `adapter/presenter`: domain packet -> terminal line

### 3.5 `infrastructure`

- 실제 외부 세계와 통신한다.
- `pcap.OpenLive`, 파일 읽기/쓰기, OS 인터페이스 탐색, stdout/stderr 접근이 여기에 속한다.
- 이 계층은 구체 구현을 제공하고, 상위 계층의 포트를 만족한다.

## 4. Dependency Boundaries

다음 규칙을 지킨다.

- `internal/domain/...`은 `github.com/google/gopacket`를 import하지 않는다.
- `internal/app/...`은 `pcap`를 직접 import하지 않는다.
- `internal/infrastructure/...`은 CLI 플래그 파싱 패키지를 import하지 않는다.
- `internal/adapter/...`는 도메인 타입을 포맷팅할 수 있지만, 인프라 세부 구현을 감추는 역할까지만 한다.

비권장 예시:

```go
package app

import "github.com/google/gopacket/pcap"
```

이건 유스케이스가 인프라 세부사항에 묶이므로 피한다.

## 5. Port Design

포트는 도메인이나 앱 계층에서 정의한다. 구현은 인프라에서 제공한다.

예시:

```go
type PacketSource interface {
	Capture(ctx context.Context, req CaptureRequest) (<-chan Packet, error)
}

type PacketSink interface {
	WritePacket(ctx context.Context, pkt Packet) error
}
```

장점:

- 테스트에서 fake 구현을 쉽게 넣을 수 있다.
- `pcap` live capture와 pcap file replay를 같은 포트로 다룰 수 있다.

## 6. Use Case Shape

유스케이스는 입력과 출력을 명확히 갖는다.

권장:

```go
type CapturePacketsInput struct {
	Interface string
	Filter    string
	Snaplen   int32
	Promisc   bool
}

type CapturePackets struct {
	Source    PacketSource
	Presenter PacketSink
}

func (uc CapturePackets) Run(ctx context.Context, in CapturePacketsInput) error
```

규칙:

- 유스케이스는 전역 상태를 직접 읽지 않는다.
- `os.Args`, 환경변수, stdout은 `cmd` 또는 adapter에서 처리한다.
- 유스케이스 메서드는 하나의 분명한 동작을 수행한다.

## 7. Error Strategy

- 도메인 에러는 의미 있는 sentinel 또는 typed error로 정의할 수 있다.
- 인프라 에러는 유스케이스 경계에서 문맥을 덧붙여 래핑한다.
- CLI 계층은 에러를 사용자 메시지와 exit code로 변환한다.

예시 흐름:

```text
pcap open error
-> infrastructure wraps low-level details
-> app adds use-case context
-> cmd prints concise message and exits non-zero
```

## 8. Testing Strategy by Layer

### `domain`

- 순수 단위 테스트
- 외부 의존성 없음

### `app`

- fake 포트 사용
- 유스케이스 흐름 검증

### `adapter`

- 포맷 검증
- CLI 입력 매핑 검증

### `infrastructure`

- 통합 테스트 위주
- 필요 시 pcap fixture 사용

원칙:

- 대부분의 테스트는 `domain`과 `app`에 있어야 한다.
- 실제 네트워크 인터페이스에 붙는 테스트는 최소화한다.

## 9. What This Means for Current `goscope`

현재 코드 기준으로 보면:

- `internal/capture`는 인프라 계층에 가깝다.
- `internal/output`은 presenter adapter에 가깝다.
- `internal/types`는 책임이 애매하므로 `domain/packet` 쪽으로 흡수하는 편이 낫다.
- `cmd/main.go`는 향후 `cmd/goscope/main.go`로 이동하는 편이 구조상 낫다.

추천 리팩터링 방향:

1. `internal/types/protocol.go`를 `internal/domain/packet/protocol.go`로 이동
2. `internal/capture/engine.go`를 `internal/infrastructure/pcap/engine.go`로 이동
3. `internal/output/terminal.go`를 `internal/adapter/presenter/terminal.go`로 이동
4. `internal/app/capture_packets.go`를 추가해 흐름 orchestration 분리
5. `cmd/goscope/main.go`에서 flags와 signal만 처리

## 10. Non-Goals

다음은 지금 단계에서 하지 않는다.

- DDD 패턴을 과도하게 도입해 복잡한 aggregate를 만들지 않는다.
- 인터페이스를 모든 타입마다 만들지 않는다.
- 작은 CLI에 웹 서버 수준의 계층 수를 억지로 만들지 않는다.

클린 아키텍처의 목적은 추상화 자체가 아니라 변경 비용 제어다.

## 11. Decision Rules

새 코드를 추가할 때 아래 질문으로 위치를 결정한다.

1. 이 코드는 비즈니스 규칙인가?
2. 이 코드는 실행 흐름 조합인가?
3. 이 코드는 외부 입출력 변환인가?
4. 이 코드는 라이브러리나 OS에 직접 붙는가?

판단 기준:

- 1이면 `domain`
- 2이면 `app`
- 3이면 `adapter`
- 4이면 `infrastructure`

## References

- Clean Architecture, Robert C. Martin
- Go Code Review Comments: https://go.dev/wiki/CodeReviewComments
- Organizing a Go module: https://go.dev/doc/modules/layout
