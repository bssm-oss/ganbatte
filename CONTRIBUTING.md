# Contributing to ganbatte

ganbatte에 기여해주셔서 감사합니다. 이 문서는 기여 방법을 안내합니다.

## Development Setup

```bash
# 1. Fork & clone
git clone https://github.com/<your-username>/ganbatte.git
cd ganbatte

# 2. Build
go build -o gnb .

# 3. Test
go test -race ./...

# 4. Lint (optional, CI에서 자동 실행)
golangci-lint run ./...
```

**요구사항**: Go 1.22+

## Branch Strategy

- `main`에 직접 push하지 마세요.
- feature branch에서 작업 후 PR을 보내주세요.
- branch 이름: `feat/description`, `fix/description`, `docs/description`

## Commit Convention

[Conventional Commits](https://www.conventionalcommits.org/) 형식:

```
feat: add gnb stats command
fix: resolve Windows path issue in export
docs: update README with new examples
chore: bump bubbletea to v1.4.0
test: add coverage for parameterized aliases
```

## Pull Request

1. 관련 이슈가 있으면 연결해주세요.
2. 새 기능이면 테스트를 포함해주세요.
3. `go test -race ./...`가 통과해야 합니다.
4. PR 설명에 변경 이유를 적어주세요.

## Project Structure

```
cmd/                 Cobra commands (no business logic)
internal/
  config/            Config loading, scoping, portability
  history/           Shell history parsers, suggest engine
  shell/             Shell detection, migration
  tui/               Bubbletea TUI model
  workflow/          Step execution engine
testdata/fixtures/   Test data files
docs/man/            Auto-generated man pages
```

**규칙**:
- `cmd/`에는 인자 파싱 + `internal/` 호출만
- `internal/` 패키지 간 순환 import 금지
- `any` / `interface{}` 사용 금지 — 명시적 타입 정의
- 에러는 `fmt.Errorf("context: %w", err)` 형식으로 wrapping

## Testing

- 모든 exported function에 테스트 작성
- `testify/assert` + `testify/require` 사용
- fixture 기반 테스트는 `testdata/fixtures/` 활용
- workflow 테스트는 `MockExecutor` 인터페이스 사용

```bash
go test ./...                   # 전체 테스트
go test -race ./...             # race detector
go test -v ./internal/config/   # 특정 패키지
go test -run TestMigrate ./...  # 특정 테스트
```

## Adding a New Command

1. `cmd/<name>.go` 생성 (기존 커맨드 패턴 참고)
2. `init()`에서 `RootCmd.AddCommand()` 호출
3. `cmd/add_test.go`의 `executeCmd` 헬퍼에 플래그 리셋 추가
4. `go run cmd/gendoc.go`로 man page 재생성
5. README.md Commands 섹션 업데이트

## Reporting Bugs

[GitHub Issues](https://github.com/justn-hyeok/ganbatte/issues)에 다음 정보와 함께 등록:

- OS / shell / Go version
- 재현 단계
- 기대 동작 vs 실제 동작
- 가능하면 `gnb doctor` 출력

## License

기여하신 코드는 MIT License로 배포됩니다.
