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

### 입력 흐름

```
저장할 파일명을 입력하세요 (생략시 Enter): capture
네트워크 인터페이스를 선택하세요:
  [0] Realtek PCIe GbE Family Controller (eth0)
  [1] Qualcomm WiFi Adapter (wlan0)
  ...
선택 (0~N, 기본값 0): 1
BPF 필터를 입력하세요 (ex: tcp port 80, 생략시 Enter): tcp port 443
패킷 캡처 시작 (Ctrl+C를 눌러 종료하세요)
```

| 항목 | 설명 |
|------|------|
| 파일명 | 저장할 `.pcap` 파일명. `.pcap` 확장자는 자동으로 붙습니다. 생략하면 파일 저장 없이 터미널에만 출력합니다. |
| 인터페이스 | 번호로 선택. Enter 입력 시 0번이 기본값입니다. |
| BPF 필터 | tcpdump 문법과 동일합니다. 잘못된 표현식은 즉시 안내 후 재입력을 요청합니다. |

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

```
패킷 캡처 시작 (Ctrl+C를 눌러 종료하세요)
[14:32:01] tcp  192.168.0.5:54321 → 443 → 93.184.216.34
[14:32:01] udp  192.168.0.5:12345 → 53  → 8.8.8.8
[14:32:02] tcp  192.168.0.5:54322 → 80  → 203.0.113.5
```

## 문서

- [전체 아키텍처](docs/overall-architecture.md)
- [Go 코딩 컨벤션](docs/go-conventions.md)
- [클린 아키텍처 가이드](docs/go-cli-clean-architecture.md)

## 라이선스

MIT
