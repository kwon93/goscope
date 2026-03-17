# Overall Architecture

## 레이어 구조

```mermaid
graph TD
    subgraph cmd["cmd/ — Entry Point"]
        MAIN["main.go\n플래그 파싱 · 인터페이스 선택\n시그널 핸들링"]
    end

    subgraph internal["internal/"]
        subgraph capture["capture/ — Infrastructure"]
            ENGINE["engine.go\npcap.OpenLive()\ngoroutine + channel"]
            OPTIONS["options.go\n함수형 옵션 패턴\nWithInterface · WithFilter ..."]
            ENGINE --> OPTIONS
        end

        subgraph output["output/ — Presenter"]
            TERMINAL["terminal.go\n패킷 포맷 출력"]
        end

        subgraph types["types/ — Domain"]
            PROTO["protocol.go\nTCP · UDP · OTHER"]
        end
    end

    MAIN --> ENGINE
    MAIN --> TERMINAL
    TERMINAL --> PROTO
```

## 실행 흐름 시퀀스

```mermaid
sequenceDiagram
    actor User
    participant main as main
    participant engine as capture.Engine
    participant output as output.Print

    User->>main: 실행 (인터페이스 · 필터 선택)
    main->>engine: Start(ctx)
    engine-->>main: packets channel

    loop 패킷 수신
        engine->>main: packet
        main->>output: Print(packet)
        output-->>User: [HH:MM:SS] proto src → dst
    end

    User->>main: Ctrl+C
    main->>engine: ctx.Done()
    engine-->>main: channel 종료
```

## 프로젝트 구조

```
goscope/
├── cmd/
│   └── main.go                  # 진입점: 플래그 파싱, 인터페이스 선택, 시그널 처리
├── internal/
│   ├── capture/
│   │   ├── engine.go            # pcap 캡처 엔진 (고루틴 기반 채널 스트림)
│   │   └── options.go           # 함수형 옵션 패턴 (WithInterface, WithFilter)
│   ├── output/
│   │   └── terminal.go          # 패킷 포맷 출력
│   └── types/
│       └── protocol.go          # 프로토콜 타입 정의 (TCP, UDP, OTHER)
└── docs/
    ├── overall-architecture.md       # 전체 아키텍처 (현재 문서)
    ├── go-conventions.md             # Go 코딩 컨벤션
    └── go-cli-clean-architecture.md  # 클린 아키텍처 가이드
```

## 의존성

| 패키지 | 역할 |
|--------|------|
| `github.com/google/gopacket` | 패킷 파싱 및 pcap 핸들링 |
