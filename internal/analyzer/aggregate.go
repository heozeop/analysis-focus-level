package analyzer

import (
	"fmt"

	"github.com/crispy/focus-time-tracker/internal/common"
	"gonum.org/v1/gonum/stat"
)

// AnalyzeFocus: 10분 단위 라벨/집중도 데이터 → FocusData 집계
func AnalyzeFocus(labels []string, scores []int) common.FocusData {
	categories := make(map[string]int)
	for _, cat := range common.Categories {
		categories[cat] = 0
	}
	totalFocus := 0
	timeSlots := make(map[string]int)
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
		// 시간대별 몰입 합계 계산 (10분 단위)
		hour := i / 6
		min := (i % 6) * 10
		timeKey := fmt.Sprintf("%02d:%02d", hour, min)
		timeSlots[timeKey] += score
	}
	return common.FocusData{
		Categories: categories,
		TotalFocus: totalFocus,
		TimeSlots:  timeSlots,
	}
}

// Regression: 카테고리별 회귀 분석 (gonum/stat 활용)
func Regression(data []common.FocusData, category string) (slope, intercept float64) {
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
