# Focus Time Tracker & Analyzer

몰입 시간(집중 시간) 기록, 분석, 시각화 자동화 도구 (Go + Google Sheets + GitHub Actions + Chart.js)

## 주요 기능
- Google Sheets 연동: 10분 단위 집중도/라벨 기록
- 일별 JSON 데이터 자동 생성 및 PR
- 카테고리별 집중 시간/회귀 분석
- GitBook용 동적 추세 그래프 데이터 제공

## 프로젝트 구조
```
cmd/tracker/         # 메인 실행 파일
internal/sheets/     # Google Sheets 연동
internal/analyzer/   # 데이터 분석/회귀
internal/exporter/   # JSON/PR 자동화
internal/config/     # 설정/환경변수
assets/data/         # 일별 JSON 데이터
scripts/             # 운영 자동화 스크립트
.github/workflows/   # GitHub Actions
```

## 빌드/실행
```sh
go build -o tracker ./cmd/tracker
./tracker
``` 