# goscope

Go로 개발한 CLI 기반 네트워크 패킷 분석 도구.
실시간 캡처, BPF 필터링, pcap 파일 저장을 지원합니다.

## 요구 사항

| OS | 필요 라이브러리 |
|----|----------------|
| Linux | `sudo apt install libpcap-dev` |
| macOS | `brew install libpcap` |
| Windows | [Npcap](https://npcap.com/) 설치 |

## 설치

**바이너리 다운로드 (권장)**

[GitHub Releases](https://github.com/kwon93/goscope/releases/latest)에서 OS에 맞는 파일을 내려받으세요.

| OS | 파일명 |
|----|--------|
| Linux | `goscope-linux-amd64` |
| macOS (Apple Silicon) | `goscope-darwin-arm64` |
| Windows | `goscope-windows-amd64.exe` |

> **macOS 주의사항**
>
> 1. **"신뢰할 수 없는 개발자" 경고가 뜨는 경우** — 서명되지 않은 바이너리는 macOS Gatekeeper가 차단합니다. 먼저 격리 속성을 제거하세요.
>
>    ```bash
>    xattr -dr com.apple.quarantine ./goscope-darwin-arm64
>    ```
>
> 2. **패킷 캡처 권한 오류 (`Permission denied`)** — macOS는 네트워크 캡처 시 루트 권한이 필요합니다. 반드시 `sudo`로 실행하세요.
>
>    ```bash
>    sudo ./goscope-darwin-arm64
>    ```

**직접 빌드**

```bash
git clone https://github.com/kwon93/goscope.git
cd goscope
go build -o goscope ./cmd/goscope/
```

## 사용법

실행하면 순서대로 질문합니다. 모든 항목은 Enter로 생략할 수 있습니다.

```bash
sudo ./goscope          # Linux / macOS
./goscope.exe           # Windows (관리자 권한 불필요, Npcap 설치 전제)
```

> Linux / macOS는 `sudo` 또는 `CAP_NET_RAW` 권한이 필요합니다.

### 백그라운드 중지

```bash
./goscope --stop        # 실행 중인 백그라운드 캡처 종료
```

### 입력 흐름

```
저장 방식을 선택하세요:
  [1] 저장 안 함 (터미널 출력만)
  [2] 단일 파일로 저장
  [3] 시간대별 파일 분할 저장
선택 (기본값 1): 3

저장 디렉터리 (기본값 현재 디렉터리 .): /var/log/pcap
파일 분할 주기 (예: 1h, 30m, 기본값 1h): 30m

네트워크 인터페이스를 선택하세요:
  [0] Realtek PCIe GbE Family Controller (eth0)
  [1] Qualcomm WiFi Adapter (wlan0)
  ...
선택 (0~N, 기본값 0): 0

BPF 필터를 입력하세요 (ex: tcp port 80, 생략시 Enter): tcp port 443

타이머를 설정하시겠습니까? [y/N]: y
시작 시각 (예: 14:00, 생략 시 즉시 시작): 14:00
종료 시각 (예: 15:00): 15:00

백그라운드로 실행하시겠습니까? (터미널을 닫아도 계속 실행됩니다) [y/N]: y
백그라운드 캡처 시작. PID: 12345
중지하려면: goscope --stop
```

| 항목 | 설명 |
|------|------|
| 저장 방식 | `1` 터미널 출력만 / `2` 단일 `.pcap` 파일 / `3` 시간 주기로 파일 자동 분할 |
| 저장 디렉터리 | 분할 저장 시 파일을 쌓을 디렉터리. 생략하면 현재 디렉터리(`.`) |
| 분할 주기 | `30m`, `1h`, `6h` 형식. 최소 1분. 생략하면 1시간 |
| 인터페이스 | 번호로 선택. Enter 입력 시 0번이 기본값입니다. |
| BPF 필터 | tcpdump 문법과 동일합니다. 잘못된 표현식은 즉시 안내 후 재입력을 요청합니다. |
| 타이머 | `N` 또는 Enter로 생략 가능. 시작 시각은 선택, 종료 시각은 필수. `HH:MM` 형식. |
| 타이머 — 시작 시각 | 생략하면 즉시 시작. 현재 시각보다 과거 시각을 입력하면 다음 날로 자동 조정됩니다. |
| 타이머 — 종료 시각 | 지정한 시각에 캡처가 자동 종료됩니다. 종료 시각이 시작보다 이르면(자정 경계) 다음 날로 자동 조정됩니다. |
| 백그라운드 | 분할 저장 선택 시에만 나타납니다. `y` 입력 시 터미널을 닫아도 캡처가 유지됩니다. |

### BPF 필터 예시

| 목적 | 표현식 |
|------|--------|
| HTTP 트래픽 | `tcp port 80` |
| HTTPS 트래픽 | `tcp port 443` |
| DNS 조회 | `udp port 53` |
| 특정 호스트 | `host 192.168.1.1` |
| 특정 호스트의 HTTP | `host 192.168.1.1 and tcp port 80` |
| 특정 포트 제외 | `not port 22` |

## 출력 예시

**일반 실행**

```
패킷 캡처 시작 (Ctrl+C를 눌러 종료하세요)
[14:32:01] tcp  192.168.0.5:54321 → 93.184.216.34:443
[14:32:01] udp  192.168.0.5:12345 → 8.8.8.8:53
[14:32:02] tcp  192.168.0.5:54322 → 203.0.113.5:80
```

**타이머 실행** (시작 대기 → 자동 종료)

```
캡처 시작까지 대기 중... (14:00 시작)
패킷 캡처 시작 (종료 시각: 15:00)
[14:00:03] tcp  192.168.0.5:54321 → 93.184.216.34:443
...
```

## 문서

- [전체 아키텍처](docs/overall-architecture.md)
- [Go 코딩 컨벤션](docs/go-conventions.md)
- [클린 아키텍처 가이드](docs/go-cli-clean-architecture.md)

## 라이선스

MIT
