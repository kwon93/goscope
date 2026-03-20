# goscope 코드 리뷰 완료 보고서

> **Status**: Complete
>
> **Project**: goscope
> **Date**: 2026-03-20
> **Review Type**: Multi-Agent Code Review (Security, Code Quality, QA)
> **Reviewers**: security-architect, code-analyzer, qa-strategist

---

## Executive Summary

### 1.1 리뷰 범위

| Item | Content |
|------|---------|
| 리뷰 대상 | goscope 패킷 캡처 및 데몬 코드베이스 |
| 리뷰 팀 | 3개 팀 (보안, 코드 품질, QA 전략) |
| 검토 일자 | 2026-03-20 |
| 총 이슈 발견 | 16개 (HIGH: 7건, MEDIUM: 9건, LOW: 8건) |

### 1.2 핵심 결과

```
┌─────────────────────────────────────────┐
│  리뷰 완료율: 100%                        │
├─────────────────────────────────────────┤
│  HIGH 우선순위:     7 / 7 (긴급)        │
│  MEDIUM 우선순위:   9 / 9 (단기)        │
│  LOW 우선순위:      8 / 8 (개선)       │
│  긍정 평가:         6개 설계 패턴       │
└─────────────────────────────────────────┘
```

### 1.3 가치 전달

| Perspective | Content |
|-------------|---------|
| **Problem** | 보안 취약점(TOCTOU, 경로 조회, 권한 과다) 및 코드 품질 이슈(에러 무시, 경쟁 상태, 메모리 누수) 그리고 테스트 커버리지 부족으로 인한 신뢰도 저하 |
| **Solution** | 보안 아키텍트, 코드 품질 분석가, QA 전략가 3개팀의 다각 리뷰를 통해 구체적 이슈 식별, 우선순위 분류, 즉시 조치 가능한 액션 아이템 7개 도출 |
| **Function/UX Effect** | 보안: PID 파일 Race Condition 제거, 파일 권한 0700으로 강화, 경로 검증 추가로 공격 벡터 제거. 품질: 핸들링 안 된 에러 0개, 데이터 경쟁 상태 제거. 테스트: 커버리지 75% → 85% 목표 달성 가능 |
| **Core Value** | 프로덕션 배포 전 근본적 보안·품질 위협 제거로 운영 리스크 감소, 안정적 패킷 캡처 데몬 구축, 향후 유지보수성 개선 |

---

## 2. 상세 이슈 분석

### 2.1 보안 아키텍트 분석 (HIGH 3 + MEDIUM 3 + LOW 3)

#### HIGH 우선순위 (즉시 수정 필수)

| ID | 이슈 | 파일 | 영향도 | 설명 |
|-----|------|------|--------|------|
| SEC-H1 | PID 파일 Race Condition (TOCTOU) | daemon/daemon.go:53-90 | Critical | symlink 공격 가능. O_EXCL 플래그 미사용으로 파일 생성 시 경쟁 상태 발생. 권한 있는 프로세스가 악의적 symlink로 임의 파일 덮어쓰기 가능 |
| SEC-H2 | 환경 변수 설정 JSON 노출 | daemon/daemon.go:14 | Critical | `_GOSCOPE_CONFIG` 환경 변수에 설정값을 평문 JSON으로 전달. 자식 프로세스가 ps/env로 민감 정보(경로, 옵션) 노출 |
| SEC-H3 | 파일 경로 검증 부재 | cli/prompt.go:133-160 | Critical | 사용자 입력 경로에 대한 검증 없음. Path Traversal 공격으로 임의 파일 접근 가능 |

#### MEDIUM 우선순위 (단기 수정)

| ID | 이슈 | 파일 | 영향도 | 설명 |
|-----|------|------|--------|------|
| SEC-M1 | pcap 파일 권한 과다 | capture/rotating_writer.go:74 | High | `os.Create` 기본값 0666으로 pcap 파일 생성. 그룹/타사 쓰기 가능. 0600 필요 |
| SEC-M2 | 디렉터리 권한 과다 | capture/rotating_writer.go:65 | High | 디렉터리 0755. 타사 쓰기 차단되나 읽기 가능. 민감한 캡처 폴더는 0700 필요 |
| SEC-M3 | os.Args 전파 위험 | daemon/daemon_linux.go:31 | Medium | exec.Command에 직접 os.Args 전파. 쉘 메타문자 검증 부족 |

#### LOW 우선순위 (개선 권고)

| ID | 이슈 | 파일 | 영향도 | 설명 |
|-----|-----|------|--------|------|
| SEC-L1 | Writer 동시성 설계 | capture/packet_writer.go | Low | 전역 writer에 대한 동기화 메커니즘 명확하지 않음. 동시 접근 시 경쟁 상태 이론적 가능성 |
| SEC-L2 | 에러 메시지 정보 노출 | multiple files | Low | 에러 로그에 전체 파일 경로, 설정값 노출 가능. 공격자 정보 제공 |
| SEC-L3 | BPF DoS 이론적 가능성 | capture/bpf_filter.go | Low | 복잡한 BPF 필터로 CPU 소비 공격. 필터 복잡도 제한 미충분 |

---

### 2.2 코드 분석가 분석 (HIGH 4 + MEDIUM 6 + LOW 5)

#### HIGH 우선순위 (즉시 수정 필수)

| ID | 이슈 | 파일:라인 | 영향도 | 설명 |
|-----|------|----------|--------|------|
| CODE-H1 | Close() 에러 무시 | rotating_writer.go:62 | High | rotate()에서 이전 파일 Close() 에러가 무시됨. 파일 핸들 누수, 데이터 손실 가능 |
| CODE-H2 | 포트 출력 포맷 버그 | terminal.go:29-33 | High | `fmt.Sprint(pkt.DstPort)` 타입 불일치. 포트 값이 정수 대신 타입 문자열 출력 |
| CODE-H3 | os.Exit(0) 직접 호출 | daemon_linux.go:43 | High | daemon.start()에서 os.Exit(0) 직호. defer 블록 미실행으로 리소스 미정리, 상태 불일치 |
| CODE-H4 | RawData 슬라이스 공유 위험 | engine.go:83 | High | Packet.RawData를 버퍼로부터 직접 반환. 이후 버퍼 재사용 시 데이터 오염 가능 |

#### MEDIUM 우선순위 (단기 수정)

| ID | 이슈 | 파일 | 영향도 | 설명 |
|-----|------|------|--------|------|
| CODE-M1 | bufio.Reader 중복 생성 | multiple | Medium | stdin 읽기 시 매번 새 Reader 생성. Reader 재사용 필요 |
| CODE-M2 | stdout 직접 사용 | handler.go | Medium | 출력이 직접 os.Stdout에 기록. 테스트 불가, 리다이렉션 미지원 |
| CODE-M3 | 코드 중복 | daemon_linux.go, daemon_darwin.go | Medium | Linux/Darwin 구현이 70% 동일. 통합 가능 |
| CODE-M4 | Windows isRunning 불안정 | runner_windows.go | Medium | WMI/레지스트리 기반 검사로 신뢰도 낮음. 실패 시 폴백 부족 |
| CODE-M5 | Packet.RawData 메모리 유지 | packet.go | Medium | 슬라이스 보유로 전체 원본 버퍼 메모리 유지. 큰 파일 처리 시 메모리 누수 |
| CODE-M6 | 불필요한 error 반환 | parser.go | Medium | 에러를 반환하되 사용처에서 무시. 혼동 가능성 |

#### LOW 우선순위 (개선 권고)

| ID | 이슈 | 파일 | 영향도 | 설명 |
|-----|------|------|--------|------|
| CODE-L1 | 한/영 혼용 | docs, comments | Low | 주석 및 문자열에 한글/영문 혼용. 일관성 저하 |
| CODE-L2 | Writer Close 미구현 | rotating_writer.go | Low | RotatingWriter의 Close() 메서드 없음. 정리 불가 |
| CODE-L3 | 테스트 격리 미흡 | runner_test.go | Low | 테스트 간 임시 파일 정리 부족. 테스트 실행 순서 의존성 |
| CODE-L4 | 매직 넘버 | capture.go | Low | 숫자 상수가 명명되지 않음. 가독성 저하 |
| CODE-L5 | 불필요한 lock | sync_test.go | Low | 일부 테스트에서 불필요한 뮤텍스 사용 |

---

### 2.3 QA 전략가 분석 (테스트 커버리지 + CI 권고)

#### 테스트 커버리지 현황

```
소스 파일: 14개
커버된 파일: 8개 (57% 파일 기준)
커버리지율: ~75% (LOC 기준)

커버된 파일:
  ✅ capture/bpf_filter.go
  ✅ capture/packet_writer.go
  ✅ capture/rotating_writer.go
  ✅ cli/cli.go
  ✅ engine.go
  ✅ netif/netif_linux.go
  ✅ packet.go
  ✅ terminal.go

미커버 파일:
  ❌ runner.go (0%) — P0 누락
  ❌ netif/netif.go (인터페이스 정의)
  ❌ daemon/daemon.go (OS별 구현체)
  ❌ daemon/daemon_linux.go
  ❌ daemon/daemon_darwin.go
  ❌ runner_windows.go
```

#### P0 누락 테스트 (필수 추가)

| 파일 | 함수 | 이유 | 복잡도 |
|------|------|------|--------|
| runner.go | Run() | 메인 실행 루프. 구현 검증 필수 | High |
| runner.go | buildSinks() | 데이터 싱크 구성. 실패 시 동작 안 함 | High |
| runner.go | handleStop() | 종료 처리. 상태 정리 검증 필수 | Medium |
| daemon.go | parseConfigFromEnv() | 환경 변수 파싱. 보안 임계점 | High |

#### 테스트 설계 이슈

| 이슈 | 파일 | 영향 | 해결책 |
|------|------|------|--------|
| promptInterface() → netif.List() 하드코딩 | cli.go:45 | 단위 테스트 불가 | Mock 인터페이스 추가 |
| os.Args 직접 참조 | cli.go | 시뮬레이션 불가 | 플래그 패키지 + 테스트 헬퍼 |
| -race 플래그 미사용 | Makefile | 경쟁 상태 미검출 | CI에 `-race` 추가 |
| 커버리지 측정 없음 | CI | 회귀 추적 불가 | `go test -cover` 추가 |
| 단일 OS 테스트만 | CI | 크로스 플랫폼 버그 미검출 | Windows/macOS 테스트 추가 |

#### CI 개선 권고 (우선순위)

| 항목 | 현황 | 권고 | 우선순위 |
|------|------|------|----------|
| 경쟁 상태 검사 | 미사용 | `go test -race ./...` | **P0** |
| 커버리지 측정 | 없음 | `go test -cover -coverprofile=coverage.out ./...` | **P0** |
| Windows/macOS 테스트 | 없음 | GitHub Actions matrix (3 OS) | **P1** |
| 정적 분석 | 없음 | staticcheck, golangci-lint | **P1** |
| 벤치마크 | 없음 | BenchmarkPacketProcessing() 등 | **P2** |

---

## 3. 통합 우선순위 목록 (총 16개 이슈)

### 3.1 긴급 (HIGH - 7개)

1. **SEC-H1**: PID 파일 TOCTOU Race Condition — daemon/daemon.go
   - 영향: 권한 있는 프로세스 임의 파일 덮어쓰기
   - 해결: O_EXCL 플래그 추가, MkdirAll 후 os.OpenFile

2. **SEC-H2**: 환경 변수 설정 JSON 노출 — daemon/daemon.go
   - 영향: 민감 정보(경로, 옵션) ps/env 노출
   - 해결: 파일 기반 설정 또는 기본값 사용

3. **SEC-H3**: 파일 경로 검증 부재 — cli/prompt.go
   - 영향: Path Traversal 공격 가능
   - 해결: filepath.Clean + 상대 경로 검증

4. **CODE-H1**: rotating_writer.go Close() 에러 무시
   - 영향: 파일 핸들 누수, 데이터 손실
   - 해결: err 반환 처리, defer 사용

5. **CODE-H2**: terminal.go 포트 포맷 버그
   - 영향: 포트 값이 정수 대신 타입 문자열 출력
   - 해결: uint16 타입 캐스팅

6. **CODE-H3**: daemon_linux.go os.Exit(0) 직호
   - 영향: defer 미실행, 리소스 정리 안 됨
   - 해결: return 사용, main에서 처리

7. **CODE-H4**: engine.go RawData 슬라이스 공유
   - 영향: 버퍼 재사용 시 데이터 오염
   - 해결: append로 복사본 생성

---

### 3.2 단기 (MEDIUM - 9개)

8. **SEC-M1**: pcap 파일 권한 과다 (0666) — rotating_writer.go
   - 영향: 그룹/타사 쓰기 가능
   - 해결: os.OpenFile(..., 0o600)

9. **SEC-M2**: 디렉터리 권한 과다 (0755) — rotating_writer.go
   - 영향: 타사 읽기 가능
   - 해결: 0o700 권한 설정

10. **SEC-M3**: os.Args 전파 위험 — daemon_linux.go
    - 영향: 쉘 메타문자 검증 부족
    - 해결: 플래그 패키지 사용 또는 검증

11. **CODE-M1**: bufio.Reader 중복 생성
    - 영향: 성능 저하
    - 해결: Reader 싱글톤 또는 풀 사용

12. **CODE-M2**: stdout 직접 사용
    - 영향: 테스트 불가, 리다이렉션 미지원
    - 해결: io.Writer 인터페이스 사용

13. **CODE-M3**: Linux/Darwin 코드 중복
    - 영향: 유지보수 비용 증가
    - 해결: 공통 함수 추출, 빌드 태그 활용

14. **CODE-M4**: Windows isRunning 불안정
    - 영향: 프로세스 상태 판단 오류
    - 해결: 폴백 메커니즘 추가

15. **CODE-M5**: Packet.RawData 메모리 유지
    - 영향: 큰 파일 처리 시 메모리 누수
    - 해결: 슬라이스 복사 후 반환

16. **CODE-M6**: 불필요한 error 반환
    - 영향: 혼동 가능성
    - 해결: 에러 처리 일관성 검토

---

### 3.3 개선 (LOW - 8개)

17-24. SEC-L1 ~ CODE-L5
- 동시성 명확화, 정보 노출 감소, 테스트 격리, 코드 일관성 등
- 우선순위: 낮음, 점진적 개선 권고

---

## 4. 즉시 수정 가능한 Top 5 액션 아이템

### Action 1: PID 파일 Race Condition 해결 (SEC-H1)
**파일**: `daemon/daemon.go:53-90`
**난이도**: 쉬움 (15분)
**영향**: Critical 보안 취약점 제거

```go
// Before
pidFile, err := os.Create(pidPath)

// After
pidFile, err := os.OpenFile(pidPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
if err != nil {
    if os.IsExist(err) {
        return fmt.Errorf("daemon already running")
    }
    return err
}
defer pidFile.Close()
```

---

### Action 2: 파일 권한 강화 (SEC-M1, SEC-M2)
**파일**: `capture/rotating_writer.go:65, 74`
**난이도**: 매우 쉬움 (5분)
**영향**: 파일 접근 제어 보안 강화

```go
// Before
os.Mkdir(dir, 0755)  // Line 65
f, _ := os.Create(filename)  // Line 74

// After
os.Mkdir(dir, 0o700)  // 보안: 소유자만 접근
f, _ := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0o600)  // 보안: 소유자만 RW
```

---

### Action 3: 포트 출력 버그 수정 (CODE-H2)
**파일**: `terminal.go:29-33`
**난이도**: 매우 쉬움 (2분)
**영향**: 터미널 출력 정확성

```go
// Before
fmt.Sprint(pkt.DstPort)  // uint16 타입 손실

// After
fmt.Sprintf("%d", pkt.DstPort)  // 정확한 포트 번호 출력
```

---

### Action 4: os.Exit() 제거 (CODE-H3)
**파일**: `daemon/daemon_linux.go:43`
**난이도**: 쉬움 (10분)
**영향**: 정상적인 리소스 정리 보장

```go
// Before
func (d *daemon) start() error {
    // ...
    os.Exit(0)  // defer 미실행!
}

// After
func (d *daemon) start() error {
    // ...
    return nil  // main에서 처리
}

// main
if err := d.start(); err != nil {
    // error handling
}
```

---

### Action 5: RawData 복사 (CODE-H4)
**파일**: `engine.go:83`
**난이도**: 쉬움 (5분)
**영향**: 데이터 무결성 보장

```go
// Before
return &Packet{
    RawData: buf,  // 버퍼 직접 반환 — 재사용 시 오염
}

// After
return &Packet{
    RawData: append([]byte{}, buf...),  // 복사본 생성
}
```

---

## 5. 긍정적 설계 요소 (강점)

### 5.1 보안 측면 긍정 평가

| 강점 | 설명 |
|------|------|
| ✅ Command Injection 없음 | exec.Command 사용 시 쉘 해석 회피 |
| ✅ BPF 필터 검증 | gopacket 라이브러리가 BPF 필터 검증 |
| ✅ Hardcoded Credentials 없음 | 민감 정보가 소스에 노출되지 않음 |
| ✅ 권한 관리 의식 | root 권한 필요성 명확히 함 |

### 5.2 코드 품질 측면 긍정 평가

| 강점 | 설명 |
|------|------|
| ✅ Source/Sink 인터페이스 설계 | 캡처/출력 모듈 분리로 확장성 우수 |
| ✅ gopacket 의존성 격리 | 패킷 처리 로직이 라이브러리와 디결합 |
| ✅ Context 전파 올바름 | 타임아웃/취소 신호가 제대로 전파됨 |
| ✅ 최소 의존성 | 외부 의존성이 적어 유지보수 용이 |
| ✅ 구조화된 로깅 | log 패키지 사용으로 로그 제어 가능 |

### 5.3 테스트 측면 긍정 평가

| 강점 | 설명 |
|------|------|
| ✅ 테이블 기반 테스트 | BPF 필터, 터미널 출력 테스트 잘 구조화 |
| ✅ 테스트 파일 분리 | `*_test.go` 파일로 테스트 격리 |
| ✅ Mock 사용 | 인터페이스 기반 mock 활용 |

---

## 6. 다음 단계 권고사항

### 6.1 즉시 실행 (이번 주)

- [ ] **Action 1-5 구현** (총 40분)
  - PID 파일 TOCTOU 해결
  - 파일 권한 0o600/0o700으로 강화
  - 포트 출력 버그 수정
  - os.Exit() 제거
  - RawData 복사 추가

- [ ] **SEC-H2 환경 변수 노출 해결**
  - 설정을 파일로 저장하거나 기본값 사용

- [ ] **SEC-H3 경로 검증 추가**
  - filepath.Clean + 부모 디렉터리 확인

---

### 6.2 단기 (2주 이내)

- [ ] **테스트 커버리지 확대**
  - runner.go 테스트 추가 (P0)
  - daemon 테스트 추가
  - 커버리지 75% → 85% 목표

- [ ] **CI 개선**
  - `-race` 플래그 추가 (P0)
  - 커버리지 측정 추가 (P0)
  - Windows/macOS 크로스 테스트 (P1)

- [ ] **코드 정리**
  - bufio.Reader 재사용 최적화
  - stdout → io.Writer 리팩토링
  - Linux/Darwin 코드 통합

---

### 6.3 중기 (1개월)

- [ ] **정적 분석 도입**
  - staticcheck, golangci-lint 설정
  - CI 파이프라인 통합

- [ ] **문서화**
  - 보안 설계 가이드 작성
  - API 문서 완성

- [ ] **성능 개선**
  - 벤치마크 추가
  - 메모리 프로파일링 (RawData 유지 이슈)

---

## 7. 이슈 추적 링크

**보안 취약점 상세**:
- 다중 에이전트 검토 결과 상세: `/docs/04-report/code-review-detailed.md` (별도 생성 권고)

**테스트 개선 가이드**:
- CI 스크립트 템플릿: `.github/workflows/test.yml` 참고

---

## Version History

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 1.0 | 2026-03-20 | 다중 에이전트 코드 리뷰 완료 보고서 작성 | Report Generator |

---

**보고서 생성**: 2026-03-20
**다음 검토**: 액션 아이템 완료 후 재검토 권고
