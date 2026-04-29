# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| latest  | Yes       |

## Reporting a Vulnerability

보안 취약점을 발견하셨다면 **공개 이슈로 등록하지 마시고**, 아래 방법으로 비공개 보고해주세요:

- GitHub Security Advisory: [Report a vulnerability](https://github.com/bssm-oss/ganbatte/security/advisories/new)

보고 시 포함해주실 내용:
- 취약점 설명
- 재현 단계
- 영향 범위
- 가능하면 수정 제안

48시간 내에 초기 응답을 드리겠습니다.

## Scope

ganbatte는 CLI 도구로, 주요 보안 관심사:

- **명령 실행**: `gnb run`은 사용자가 등록한 명령을 `sh -c`로 실행합니다. 설정 파일을 신뢰할 수 없는 소스에서 가져올 때 주의하세요.
- **설정 파일**: `.ganbatte.*` 파일을 repo에서 clone할 때 워크플로우 내용을 확인하세요.
- **히스토리 접근**: `gnb suggest`와 `gnb migrate`는 셸 히스토리/설정 파일을 읽습니다. 민감한 명령이 포함될 수 있습니다.

## Best Practices

- `gnb import`로 외부 설정을 가져올 때는 `--replace` 없이 먼저 내용을 확인하세요.
- 프로젝트 `.ganbatte.*` 파일의 워크플로우를 실행 전에 `gnb show`로 검토하세요.
- `confirm = true`를 파괴적 명령에 적극 활용하세요.
