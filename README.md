# Focus Time Tracker

## 소개

이 프로젝트는 몰입 시간 분석 및 시각화, Google Sheets 연동, 자동화된 리포트/그래프 생성 등을 지원하는 Go 기반 분석 도구입니다.

## 개발 환경
- Go 1.17 이상 권장
- 주요 의존성: gonum/plot, Google Sheets API 등

## 코드 포맷팅 (Google 스타일)

코드 스타일은 Google Go 스타일을 따르며, 아래 도구를 사용해 자동 포맷팅합니다:
- **gofmt**: 기본 Go 포맷터
- **goimports**: import 자동 정리
- **gofumpt**: Google 스타일에 더 엄격한 포맷터

### 포맷팅 실행 방법

```sh
make format
```
- gofmt, goimports, gofumpt가 순서대로 실행됩니다.
- goimports, gofumpt가 설치되어 있지 않으면 설치 안내 메시지가 출력됩니다.

#### 도구 설치
아래 명령어로 필요한 도구를 설치하세요:
```sh
go install golang.org/x/tools/cmd/goimports@latest
go install mvdan.cc/gofumpt@latest
```
설치 후, `$HOME/go/bin`이 PATH에 포함되어 있어야 합니다.

### PATH 설정 예시
```sh
export PATH=$HOME/go/bin:$PATH
```

## 테스트

```sh
go test ./...
```

## 주요 Makefile 명령어
- `make format` : 코드 자동 포맷팅 (Google 스타일)

## 기타
- Google Sheets 연동, 자동화, 그래프 생성 등은 소스 내 주석 및 예시 코드를 참고하세요.
- 린트(lint)는 별도로 제공하지 않으며, 필요시 golangci-lint 등 추가 도구를 직접 설치해 사용할 수 있습니다.

---

문의/기여는 PR 또는 이슈로 남겨주세요. 