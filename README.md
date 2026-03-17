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

```bash
git clone https://github.com/kwon93/goscope.git
cd goscope
go build -o goscope ./cmd/main.go
```

## 사용법

```
goscope [-i 인터페이스] [-f BPF필터] [-w 파일명.pcap]
```

| 플래그 | 설명 |
|--------|------|
| `-i` | 캡처할 네트워크 인터페이스 (생략 시 목록에서 선택) |
| `-f` | BPF 필터 표현식 |
| `-w` | 저장할 `.pcap` 파일명 |

```bash
# 인터페이스 선택 후 전체 캡처
sudo ./goscope

# HTTP 트래픽만 캡처
sudo ./goscope -i eth0 -f "tcp port 80"

# 파일로 저장
sudo ./goscope -i eth0 -f "tcp port 443" -w tls.pcap
```

> Linux / macOS는 `sudo` 또는 `CAP_NET_RAW` 권한이 필요합니다.

## 출력 예시

```
패킷 캡쳐 시작 (Ctrl+C를 눌러 종료하세요)
[14:32:01] TCP  192.168.0.5:54321 -> 443 → 93.184.216.34
[14:32:01] UDP  192.168.0.5:53    -> 53  → 8.8.8.8
```

## 문서

- [전체 아키텍처](docs/overall-architecture.md)
- [Go 코딩 컨벤션](docs/go-conventions.md)
- [클린 아키텍처 가이드](docs/go-cli-clean-architecture.md)

## 라이선스

MIT
