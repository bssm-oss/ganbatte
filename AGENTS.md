# AGENTS.md

이 문서는 ganbatte 레포에서 AI 코딩 에이전트(Claude Code, Codex, Cursor 등)가 따라야 할 규칙이다. 사람도 읽을 수 있지만 1차 독자는 에이전트.

전체 스펙은 `docs/spec.md` 참조. 이 문서는 **"어떻게 작업하느냐"**에 집중한다.

---

## 0. 가장 중요한 규칙 (어기면 작업 거부)

1. **한국어로 응답한다.** 코드 주석/커밋/문서 영문 OK, 사람과의 대화는 한국어.
2. **`any` / `interface{}` 금지.** 타입은 명시적으로 정의한다. viper에서 값 꺼낼 때도 구조체로 unmarshal해라.
3. **`main` 브랜치에 직접 푸시 금지.** 모든 변경은 브랜치 → PR → merge. force push 금지.
4. **시크릿/`.env` 파일 절대 커밋 금지.** 실수로 읽거나 출력하지도 마라.
5. **확인 없이 destructive 명령 실행 금지.** `rm -rf`, `git reset --hard`, DB drop, 파일 대량 삭제 등은 사용자에게 먼저 물어봐라.
6. **모르면 추측하지 말고 Context7으로 확인해라.** 특히 cobra/viper/bubbletea 버전별 API.

---

## 1. 프로젝트 개요

- **이름**: ganbatte (바이너리: `gnb`)
- **한 줄**: lazy developers를 위한 워크플로우/단축어 관리 CLI
- **언어**: Go
- **상태**: feature-complete pre-release (스코프는 `docs/spec.md` 참조)

핵심 차별점은 두 가지:
- **멀티포맷 설정 (TOML / YAML / JSON 동등 지원)**
- **shell history mining 기반 alias/workflow 추천**

이 두 가지를 망가뜨리는 변경은 신중하게 접근해라.

---

## 2. 기술 스택 (버전 픽스)

| 라이브러리 | 용도 | 비고 |
|---|---|---|
| `github.com/spf13/cobra` | CLI 명령 정의 | completion 생성도 여기서 |
| `github.com/spf13/viper` | 설정 로드/저장 | 멀티포맷의 핵심 |
| `github.com/charmbracelet/bubbletea` | TUI 메인 루프 | Elm 아키텍처 |
| `github.com/charmbracelet/lipgloss` | 스타일링 | |
| `github.com/charmbracelet/bubbles` | TUI 컴포넌트 | textinput, list, viewport |
| `github.com/sahilm/fuzzy` | 퍼지 검색 | |

새 라이브러리 추가하려면 PR description에 **이유**를 적어야 한다. "그냥 좋아 보여서"는 안 됨. 표준 라이브러리로 충분한 건 표준 라이브러리로.

**버전 확인이 필요하면 Context7 MCP를 써라.** 메모리에 의존하지 마라. cobra v1.7과 v1.8은 작은 차이라도 있을 수 있다.

---

## 3. 아키텍처 규칙

```
cmd/        ← cobra 명령. 비즈니스 로직 금지. 인자 파싱 + internal/ 호출만.
internal/
  config/   ← viper 래퍼. IR 구조체, scoping, merge, portability, schema.
  workflow/ ← 실행 엔진. config를 import하지만 cmd는 import 안 함.
  history/  ← shell history 파싱 + suggest 엔진. 외부 파일 시스템 접근은 여기로 격리.
  tui/      ← bubbletea 모델. 다른 internal 패키지를 import해서 표시만.
  shell/    ← 셸 감지, history path, alias 마이그레이션 파서.
```

**규칙**:
- `cmd/`는 `internal/` 호출만. 직접 파일 I/O 금지.
- `internal/` 패키지끼리 순환 import 금지.
- `config`가 IR(Intermediate Representation)을 정의하고, 다른 패키지는 IR에만 의존한다. 포맷별 차이가 IR 밖으로 새면 안 됨.
- 전역 상태 금지. 설정은 명시적으로 인자로 넘겨라.

---

## 4. 멀티포맷의 신성함

JSON / YAML / TOML 세 포맷은 **100% 동등**해야 한다. 이건 ganbatte의 정체성이다.

작업 규칙:
- 새 설정 필드 추가하면 `testdata/fixtures/`에 **세 포맷 모두** 픽스처를 추가하고 동등성 테스트가 통과해야 한다.
- 한 포맷에서만 작동하는 기능 금지. 한 포맷에서 작동하면 세 포맷에서 작동해야 한다.
- 포맷별 직렬화 차이 (TOML의 array of tables vs YAML의 nested list 등)는 `internal/config/`에서 흡수하고 외부로 새지 않게 해라.
- 동등성 테스트는 `internal/config/equivalence_test.go`에 모은다.

---

## 5. 코드 컨벤션

### Go 일반
- `gofmt` 통과는 기본. CI에서 잡는다.
- `golangci-lint run` 통과 필수. 설정은 `.golangci.yml`.
- `errors.Is` / `errors.As` 사용. `err.Error() == "..."` 비교 금지.
- 에러는 wrap해서 올린다: `fmt.Errorf("loading config: %w", err)`.
- panic은 진짜 unrecoverable한 경우만. 일반 에러는 return.

### 타입
- **`any` / `interface{}` 금지** (재차 강조).
- viper에서 값 꺼낼 때: `viper.Unmarshal(&cfg)` 패턴. `viper.Get("key").(string)` 같은 type assertion 금지.
- 빈 인터페이스가 정말 필요한 자리는 PR에서 명시적으로 정당화해라.

### 네이밍
- 패키지명은 단수, 짧게 (`config`, `workflow`, `history`).
- exported 함수는 패키지명을 반복하지 마라 (`workflow.NewWorkflow` ❌, `workflow.New` ✅).
- 약어는 일관되게 (`URL` 전체 대문자, `Url` 금지).

### 주석
- exported 심볼은 모두 godoc 주석. `// FuncName ...`로 시작.
- 주석 영문 OK. 사용자 메시지는 영문 (i18n은 v1 이후).

---

## 6. 테스트

- **신규 기능에는 테스트가 함께 와야 한다.** "나중에 테스트 추가"는 금지.
- 테스트 파일은 같은 패키지 안에 `_test.go`.
- 통합 테스트는 `tests/integration/`. 빌드 태그 `//go:build integration`.
- 멀티포맷 동등성 테스트는 의무 (§4 참조).
- 외부 명령 실행이 필요한 부분(`os/exec`)은 인터페이스로 추상화해서 mock 가능하게 해라.

테스트 실행:
```bash
go test ./...                    # 단위 테스트
go test -tags=integration ./...  # 통합 포함
```

---

## 7. 빌드 / 실행

```bash
go build -o gnb .         # 로컬 빌드
go run . init             # 빌드 없이 실행
go install                # ~/go/bin에 설치
```

릴리즈 빌드는 `goreleaser` 사용 (v0.3에서 셋업 예정). 그 전까지는 수동.

---

## 8. Git 규칙

- **브랜치 네이밍**: `feat/`, `fix/`, `docs/`, `refactor/`, `test/`, `chore/` prefix.
  - 예: `feat/history-mining`, `fix/yaml-parser-edge-case`
- **커밋 메시지**: Conventional Commits.
  - `feat(history): add zsh history parser`
  - `fix(config): handle missing tags field in toml`
- **`main`은 보호 브랜치**. PR을 거쳐 merge만 가능. force push 금지, rebase merge 금지 (merge commit 유지).
- **PR은 작게**. 한 PR이 500줄 넘으면 쪼갤 방법을 먼저 고민해라.
- 커밋 메시지에 AI 도구 이름 박지 마라 ("Generated with Claude" 같은 거 금지).

---

## 9. 작업 흐름 (에이전트가 따라야 할 순서)

1. **이슈/요구사항 읽기**. 모호하면 작업 시작 전에 사용자에게 질문해라.
2. **관련 코드 탐색**. 손대기 전에 해당 영역이 어떻게 작동하는지 먼저 파악.
3. **계획 공유**. 큰 변경이면 코드 짜기 전에 접근 방식을 한국어로 사용자에게 설명하고 OK 받기.
4. **작은 단위로 구현**. 한 번에 한 가지.
5. **테스트 작성/실행**. 통과 확인.
6. **`go vet ./...`, `golangci-lint run` 통과 확인**.
7. **PR 생성**. description에 "왜"를 적어라. "무엇"은 diff가 말해준다.

---

## 10. 하지 마라 (Non-goals 재확인)

스펙 §8과 동일하지만 코드 작성 중에도 명심:

- ❌ 머신 간 sync 기능 추가 시도 (syncingsh로 위임)
- ❌ remote workflow 실행 (Ansible 아님)
- ❌ Make/just 대체 시도
- ❌ AI 기반 명령 생성 (history mining은 통계 기반, LLM 호출 없음)
- ❌ 플러그인 시스템 (v1.x 이후 검토)

이 카테고리에 해당하는 기능 요청이 들어오면 **거절하고 사용자에게 확인 요청**해라.

---

## 11. 사용자 환경 메모

이 레포 메인테이너(`@justn-hyeok`)가 작업할 때 알아둘 것:

- 주 개발 환경: macOS, Claude Code 기반
- 출력 언어: 한국어, 직설적/간결한 톤 선호
- 불필요한 사과, 과도한 설명, 부탁 안 한 제안 싫어함
- 잘못된 분석을 주면 push back함 — 받아들이고 수정해라
- 메인테이너가 "확인해줘"라고 하면 코드 변경 전에 정말로 확인 단계를 거쳐라

---

## 12. 이 문서 자체에 대해

- 이 AGENTS.md는 에이전트 지시문의 **단일 출처**다. CLAUDE.md, .cursorrules 등을 따로 두지 마라. 필요하면 그것들이 이 파일을 symlink/참조하게 해라.
- 규칙이 바뀌면 이 파일을 먼저 업데이트하고, 그 PR에서 코드 변경을 함께 정당화해라.
- 에이전트가 이 문서의 규칙과 사용자 지시가 충돌한다고 느끼면, 작업 멈추고 사용자에게 물어봐라.
