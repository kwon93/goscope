# code-review-fix Analysis Report

> **Analysis Type**: Gap Analysis (Plan vs Implementation)
>
> **Project**: goscope
> **Analyst**: gap-detector agent
> **Date**: 2026-03-20
> **Plan Doc**: [code-review-fix.plan.md](../01-plan/features/code-review-fix.plan.md)

---

## 1. Analysis Overview

### 1.1 Analysis Scope

- **Plan Document**: `docs/01-plan/features/code-review-fix.plan.md`
- **Implementation Files**:
  - `internal/capture/terminal.go`
  - `internal/capture/engine.go`
  - `internal/capture/rotating_writer.go`
  - `internal/daemon/daemon.go`
  - `internal/daemon/daemon_unix.go` (신규)
  - `internal/cli/runner.go`
  - `internal/cli/prompt.go`
  - `internal/cli/runner_test.go` (신규)
  - `.github/workflows/test.yml`
- **Analysis Date**: 2026-03-20

---

## 2. Gap Analysis (Plan vs Implementation)

### 2.1 Phase 1 — HIGH Issues

| # | Issue | File | Verdict | Evidence |
|---|-------|------|:-------:|----------|
| H-1 | PID 파일 권한 0600 | `daemon.go:54` | DONE | `os.WriteFile(..., 0600)` |
| H-1 | PID 파일 `O_EXCL` + user-dir | `daemon_unix.go` | **PARTIAL** | 경로 여전히 `/tmp/goscope.pid`, `O_EXCL` 없음 |
| H-2 | 환경변수 설정 → 임시파일 | `daemon_unix.go:34-36` | DEFERRED | Plan 명시적 연기; env var 계속 사용 |
| H-3 | 경로 순회 검증 추가 | `prompt.go:140-144,162` | DONE | `..` 및 null byte 체크 존재 |
| H-3 | `filepath.Clean` 사용 | `prompt.go` | **DEVIATION** | `strings.Contains` 사용, `filepath.Clean` 미적용 |
| H-4 | RotatingWriter Close 에러 처리 | `rotating_writer.go:62-64` | DONE | 에러 wrap 후 반환 |
| H-5 | 터미널 포트 포맷 버그 | `terminal.go:29-34` | DONE | `SrcAddr:SrcPort` / `DstAddr:DstPort` 정상 |
| H-6 | `os.Exit` 리팩터 | `daemon_unix.go:45` | DEFERRED | Plan 명시적 연기 |
| H-7 | RawData 버퍼 복사 | `engine.go:83` | DONE | `append([]byte(nil), raw.Data()...)` |

### 2.2 Phase 2 — MEDIUM Issues

| # | Issue | File | Verdict | Evidence |
|---|-------|------|:-------:|----------|
| M-1 | pcap 파일 권한 0600 | `runner.go:123` | DONE | `os.OpenFile(..., 0600)` |
| M-2 | 디렉터리 권한 0700 | `rotating_writer.go:67` | DONE | `os.MkdirAll(r.dir, 0700)` |
| M-3 | Linux/Darwin → `daemon_unix.go` | `daemon_unix.go` | DONE | `//go:build !windows`; 구 파일 삭제 |
| M-6 | `fmt.Println` → `fmt.Fprintln(errOut)` | `runner.go` | DONE | `fmt.Println` 잔존 없음 |

### 2.3 Phase 3 — CI/Test

| # | Issue | File | Verdict | Evidence |
|---|-------|------|:-------:|----------|
| Q-1 | CI `-race` 플래그 | `test.yml:28` | DONE | `go test -race` 존재 |
| Q-2 | CI `-coverprofile` | `test.yml:28` | DONE | `-coverprofile=coverage.out` 존재 |
| Q-3 | `runner_test.go` 신규 작성 | `cli/runner_test.go` | DONE | 8개 테스트 함수 |

### 2.4 Match Rate 요약

```
┌─────────────────────────────────────────────────┐
│  Overall Match Rate: 92%                 [Pass]  │
├─────────────────────────────────────────────────┤
│  DONE:                11개                       │
│  PARTIAL:              2개 (H-1b, H-3b)          │
│  DEFERRED (제외):       2개 (H-2, H-6)            │
│  NOT IMPLEMENTED:      0개                       │
├─────────────────────────────────────────────────┤
│  계산: (11 + 2×0.5) / 12 = 92%                  │
│  (Deferred 항목은 분모에서 제외)                   │
└─────────────────────────────────────────────────┘
```

---

## 3. 차이점 상세

### 3.1 미구현 항목

| Item | Plan 내용 | 설명 | 심각도 |
|------|-----------|------|:------:|
| H-1 `O_EXCL` + user-dir | plan.md L34 | PID 파일 `/tmp/goscope.pid`, `O_EXCL` 없어 TOCTOU 위험 잔존 | HIGH |
| H-3 `filepath.Clean` | plan.md L36 | `strings.Contains("..")` 사용 — `foo/./bar/../../..` 같은 우회 가능 | MEDIUM |

### 3.2 연기된 항목 (Plan 명시)

| Item | Plan 주석 | 현재 상태 |
|------|-----------|----------|
| H-2 | "현재 세션: 권한만 수정, 임시파일 전환 미수행" | `_GOSCOPE_CONFIG` 환경변수 계속 사용 |
| H-6 | "현재 세션: 미수행" | `os.Exit(0)` `daemon_unix.go:45` 잔존 |

### 3.3 변경된 구현 (Plan 명세 ≠ 실제)

| Item | Plan 명세 | 실제 구현 | 영향 |
|------|-----------|----------|:----:|
| H-3 메커니즘 | `filepath.Clean` + `../` 검증 | `strings.Contains("..")` + `\x00` 검사 | LOW |

---

## 4. 신규/삭제 파일

| 파일 | 상태 | 설명 |
|------|:----:|------|
| `internal/daemon/daemon_unix.go` | 신규 생성 | Linux/Darwin 통합 (`//go:build !windows`) |
| `internal/cli/runner_test.go` | 신규 생성 | runner.go 8개 테스트 |
| `internal/daemon/daemon_linux.go` | 삭제 | `daemon_unix.go`로 통합 |
| `internal/daemon/daemon_darwin.go` | 삭제 | `daemon_unix.go`로 통합 |

---

## 5. 권고 사항

### 5.1 즉시 (다음 릴리즈 전)

| 우선순위 | 항목 | 파일 | 액션 |
|:--------:|------|------|------|
| 1 | H-1 완성 | `daemon.go`, `daemon_unix.go` | `O_EXCL` 추가; 경로를 `os.UserCacheDir()/goscope/goscope.pid`로 이동 |
| 2 | H-3 강화 | `prompt.go` | `filepath.Clean(name)` 적용 후 `..` 체크 |

### 5.2 다음 이터레이션 (연기 항목 해소 시)

| 우선순위 | 항목 | 파일 | 액션 |
|:--------:|------|------|------|
| 3 | H-2 | `daemon_unix.go` | 임시파일(0600)로 설정 전달 후 즉시 삭제 |
| 4 | H-6 | `daemon_unix.go` | `exitFn` 주입 또는 `os.Exit`을 `main()`에서만 호출 |

---

## 6. 최종 점수

```
┌──────────────────────────────────────────────────┐
│  Overall Score                                    │
├──────────────────────────────────────────────────┤
│  전체 Match Rate (Plan 대비):  92%      [Pass ✅] │
│  Phase 1 HIGH (비연기 항목):   80%      [Warn ⚠️] │
│  Phase 2 MEDIUM:              100%      [Pass ✅] │
│  Phase 3 CI/Test:             100%      [Pass ✅] │
└──────────────────────────────────────────────────┘
```

---

## Version History

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 1.0 | 2026-03-20 | 초기 Gap 분석 | gap-detector |
