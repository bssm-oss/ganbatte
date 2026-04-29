# ganbatte (`gnb`)

[English](./README.md) | [명령어 레퍼런스](./docs/man) | [스펙](./docs/spec.md) | [릴리즈](https://github.com/bssm-oss/ganbatte/releases)

> lazy developers를 위한 워크플로우/단축어 관리 CLI. 頑張って!

`ganbatte`는 흩어진 shell alias와 반복 명령 시퀀스를 **이식 가능하고, 검색 가능하고, 프로젝트 스코프를 이해하는 CLI**로 정리해준다. shell alias의 빠른 사용감은 유지하면서, 파라미터, 위험 명령 확인, workflow, 히스토리 기반 추천을 더한다.

```bash
gnb add gs "git status -sb"
eval "$(gnb shell-init)"
gs
```

## 왜 만들었나

shell alias는 빠르지만 많아질수록 dotfile 더미가 된다. make, just, task는 프로젝트 빌드 태스크에는 좋지만 개인 단축어, cross-shell alias 관리, 히스토리 기반 추천에는 맞지 않는다.

`gnb`는 그 중간 지점에 있다.

| 필요 | shell alias | make/just/task | ganbatte |
|---|---|---|---|
| 개인 단축어 | 지원 | 부적합 | 지원 |
| 프로젝트 온보딩 workflow | 수동 | 지원 | 지원 |
| cross-shell 설정 | 어려움 | 일부 | 지원 |
| shell function 없는 파라미터 | 어려움 | 지원 | 지원 |
| 위험 명령 확인 | 없음 | 수동 | 지원 |
| TUI 탐색 | 없음 | 보통 없음 | 지원 |
| 히스토리 기반 추천 | 없음 | 없음 | 지원 |

## 킬러 피처

### 1. 기존 shell alias 마이그레이션

`.zshrc`, `.bashrc`, `.bash_aliases`, fish 설정에 흩어진 alias를 `ganbatte` 설정으로 가져온다.

```bash
gnb migrate
gnb migrate --dry-run
gnb migrate --shell zsh
```

예시:

```text
$ gnb migrate
Found 17 aliases in /Users/you/.zshrc
Found 8 aliases in /Users/you/.bash_aliases

25 new alias(es) to import:
  gs = "git status -sb"
  ll = "eza -alF --git"
  dc = "docker compose"

Import all? [Y/n] y
```

### 2. 실제 사용 기록 기반 추천

`gnb shell-init`을 켜면 명령 실행 기록을 `~/.local/share/ganbatte/track.log`에 가볍게 기록할 수 있다. 이때 매 명령마다 `gnb` 바이너리를 띄우지 않고 shell built-in으로 append하므로 레이턴시가 거의 없다.

`gnb suggest`는 shell history 또는 track log를 분석해 alias, 파라미터 alias, workflow 후보를 추천한다.

```bash
gnb suggest
gnb suggest --apply
gnb suggest --from-history
```

추천은 단순 빈도가 아니라 예상 키 입력 절약량 기준으로 정렬된다. `--apply` 중 파괴적 명령이 감지되면 자동으로 `confirm = true`가 붙는다.

## 설치

### Homebrew

```bash
brew install --cask bssm-oss/tap/ganbatte
```

### Go

```bash
go install github.com/bssm-oss/ganbatte/cmd/gnb@latest
```

### 릴리즈 아카이브

[GitHub Releases](https://github.com/bssm-oss/ganbatte/releases)에서 플랫폼별 아카이브를 받고, `gnb`를 `PATH`에 넣으면 된다.

## 빠른 시작

### 기존 alias가 있을 때

```bash
gnb migrate
echo 'eval "$(gnb shell-init)"' >> ~/.zshrc
source ~/.zshrc

gs           # 가져온 alias 실행
dc up -d     # native shell function처럼 동작
```

bash는 같은 `eval` 줄을 `~/.bashrc`에 넣으면 된다. fish는 다음처럼 쓴다.

```fish
gnb shell-init | source
```

### 처음부터 시작할 때

```bash
gnb init
gnb add gs "git status -sb"
gnb run gs
```

인자 없이 `gnb`를 실행하면 TUI 브라우저가 열린다.

## 핵심 기능

### Global / Project 스코프

Global 설정은 `~/.config/ganbatte/config.{toml,yaml,json}`에 저장된다. Project 설정은 현재 디렉토리 또는 신뢰 가능한 상위 repo의 `.ganbatte.{toml,yaml,json}`에 저장된다.

같은 이름이 있으면 project 항목이 global 항목을 override한다. 이때 `gnb run`은 실행 전 확인을 요구한다.

```bash
gnb list --scope global
gnb list --scope project
gnb run setup --yes
```

### 파라미터 alias

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
gco feature/login
glog
glog 30
```

### Workflow

```toml
[workflow.deploy]
description = "Lint, test, build, and push"
params = ["branch"]
tags = ["deploy", "ci"]

[[workflow.deploy.steps]]
run = "pnpm lint"

[[workflow.deploy.steps]]
run = "pnpm test"
on_fail = "stop"

[[workflow.deploy.steps]]
run = "git push origin {branch}"
confirm = true
```

```bash
gnb run deploy main --dry-run
gnb run deploy main
```

### 위험 명령 가드

실수로 실행되면 안 되는 명령에는 `confirm = true`를 붙인다.

```toml
[alias.clean-docker]
cmd = "docker system prune -af"
confirm = true
```

### TUI 브라우저

```bash
gnb
```

TUI에서는 fuzzy search, preview, 태그 필터, global/project 라벨을 지원한다.

### 멀티포맷 설정

TOML, YAML, JSON을 동등하게 지원한다.

```bash
gnb init --format toml
gnb config convert --to yaml
gnb config path
```

## 명령어 요약

| 명령어 | 용도 |
|---|---|
| `gnb` | TUI 브라우저 열기 |
| `gnb init` | global 또는 project 설정 생성 |
| `gnb add <name> <cmd>` | alias 추가 |
| `gnb edit <name> <cmd>` | alias 수정 |
| `gnb remove <name>` | alias 또는 workflow 삭제 |
| `gnb list` | alias/workflow 목록 |
| `gnb show <name>` | 상세 정보 보기 |
| `gnb run <name> [args...]` | alias 또는 workflow 실행 |
| `gnb migrate` | shell alias 가져오기 |
| `gnb suggest` | history 기반 추천 보기 |
| `gnb shell-init` | shell integration 함수 출력 |
| `gnb doctor` | 설정/환경 진단 |
| `gnb export` / `gnb import` | 설정 이동/백업 |

전체 명령어 문서는 [docs/man](./docs/man)을 보면 된다.

## 설정 예시

```toml
version = "0.1.0"

[alias.gs]
cmd = "git status -sb"
tags = ["git"]

[alias.gco]
cmd = "git checkout {branch}"
params = ["branch"]

[alias.nuke]
cmd = "git reset --hard HEAD"
confirm = true

[workflow.test]
description = "Run checks"
tags = ["ci"]

[[workflow.test.steps]]
run = "go test ./..."

[[workflow.test.steps]]
run = "go vet ./..."
```

## 문서

- [문서 인덱스](./docs/README.md)
- [전체 프로젝트 스펙](./docs/spec.md)
- [생성된 man page](./docs/man)
- [기여 가이드](./CONTRIBUTING.md)
- [변경 기록](./CHANGELOG.md)

## 개발

```bash
go build -o gnb .
go test ./...
go test -race ./...
go vet ./...
go run cmd/gendoc.go
```

릴리즈 빌드는 `v*` 태그 push 시 GoReleaser가 처리한다. 현재 릴리즈는 GitHub release archive, `go install github.com/bssm-oss/ganbatte/cmd/gnb@v1.5.6`, Homebrew Cask 경로로 실제 설치 검증을 마쳤다.

## 하지 않는 것

`ganbatte`는 머신 간 sync 서비스, remote execution 시스템, Make/just 대체제, AI 명령 생성기, plugin platform이 아니다.

## License

MIT
