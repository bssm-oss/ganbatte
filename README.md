# ganbatte

> for lazy developers | 頑張って !

워크플로우/단축어 관리 CLI. 단순 alias 관리를 넘어 **명령 시퀀스를 워크플로우로 묶고, shell history에서 패턴을 자동 발굴해 추천**하는 도구.

## Features

- **Alias Management** — 자주 쓰는 명령에 이름 붙이기
- **Workflow Engine** — 여러 명령을 시퀀스로 묶어 실행, 파라미터 치환 지원
- **Multi-Format Config** — TOML / YAML / JSON 동등 지원
- **History Mining** — shell history 분석 → alias/workflow 자동 추천
- **TUI Browser** — fuzzy search + 미리보기로 빠르게 찾아 실행
- **Dry-run** — 실행 전 단계 미리보기, destructive 명령 강조
- **Import/Export** — 설정 백업 및 공유
- **Format Conversion** — 설정 포맷 간 변환
- **Dual Scope** — global + project 설정 분리, 자동 merge

## Install

```bash
# From source
go install github.com/bssm-oss/ganbatte@latest

# Or build locally
git clone https://github.com/bssm-oss/ganbatte.git
cd ganbatte
go build -o gnb .
```

## Quick Start

```bash
# 초기화 (셸 감지 + 포맷 선택)
gnb init

# alias 추가
gnb add gs "git status -sb"
gnb add ll "ls -la"

# 실행
gnb run gs

# 목록 보기
gnb list

# TUI 브라우저 (fuzzy search + preview)
gnb
```

## Commands

### Setup
```bash
gnb init                        # 설정 초기화
gnb init --format yaml          # YAML 포맷으로 초기화
gnb doctor                      # 환경 진단
```

### Alias / Workflow CRUD
```bash
gnb add <name> <command>        # alias 추가
gnb add --global gs "git status"  # global scope에 추가
gnb edit <name> <command>       # 수정
gnb remove <name>               # 삭제
gnb list                        # 전체 목록
gnb list --tag deploy           # 태그 필터링
gnb show <name>                 # 상세 정보
```

### Execution
```bash
gnb run <name> [args...]        # 실행
gnb run deploy main             # 파라미터 전달
gnb run deploy main --dry-run   # 실행 없이 미리보기
gnb                             # TUI 브라우저
```

### History Mining
```bash
gnb suggest                     # history 분석 → 추천
gnb suggest --apply             # 추천 항목 적용
gnb suggest --min-frequency 3   # 최소 빈도 조정
```

### Config Management
```bash
gnb config path                 # 활성 설정 파일 경로
gnb config convert --to yaml    # 포맷 변환
gnb export --output backup.toml # 내보내기
gnb import backup.toml          # 가져오기
gnb import shared.yaml --replace  # 덮어쓰기 모드
```

## Configuration

설정 파일은 TOML, YAML, JSON 중 원하는 포맷으로 작성:

```toml
# ~/.config/ganbatte/config.toml
version = "0.1.0"

[alias.gs]
cmd = "git status -sb"

[alias.gp]
cmd = "git push origin main"

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
- **Project**: `.ganbatte.{toml,yaml,json}` (현재 디렉토리)

Project scope가 global을 override. `gnb list`에서 scope 표시.

## TUI Browser

`gnb`를 인자 없이 실행하면 TUI 브라우저 진입:

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

## Supported Shells

History mining은 zsh, bash, fish를 지원:

| Shell | History File |
|-------|-------------|
| zsh | `~/.zsh_history` |
| bash | `~/.bash_history` |
| fish | `~/.local/share/fish/fish_history` |

## Development

```bash
go test ./...                   # 테스트
go test ./... -cover            # 커버리지
go vet ./...                    # 정적 분석
go build -o gnb .               # 빌드
```

## License

MIT
