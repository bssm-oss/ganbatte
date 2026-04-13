# ganbatte

> for lazy developers | 頑張って !

워크플로우/단축어 관리 CLI. lazyasf의 정신적 후속작으로, 단순 alias 관리를 넘어 **명령 시퀀스를 워크플로우로 묶고, shell history에서 패턴을 자동 발굴해 추천**하는 도구.

---

## 0. 철학

- **Lazy by default**: 사용자가 손가락을 덜 움직일수록 ganbatte가 잘하고 있는 것
- **Do one thing well**: 워크플로우 관리에만 집중. 머신 간 sync는 syncingsh에 위임
- **No magic, but obvious defaults**: `gnb init` 한 번이면 즉시 쓸 수 있어야 함
- **Shell-agnostic**: zsh / bash / fish 모두 동등 지원
- **Format-agnostic**: JSON / YAML / TOML 모두 지원, 사용자가 고르게 둠

---

## 1. 핵심 개념

### 1.1 Alias
가장 단순한 단위. 단일 명령 또는 짧은 셸 표현식에 이름을 붙인 것.

```toml
[alias.gs]
cmd = "git status -sb"
```

### 1.2 Workflow
여러 단계(step)를 묶은 시퀀스. 파라미터, 조건부 실행, dry-run을 지원함.

```toml
[workflow.deploy]
description = "Lint, test, build, push"
params = ["branch"]
steps = [
  { run = "pnpm lint" },
  { run = "pnpm test", on_fail = "stop" },
  { run = "pnpm build" },
  { run = "git push origin {branch}", confirm = true },
]
tags = ["deploy", "ci"]
```

### 1.3 Scope
- **Global**: `~/.config/ganbatte/config.{toml,yaml,json}` — 어디서든 작동
- **Project**: 현재 디렉토리 또는 상위 디렉토리에 `.ganbatte.{toml,yaml,json}`이 있으면 자동 로드. global을 override하지 않고 merge

### 1.4 Tag
워크플로우/alias에 라벨링. TUI에서 필터링, `gnb list --tag deploy` 같은 조회에 사용.

---

## 2. 멀티 포맷 지원 (JSON / YAML / TOML)

### 2.1 원칙
- 사용자가 어느 포맷을 쓰든 동등하게 작동
- 내부적으로 동일한 IR(Intermediate Representation)로 파싱 → 같은 코드 경로
- `gnb init` 시 사용자에게 포맷 선택을 묻고, 그 이후로는 해당 포맷으로 통일
- `gnb config convert --to yaml` 같은 명령으로 언제든 전환 가능

### 2.2 파일 탐색 우선순위
같은 디렉토리에 여러 포맷이 있으면 다음 순서로 채택 (충돌 시 경고):

1. `.ganbatte.toml` / `config.toml`
2. `.ganbatte.yaml` / `config.yaml` (`.yml`도 허용)
3. `.ganbatte.json` / `config.json`

### 2.3 동등 예시

**TOML**
```toml
[workflow.test]
steps = [
  { run = "pnpm lint" },
  { run = "pnpm test" },
]
tags = ["dev"]
```

**YAML**
```yaml
workflow:
  test:
    steps:
      - run: pnpm lint
      - run: pnpm test
    tags: [dev]
```

**JSON**
```json
{
  "workflow": {
    "test": {
      "steps": [
        { "run": "pnpm lint" },
        { "run": "pnpm test" }
      ],
      "tags": ["dev"]
    }
  }
}
```

세 파일은 100% 동일하게 동작해야 함. 테스트도 같은 픽스처를 세 포맷으로 만들어 동등성 검증.

---

## 3. 명령어 레퍼런스

### 3.1 셋업
| 명령 | 설명 |
|---|---|
| `gnb init` | 인터랙티브 위자드. 셸 감지, 포맷 선택, 예시 워크플로우 생성 |
| `gnb completion <shell>` | 셸 자동완성 스크립트 출력 |
| `gnb doctor` | 설정 파일 유효성, 셸 통합 상태, 충돌 진단 |

### 3.2 Alias / Workflow CRUD
| 명령 | 설명 |
|---|---|
| `gnb add <name> <cmd...>` | alias 추가 |
| `gnb edit <name> <cmd>` | 기존 alias 수정 (워크플로우는 설정 파일 직접 편집) |
| `gnb edit <name> <cmd> [--global]` | 기존 항목 수정 |
| `gnb remove <name>` | 삭제 |
| `gnb list [--tag <tag>] [--scope global|project]` | 목록 |
| `gnb show <name>` | 상세 정보 |

### 3.3 실행
| 명령 | 설명 |
|---|---|
| `gnb run <name> [args...]` | 워크플로우/alias 실행 |
| `gnb run <name> --dry-run` | 실제 실행 없이 단계 미리보기 |
| `gnb` | TUI 브라우저 진입 (인자 없을 때) |

### 3.4 History Mining
| 명령 | 설명 |
|---|---|
| `gnb suggest` | shell history 분석해 alias/workflow 후보 추천 |
| `gnb suggest --apply` | 추천된 항목 인터랙티브하게 등록 |

### 3.5 설정
| 명령 | 설명 |
|---|---|
| `gnb config path` | 활성 설정 파일 경로 출력 |
| `gnb config convert --to <format>` | 설정 파일 포맷 변환 |

---

## 4. 주요 기능 상세

### 4.1 History Mining (시그니처 기능)
- `~/.zsh_history`, `~/.bash_history`, fish history DB를 읽어 빈도/패턴 분석
- 같은 명령을 N회 이상 친 경우 alias 후보로 추천
- 연속해서 함께 등장하는 명령 시퀀스는 workflow 후보로 추천
- 예: "너 이번 주에 `git add . && git commit -m ... && git push`를 23번 쳤어. workflow로 만들래?"
- 추천만 하고 자동 등록은 안 함 (`--apply`로 명시 동의 필요)

### 4.2 TUI 브라우저
- ratatui 기반
- 인자 없이 `gnb`만 치면 진입
- fzf 스타일 fuzzy search
- 화살표/검색으로 워크플로우 선택 → Enter로 실행
- 미리보기 패널에 명령 시퀀스 표시
- `?`로 도움말, `e`로 편집, `d`로 삭제, `t`로 태그 필터

### 4.3 Dry-run / Conditional Steps
- `--dry-run`: 모든 단계를 실행 없이 출력. destructive 명령(`rm`, `git push -f`, `DROP TABLE` 등)은 빨간색 강조
- `confirm = true`인 단계는 실행 전 y/N 프롬프트
- `on_fail`: `stop` (기본) / `continue` / `prompt`
- 환경변수 `{branch}` 같은 플레이스홀더는 실행 시점 치환

### 4.4 Project Scope
- 디렉토리 진입 시 `.ganbatte.*` 자동 탐색 (현재 → 상위 디렉토리 순)
- global과 project가 같은 이름을 가지면 project가 우선
- `gnb list`는 두 스코프를 구분해서 표시

---

## 5. syncingsh 연동

ganbatte는 머신 간 sync를 직접 구현하지 않음. 대신:

- 모든 사용자 데이터를 `~/.config/ganbatte/` 한 곳에 둠
- `gnb config path`로 syncingsh가 동기화할 디렉토리를 알 수 있게 함
- syncingsh 측에서 `gnb config path` 출력을 그대로 watch 대상에 등록하면 끝
- ganbatte는 sync 충돌, 인증, 네트워크에 대해 아무것도 모름

이렇게 하면 ganbatte는 워크플로우 관리에만, syncingsh는 동기화에만 집중하면서 두 도구가 자연스럽게 물림.

---

## 6. 기술 스택

| 영역 | 선택 | 이유 |
|---|---|---|
| 언어 | Go | 단일 바이너리, 빠른 컴파일, AI 친화적 idiom, GC로 borrow 에러 없음 |
| CLI | cobra | 표준. gh / kubectl / hugo / helm 모두 사용 |
| 설정 | viper | JSON/YAML/TOML 멀티포맷이 라이브러리 차원에서 공짜 |
| TUI | bubbletea + lipgloss + bubbles | Elm 아키텍처, React 익숙하면 친숙 |
| 퍼지 검색 | sahilm/fuzzy | bubbletea와 잘 물림 |
| 셸 통합 | os/exec + cobra completion | completion 생성은 cobra 내장 |

---

## 7. 디렉토리 구조 (예정)

```
ganbatte/
├── go.mod
├── go.sum
├── main.go
├── cmd/                  # cobra 명령 정의 (init, run, add, list, ...)
│   ├── root.go
│   ├── init.go
│   ├── run.go
│   └── ...
├── internal/
│   ├── config/           # viper 래퍼, 멀티포맷 로드/저장, IR
│   ├── workflow/         # 실행 엔진, dry-run, conditional
│   ├── alias/
│   ├── history/          # zsh/bash/fish 히스토리 파서, mining
│   ├── tui/              # bubbletea 모델/뷰
│   └── shell/            # 셸 감지, completion 헬퍼
├── testdata/
│   └── fixtures/         # 같은 설정의 toml/yaml/json 동등 픽스처
├── AGENTS.md
└── README.md
```

---

## 8. 비목표 (Non-goals)

ganbatte가 **하지 않는 것**을 명확히 해두자:

- ❌ 머신 간 sync (→ syncingsh)
- ❌ remote workflow 실행 (Ansible 아님)
- ❌ task runner 대체 (Make/just 아님 — 그것들을 alias로 묶을 수는 있음)
- ❌ shell 자체 대체
- ❌ AI 기반 명령 생성 (history mining은 통계 기반)

---

## 9. 버전별 스코프

### v0.1 — MVP "It works"
**목표**: 혼자 쓰면서 dogfooding 가능한 최소 범위

- [x] 멀티 포맷 설정 파일 (TOML / YAML / JSON) 파싱·저장
- [x] alias CRUD (`add`, `remove`, `list`, `edit`)
- [x] workflow CRUD + 시퀀스 실행
- [x] 파라미터 플레이스홀더 (`{branch}` 같은)
- [x] global / project scope
- [x] `gnb run` 기본 실행
- [x] `gnb init` 위자드 (셸 감지 + 포맷 선택 + 예시 생성)
- [x] shell completion 자동 생성 (`gnb completion`)
- [x] 태그 필드 (스키마만, 필터링은 v0.2)
- [x] 기본 에러 처리, `--help`

**Definition of Done**: 본인이 일주일간 lazyasf 대신 ganbatte만 써도 불편 없을 것

---

### v0.2 — "It's actually useful"
**목표**: 시그니처 기능 도입, 외부 사용자 받기 시작

- [ ] **History mining** (`gnb suggest`, `gnb suggest --apply`)
  - zsh / bash / fish 히스토리 파서
  - 빈도 기반 alias 추천
  - 시퀀스 패턴 기반 workflow 추천
- [ ] **TUI 브라우저** (`gnb` 단독 실행)
  - ratatui + nucleo
  - 검색 / 실행 / 미리보기
- [ ] **Dry-run** (`--dry-run`)
- [ ] **Conditional steps** (`on_fail`, `confirm`)
- [ ] 태그 필터링 (`gnb list --tag`)
- [ ] `gnb show` 상세 보기
- [ ] `gnb doctor` 진단 명령

**Definition of Done**: README에 데모 GIF 박았을 때 "오 이거 뭐야" 반응이 나오는 수준

---

### v0.3 — "Plays well with others"
**목표**: 생태계 통합, syncingsh 연동, 외부 공유

- [ ] `gnb config convert --to <format>` (포맷 간 변환)
- [ ] `gnb config path` (syncingsh 연동용)
- [ ] **Import / Export** 명시적 명령 (`gnb export`, `gnb import`)
- [ ] 카테고리 UI (TUI에서 그룹 표시)
- [ ] 충돌 해결 UX (같은 이름 alias가 두 스코프에 있을 때)
- [ ] syncingsh 연동 가이드 문서
- [ ] CI: GitHub Actions에서 멀티 OS 빌드 (Linux / macOS / Windows)
- [ ] Homebrew formula, cargo install, 바이너리 릴리즈

**Definition of Done**: 친구가 깔아서 자기 머신 두 대 사이에 syncingsh로 sync해서 쓰는 게 가능

---

### v1.0 — "Stable"
**목표**: API 안정화, 1.x 동안 깨지 않을 약속

- [ ] 설정 파일 스키마 freeze + 마이그레이션 도구
- [ ] 안정된 CLI 인터페이스 (deprecation 정책 수립)
- [ ] 전체 명령어 man page
- [ ] 공식 문서 사이트 (또는 README 충실화)
- [ ] 100+ 테스트 케이스, 멀티 포맷 동등성 보장
- [ ] 퍼블릭 릴리즈, 조직 레포에서 Public 전환

---

### v1.x+ (아이디어 풀, 미정)

- 플러그인 시스템 (외부 명령으로 ganbatte 확장)
- 워크플로우 마켓플레이스 (커뮤니티 공유)
- 셸 prompt 통합 (현재 디렉토리에서 사용 가능한 워크플로우 힌트)
- `gnb watch` (파일 변경 감지 → workflow 트리거)
- macOS Shortcuts / Raycast 연동

---

## 10. 네이밍 / 브랜딩

- 풀네임: **ganbatte**
- 바이너리: `gnb` (3글자, lazy하게)
- 태그라인: *for lazy developers | 頑張って !*
- 컨셉의 핵심 아이러니: lazy한 개발자한테 "힘내(頑張って)"라고 외쳐주는 도구
