# Go Conventions

이 문서는 `goscope` 저장소에서 따를 Go 코드 컨벤션을 정의한다. 기준은 Go 공식 문서와 Go Code Review Comments이며, 이 저장소의 현재 구조에 맞게 좁혀서 정리한다.

## 1. Project Layout

- 실행 엔트리포인트는 `cmd/<app>/main.go` 형태로 둔다.
- 재사용 가능한 내부 구현은 `internal/<domain>` 아래에 둔다.
- 외부 공개 의도가 없는 패키지는 `internal` 밖으로 빼지 않는다.
- 패키지는 책임 기준으로 나누고, `util`, `common`, `misc`, `types` 같은 의미가 약한 이름은 만들지 않는다.

현재 저장소 기준 예시:

```text
cmd/goscope/main.go
internal/capture/
internal/output/
internal/protocol/
```

## 2. Package Naming

- 패키지 이름은 짧고, 소문자이며, 의미가 분명해야 한다.
- 패키지 이름을 타입명에 반복하지 않는다.
- `capture.Engine`처럼 읽히도록 타입명을 정한다.
- 패키지명이 이미 맥락을 제공하면 `PacketCaptureEngine` 같은 중복 이름은 피한다.

권장:

```go
package capture

type Engine struct{}
```

비권장:

```go
package capture

type CaptureEngine struct{}
```

## 3. File and Declaration Style

- 파일은 `gofmt` 기준으로 유지한다.
- import 정렬은 `goimports` 기준을 따른다.
- exported 식별자에는 doc comment를 붙인다.
- 비공개 식별자라도 로직이 비자명하면 짧은 설명을 붙인다.
- 주석은 "무엇을 하는지"보다 "왜 이렇게 하는지"를 설명할 때만 쓴다.

## 4. Naming

- 로컬 변수명은 짧게 둔다. 범위가 좁으면 `i`, `pkt`, `cfg` 같은 이름을 쓴다.
- 범위가 넓거나 의미가 모호하면 더 설명적인 이름을 쓴다.
- Go 스타일의 mixedCaps를 사용한다.
- 약어는 Go 관례를 따른다: `ID`, `IP`, `TCP`, `UDP`, `URL`.

권장:

```go
srcIP := flow.Src()
packetCount := 0
```

비권장:

```go
srcIp := flow.Src()
packet_count := 0
```

## 5. Error Handling

- 에러는 무시하지 않는다.
- `_`로 버리는 대신 처리, 반환, 래핑 중 하나를 한다.
- 에러 메시지는 소문자로 시작하고 불필요한 마침표를 붙이지 않는다.
- 컨텍스트가 필요한 경우 `%w`로 래핑한다.

권장:

```go
handle, err := pcap.OpenLive(iface, snaplen, promisc, pcap.BlockForever)
if err != nil {
	return fmt.Errorf("open interface %q: %w", iface, err)
}
```

비권장:

```go
handle, _ := pcap.OpenLive(iface, snaplen, promisc, pcap.BlockForever)
```

## 6. Context and Concurrency

- 장시간 실행되는 캡처 루프는 `context.Context`를 받아야 한다.
- goroutine은 종료 경로가 명확해야 한다.
- 채널 수신 시 종료 가능성을 고려해 `ok`를 확인한다.
- 송신이 블로킹될 수 있는 채널에는 cancellation 경로를 함께 둔다.

권장:

```go
select {
case <-ctx.Done():
	return
case packets <- pkt:
}
```

## 7. CLI Rules

- CLI 플래그 파싱은 `main` 패키지에서만 처리한다.
- `internal` 패키지는 CLI 프레임워크에 의존하지 않는다.
- 기본 인터페이스 이름을 하드코딩하지 않는다.
- 사용자가 `-i`, `-f`, `-snaplen`, `-promisc` 등을 명시할 수 있게 한다.

## 8. Output Rules

- 패킷 출력은 가능한 한 정보 손실 없이 요약한다.
- 해석 불가능한 패킷도 버리지 말고 최소 메타데이터는 보여준다.
- 출력 포맷은 한 번 정하면 일관되게 유지한다.
- 사람이 읽는 출력과 기계가 읽는 출력(JSON 등)은 분리한다.

권장 최소 필드:

- timestamp
- protocol
- source
- destination
- length
- transport details

## 9. Testing

- 새 패키지에는 동일 디렉터리에 `_test.go` 파일을 둔다.
- 테스트는 table-driven 스타일을 기본으로 한다.
- 실패 메시지는 `got`과 `want`를 모두 포함한다.
- pcap 샘플 기반 테스트가 가능하면 golden input을 사용한다.

권장:

```go
if got != want {
	t.Fatalf("protocol = %v; want %v", got, want)
}
```

## 10. Dependencies

- 의존성은 꼭 필요한 경우에만 추가한다.
- 패킷 파싱은 우선 `gopacket`을 활용하고, 직접 구현은 필요성이 생길 때만 한다.
- CLI, 로깅, 설정 라이브러리는 표준 라이브러리로 부족할 때만 도입한다.

## 11. Formatting and Tooling

로컬 개발과 CI에서 아래 명령을 기준으로 맞춘다.

```bash
gofmt -w .
go test ./...
go vet ./...
```

`golangci-lint`를 추가할 경우, 규칙은 이 문서와 충돌하지 않게 최소한으로 구성한다.

## 12. Repository-Specific Decisions

- `cmd/main.go` 단일 파일보다 `cmd/goscope/main.go` 구조를 우선한다.
- `internal/types`는 범용 타입 보관소로 키우지 않는다.
- 프로토콜별 타입이나 동작은 가능한 한 해당 도메인 패키지로 이동한다.
- 운영체제별 기본값에 의존하는 코드는 옵션화하거나 런타임 탐지로 처리한다.

## References

- Go Code Review Comments: https://go.dev/wiki/CodeReviewComments
- Effective Go: https://go.dev/doc/effective_go
- Organizing a Go module: https://go.dev/doc/modules/layout
