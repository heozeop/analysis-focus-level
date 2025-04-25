package analyzer

import "gonum.org/v1/gonum/stat"

// FocusData: 1일치 집중도/라벨 데이터 구조
// (PRD 예시 JSON 구조와 일치)
type FocusData struct {
	Date        string            `json:"date"`
	TotalFocus  int               `json:"totalFocus"`
	Categories  map[string]int    `json:"categories"`
}

// AnalyzeFocus: 10분 단위 라벨/집중도 데이터 → FocusData 집계
func AnalyzeFocus(labels []string, scores []int) FocusData {
	categories := map[string]int{
		"업무": 0,
		"학습": 0,
		"취미": 0,
		"수면": 0,
		"이동": 0,
	}
	totalFocus := 0
	for i, label := range labels {
		score := 0
		if i < len(scores) {
			score = scores[i]
		}
		if _, ok := categories[label]; ok {
			categories[label] += score
			if label != "이동" {
				totalFocus += score
			}
		}
	}
	return FocusData{
		Categories:  categories,
		TotalFocus:  totalFocus,
	}
}

// Regression: 카테고리별 회귀 분석 (gonum/stat 활용)
func Regression(data []FocusData, category string) (slope, intercept float64) {
	xs := make([]float64, len(data))
	ys := make([]float64, len(data))
	for i, d := range data {
		xs[i] = float64(i)
		ys[i] = float64(d.Categories[category])
	}
	if len(xs) < 2 {
		return 0, 0
	}
	slope, intercept = stat.LinearRegression(xs, ys, nil, false)
	return
} 