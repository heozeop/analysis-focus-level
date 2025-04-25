오케이! **10분 단위 세분화 + 라벨/집중도 기록 + 이동 제외 집계**까지 PRD에 반영해서 완성해줄게.

---

# 📄 Focus Time Tracker & Analyzer PRD (Golang + Interactive Graphs Version - Finalized)

## 1. **목표 (Goal)**

- **Google Sheets**를 DB로 사용해 몰입 시간(집중 시간)을 기록하고, **일간 점수화**, **회귀 분석**을 자동화.
- 분석 결과를 **다른 GitHub 레포지토리**에 PR & 머지 → **GitBook**으로 문서화.
- **3개월 / 6개월 / 12개월** 기간별 동적 추세 그래프를 **HTML + TypeScript + Chart.js** 기반으로 GitBook에서 인터랙티브하게 제공.
- **카테고리 라벨 (업무, 학습, 취미, 수면, 이동)**을 **10분 단위 셀**에 설정하고, **라벨별 집중도**를 입력해 **카테고리별 분석 및 시각화**를 지원한다.  
- **이동** 카테고리는 **집계에서 제외**한다.

---

## 2. **핵심 기능 (Key Features)**

### 2.1. **Google Sheets 자동 생성 (cron)**  
- 매년 **1월 1일** GitHub Actions `cron`으로:
  - 새로운 Google Sheets 생성 (파일명: `YYYY Focus Log`)
  - 12개 탭 (`1월` ~ `12월`) 생성  
  - 각 탭 구조:
    - A열: `시간 (00:00 ~ 23:50)` → **10분 단위**로 분할 (144개 row)
    - B열 이후: 날짜 + 요일 (`1일(월)` …)  
    - 각 날짜는 **2개 컬럼으로 구성**:
      1. **Label 컬럼**: `업무`, `학습`, `취미`, `수면`, `이동` 중 선택 (Data Validation)
      2. **Focus 컬럼**: **집중도 점수** 입력 (0 ~ 100)

### 2.2. **일일 데이터 생성 및 저장**
- **매일 (23:55)**:
  - Google Sheets에서 **10분 단위 데이터** 파싱:
    - 각 셀의 **라벨** + **집중도 점수**
    - **이동**은 총집계에서 제외
  - **일별 JSON 데이터 파일** 생성:
    - 예시: `assets/data/2025-04-25.json`
    - 구조:
      ```json
      {
        "date": "2025-04-25",
        "totalFocus": 450,
        "categories": {
          "업무": 300,
          "학습": 100,
          "취미": 50,
          "수면": 0,
          "이동": 120  // 별도 집계
        }
      }
      ```
  - GitBook repo의 **assets/data/** 경로에 저장  
  - PR 생성 → GitBook repo에 머지

### 2.3. **회귀 분석 및 동적 추세 그래프 (기간 선택 가능)**
- **GitBook 페이지 로드 시**:
  - **assets/data/**에 있는 **일별 JSON 파일 전체 로드**
  - **JavaScript + Chart.js**로:
    - **3개월 / 6개월 / 12개월** 기간 선택  
    - **카테고리별 집중 시간 추세선** + **회귀선** 시각화  
    - **이동**은 별도 표시 (총몰입시간 집계에서 제외)

---

## 3. **기술 스택 (Tech Stack)**

| 기능                         | 기술/패키지                                 |
|----------------------------|------------------------------------------|
| Google Sheets 연동            | `google.golang.org/api/sheets/v4`         |
| 데이터 분석 (회귀)            | `gonum/stat`                               |
| 일별 데이터 저장              | JSON 포맷 → GitBook repo `assets/data/`   |
| 그래프 생성 (동적)            | **Chart.js** + **TypeScript** + **HTML**  |
| 자동화 (스케줄링 + PR)        | GitHub Actions (`cron`) + `go-git` or `os/exec` |

---

## 4. **워크플로우 (Workflow)**

1. **Google Sheets 생성 (매년 1월 1일)**:
   - GitHub Actions cron → Golang 앱으로 Sheets 생성  
   - 12개월 탭 + 날짜/시간 구조 설정 (10분 단위 / 날짜별 2컬럼 → Label + Focus 점수)

2. **일일 데이터 생성 및 저장 (매일 23:55)**:
   - Google Sheets → **일별 JSON 데이터 생성** (`assets/data/2025-04-25.json`)
   - **이동** 제외한 카테고리별 몰입 시간 집계  
   - PR 생성 → GitBook repo의 **assets/data/** 경로에 저장

3. **추세 그래프 & 회귀 분석 (GitBook 페이지 로드 시)**:
   - GitBook에서 **assets/data/**에 있는 **일별 데이터 전체 로드**
   - 기간 & 카테고리 선택 → 추세선 + 회귀선 렌더링

---

## 5. **GitBook 구조 예시**

```
/analytics/
  └── 2025/
      ├── 01.md
      ├── trend.md                 # 인터랙티브 그래프 페이지
/assets/
  └── data/
      ├── 2025-04-23.json
      ├── 2025-04-24.json
      ├── 2025-04-25.json
```

---
