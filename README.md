# ganbatte

> for lazy developers | 頑張って !

워크플로우/단축어 관리 CLI. 단순 alias 관리를 넘어 **셸 alias 자동 마이그레이션, 히스토리 패턴 발굴, 위험 명령 가드레일, 프로젝트 온보딩**까지 하나의 바이너리로.

## Why ganbatte?

### vs shell alias

| | shell alias | ganbatte |
|---|---|---|
| 등록 | `.zshrc` 직접 편집 | `gnb add gs "git status"` |
| 사용 | `gs` | `gs` (`eval "$(gnb shell-init)"` 이후 동일) |
| 파라미터 | function 작성 필요, 셸마다 문법 다름 | `cmd = "git checkout {branch}"` 선언 한 줄 |
| 위험 명령 보호 | 없음 | `confirm = true` 한 줄로 실행 전 확인 |
| 다른 머신 이동 | dotfiles 복사 | `gnb export/import` |
| 뭐 등록했는지 확인 | `alias` 치고 눈으로 스캔 | TUI fuzzy search + 미리보기 |
| 히스토리 분석 | `sort \| uniq -c` 직접 | `gnb suggest --apply` 원클릭 |
| 기존 alias 가져오기 | N/A | `gnb migrate` 원커맨드 |

### vs just / task / make

| | just/task/make | ganbatte |
|---|---|---|
| 용도 | 프로젝트 빌드/태스크 러너 | 개인 alias + 프로젝트 워크플로우 |
| 발견성 | `just --list` (텍스트) | TUI fuzzy search + 미리보기 |
| 개인 alias | 불가 (프로젝트 전용) | 글로벌 + 프로젝트 스코프 공존 |
| 히스토리 마이닝 | 없음 | `gnb suggest` |
| 크로스 셸 | 셸 의존적 | bash/zsh/fish 동일 설정 |

**한 문장으로**: alias가 많아질수록, 머신이 늘어날수록, 팀이 커질수록 ganbatte의 가치가 올라간다.

## Killer Feature: `gnb migrate`

기존 셸에 흩어진 alias를 원커맨드로 가져온다:

```bash
$ gnb migrate
Found 30 aliases in /Users/you/.zshrc
  gs = "git status"
  ll = "eza -alF --git"
  dc = "docker compose"
  ...

30 new alias(es) to import:
Import all? [Y/n] y
✓ 30 alias(es) imported
```

```bash
gnb migrate                # 자동 감지, 대화형 임포트
gnb migrate --shell zsh    # 특정 셸 지정
gnb migrate --dry-run      # 미리보기만
```

## Install

```bash
# Homebrew (macOS / Linux)
brew install justn-hyeok/tap/ganbatte

# go install
go install github.com/justn-hyeok/ganbatte@latest

# Or build locally
git clone https://github.com/justn-hyeok/ganbatte.git
cd ganbatte
go build -o gnb .
```

## Quick Start

```bash
# 1. 기존 alias 마이그레이션 (가장 빠른 시작)
gnb migrate

# 2. 셸에 등록
echo 'eval "$(gnb shell-init)"' >> ~/.zshrc
source ~/.zshrc

# 3. 바로 사용 — gs, ll 등 기존 alias가 그대로 동작
gs
```

또는 처음부터 시작:

```bash
gnb init                           # 설정 초기화
gnb add gs "git status -sb"        # alias 추가
eval "$(gnb shell-init)"           # 셸에 등록
gs                                 # 바로 사용
```

## Features

### Shell Integration

```bash
# .zshrc / .bashrc에 추가
eval "$(gnb shell-init)"

# fish
gnb shell-init | source
```

등록된 모든 alias/workflow가 네이티브 셸 명령으로 동작.

### Parameterized Aliases

셸 function 대신 선언적으로:

```toml
[alias.gco]
cmd = "git checkout {branch}"
params = ["branch"]

[alias.glog]
cmd = "git log --oneline -{count}"
params = ["count"]
default_params = { count = "10" }
```

```bash
gco feature/login    # → git checkout feature/login
glog                 # → git log --oneline -10
glog 30              # → git log --oneline -30
```

### Confirm Guard

파괴적 명령에 선언적 안전장치:

```toml
[alias.nuke]
cmd = "git reset --hard HEAD"
confirm = true
```

```
$ nuke
⚠ Run "git reset --hard HEAD"? [y/N]: y
Running: git reset --hard HEAD
```

`gnb run nuke --yes`로 CI에서는 스킵 가능.

### Workflow Engine

여러 명령을 시퀀스로 묶어 실행:

```toml
[workflow.deploy]
description = "Build and deploy"
params = ["branch"]
steps = [
  { run = "pnpm lint" },
  { run = "pnpm test", on_fail = "stop" },
  { run = "pnpm build" },
  { run = "git push origin {branch}", confirm = true },
]
tags = ["deploy", "ci"]
```

`on_fail` 옵션: `stop` (중단), `continue` (무시), `prompt` (사용자에게 질문).

### History Mining

```bash
gnb suggest                     # 빈도 기반 alias + 시퀀스 기반 workflow 추천

=== Alias Suggestions (frequency) ===
  1. gs = "git status -sb"
     Used 47 times

=== Workflow Suggestions (sequences) ===
  1. wf-g
     Step 1: git add .
     Step 2: git commit
     Step 3: git push
     Sequence appeared 23 times

gnb suggest --apply             # 추천 항목 즉시 적용
```

### Project Onboarding

`.ganbatte.toml`을 repo에 커밋하면 새 팀원이 즉시 생산적:

```bash
gnb init --project               # 프로젝트 설정 생성

# .ganbatte.toml
[workflow.setup]
description = "프로젝트 초기 세팅"
steps = [
  { run = "npm install" },
  { run = "cp .env.example .env" },
]
tags = ["onboarding"]
```

```bash
git clone <repo> && cd <repo>
gnb                              # TUI에서 프로젝트 워크플로우 탐색
```

TUI에서 `[project]` 태그로 글로벌/프로젝트 스코프 구분 표시.

### TUI Browser

`gnb`를 인자 없이 실행하면 TUI 브라우저:

| Key | Action |
|-----|--------|
| `↑/k` `↓/j` | 이동 |
| `Enter` | 실행 |
| `/` | 검색 |
| `e` | 에디터에서 편집 |
| `d` | 삭제 (확인) |
| `t` | 태그 필터 순환 |
| `?` | 도움말 |
| `q` | 종료 |

## Commands

```bash
# Setup
gnb init                           # 글로벌 설정 초기화
gnb init --project                 # 프로젝트 설정 초기화
gnb init --format yaml             # 특정 포맷으로
gnb doctor                         # 환경 진단
gnb migrate                        # 셸 alias 마이그레이션

# Shell Integration
gnb shell-init                     # 셸 함수 출력
gnb shell-init --shell fish        # 특정 셸

# CRUD
gnb add <name> <command>           # alias 추가
gnb edit <name> <command>          # 수정
gnb remove <name>                  # 삭제
gnb list                           # 전체 목록
gnb list --tag deploy              # 태그 필터링
gnb show <name>                    # 상세 정보

# Execution
gnb run <name> [args...]           # 실행
gnb run deploy main --dry-run      # 미리보기
gnb run nuke --yes                 # confirm 스킵

# History Mining
gnb suggest                        # 추천
gnb suggest --apply                # 추천 + 적용

# Config Management
gnb config path                    # 설정 파일 경로
gnb config convert --to yaml       # 포맷 변환
gnb export -o backup.toml          # 내보내기
gnb import backup.toml             # 가져오기
```

## Configuration

TOML, YAML, JSON 동등 지원:

```toml
# ~/.config/ganbatte/config.toml
version = "1.0.0"

[alias.gs]
cmd = "git status -sb"

[alias.gco]
cmd = "git checkout {branch}"
params = ["branch"]

[alias.nuke]
cmd = "git reset --hard HEAD"
confirm = true

[workflow.deploy]
description = "Build and deploy"
params = ["branch"]
steps = [
  { run = "pnpm lint" },
  { run = "pnpm test", on_fail = "stop" },
  { run = "pnpm build" },
  { run = "git push origin {branch}", confirm = true },
]
tags = ["deploy", "ci"]
```

### Scopes

- **Global**: `~/.config/ganbatte/config.{toml,yaml,json}`
- **Project**: `.ganbatte.{toml,yaml,json}` (현재 디렉토리, repo에 커밋)

Project scope가 global을 override.

### Supported Shells

| Shell | History | Config |
|-------|---------|--------|
| zsh | `~/.zsh_history` | `~/.zshrc` |
| bash | `~/.bash_history` | `~/.bashrc`, `~/.bash_aliases` |
| fish | `~/.local/share/fish/fish_history` | `~/.config/fish/config.fish` |

## Development

```bash
go test ./...                      # 테스트
go test -race ./...                # race detector
go vet ./...                       # 정적 분석
go build -o gnb .                  # 빌드
go run cmd/gendoc.go               # man page 생성
```

## License

MIT
