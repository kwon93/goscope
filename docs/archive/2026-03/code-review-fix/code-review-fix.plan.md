# Plan: code-review-fix

## Executive Summary

| Perspective | Content |
|-------------|---------|
| **Problem** | 코드 리뷰에서 발견된 24개 이슈(HIGH 7, MEDIUM 9, LOW 8) — 보안 취약점, 코드 품질 버그, 테스트 공백이 프로덕션 신뢰도를 저하시킴 |
| **Solution** | 심각도 순으로 이슈를 그룹화하여 단계적으로 수정 — HIGH는 즉시, MEDIUM은 단기, LOW는 개선 사이클에서 처리 |
| **Function/UX Effect** | 보안 공격 벡터 제거, 데이터 손실 버그 수정, 터미널 출력 정확성 개선, CI에 `-race` 및 커버리지 측정 추가 |
| **Core Value** | 안전하고 신뢰할 수 있는 패킷 캡처 도구로 운영 리스크 최소화 |

---

## 1. 배경 및 목표

### 1.1 배경
2026-03-20 멀티 에이전트 코드 리뷰(security-architect, code-analyzer, qa-strategist)를 통해
goscope 전체 코드베이스에서 총 24개 이슈를 식별함.

### 1.2 목표
- HIGH 이슈 7건 전량 수정
- 핵심 MEDIUM 이슈 수정 (파일 권한, 코드 중복)
- CI 파이프라인 품질 강화 (`-race`, 커버리지)
- `runner.go` 테스트 추가

---

## 2. 이슈 목록 및 우선순위

### Phase 1 — 즉시 수정 (HIGH, ~1시간)

| # | 이슈 | 파일 | 수정 방법 |
|---|------|------|----------|
| H-1 | PID 파일 symlink/TOCTOU 공격 | `daemon/daemon.go:53-90` | `O_EXCL` 플래그 추가, PID 파일을 사용자 전용 디렉터리로 이동 |
| H-2 | 환경변수 설정 JSON 노출 | `daemon/daemon.go:14` | 임시파일(0600)로 설정 전달 후 즉시 삭제 |
| H-3 | 파일 경로 검증 부재 | `cli/prompt.go:133-160` | `filepath.Clean` + `../` 패턴 검증 추가 |
| H-4 | `RotatingWriter.rotate()` Close 에러 무시 | `rotating_writer.go:62` | 에러 반환 또는 로깅 추가 |
| H-5 | 터미널 포트 출력 포맷 버그 | `terminal.go:29-33` | `SrcAddr:SrcPort → DstAddr:DstPort` 형태로 수정 |
| H-6 | `daemon.start()`의 `os.Exit(0)` 직접 호출 | `daemon_linux.go:43` | exit 함수 주입 또는 부모 프로세스에서만 호출 보장 |
| H-7 | `RawData` 슬라이스 버퍼 공유 위험 | `engine.go:83` | `append([]byte(nil), raw.Data()...)` 복사본 사용 |

### Phase 2 — 단기 수정 (MEDIUM, ~2시간)

| # | 이슈 | 파일 | 수정 방법 |
|---|------|------|----------|
| M-1 | 파일 권한 과다 (0666) | `rotating_writer.go:74`, `runner.go:123` | `os.OpenFile(..., 0600)` 사용 |
| M-2 | 디렉터리 권한 과다 (0755) | `rotating_writer.go:65` | `os.MkdirAll(..., 0700)` 변경 |
| M-3 | Linux/Darwin daemon 코드 중복 | `daemon_linux.go`, `daemon_darwin.go` | `//go:build !windows` 단일 파일로 통합 |
| M-4 | `bufio.NewReader` 중복 생성 | `cli/prompt.go` | 단일 Reader를 상위에서 생성 후 전달 |
| M-5 | Windows `isRunning` 불안정 | `daemon_windows.go:63` | `os.FindProcess` + 신호 기반으로 교체 |
| M-6 | `fmt.Println` stdout 직접 사용 | `runner.go:110` | `errOut` writer 사용으로 통일 |

### Phase 3 — CI/테스트 개선 (QA, ~1시간)

| # | 이슈 | 수정 방법 |
|---|------|----------|
| Q-1 | `-race` 플래그 미사용 | `test.yml`에 `go test -race` 추가 |
| Q-2 | 커버리지 측정 없음 | `-coverprofile=coverage.out` 추가 |
| Q-3 | `runner.go` 테스트 없음 | `runner_test.go` 신규 작성 |
| Q-4 | Windows/macOS 테스트 없음 | `release.yml`에 크로스 플랫폼 테스트 추가 |

---

## 3. 구현 순서

```
Phase 1 (HIGH 보안/품질)
  ├── H-1: PID 파일 O_EXCL + 경로 변경
  ├── H-2: 환경변수 → 임시파일 전달
  ├── H-3: 경로 검증 추가
  ├── H-4: Close() 에러 처리
  ├── H-5: 터미널 포맷 버그
  ├── H-6: os.Exit() 리팩터
  └── H-7: RawData 복사본

Phase 2 (MEDIUM)
  ├── M-1,2: 파일/디렉터리 권한 강화
  ├── M-3: Linux/Darwin 통합
  ├── M-4: bufio Reader 정리
  ├── M-5: Windows isRunning 개선
  └── M-6: stdout 출력 통일

Phase 3 (CI/Test)
  ├── Q-1,2: test.yml 개선
  ├── Q-3: runner_test.go 작성
  └── Q-4: 크로스 플랫폼 CI
```

---

## 4. 범위 외 (이번 작업에서 제외)

- LOW 이슈 (한/영 혼용, Writer Close 등) — 별도 개선 사이클
- `netif.Lister` 인터페이스 추출 — 별도 리팩터링
- Race detector CI 적용 (WSL/Linux 환경 필요)
