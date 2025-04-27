package analyzer

import (
	"fmt"

	"github.com/crispy/focus-time-tracker/internal/common"
	"gonum.org/v1/gonum/stat"
)

// AnalyzeFocus: 10분 단위 라벨/집중도 데이터 → FocusData 집계
// - labels: 각 10분 구간의 카테고리명 배열
// - scores: 각 10분 구간의 집중도 점수 배열
// 반환: FocusData (카테고리별 합계, 총점, 시간대별 점수)
func AnalyzeFocus(labels []string, scores []int) common.FocusData {
	categories := make(map[string]int) // 카테고리별 점수 합계
	maxScore := make(map[string]int)   // 카테고리별 최대 점수
	for _, cat := range common.Categories {
		categories[cat] = 0
		maxScore[cat] = 0
	}
	totalFocus := 0 // 하루 총 몰입 점수
	timeSlots := make(map[string]int) // 시간대별 점수 합계 (ex: "09:30" -> 40)
	for i, label := range labels {
		score := 0
		if i < len(scores) {
			score = scores[i]
		}
		if _, ok := categories[label]; ok {
			categories[label] += score // 카테고리별 합산
			maxScore[label] += 5       // 해당 카테고리 row 수 * 5
			if label != "이동" {
				totalFocus += score // "이동" 제외 총점
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
		MaxScore:   maxScore,
		TimeSlots:  timeSlots,
	}
}

// Regression: 카테고리별 회귀 분석 (gonum/stat 활용)
// - data: 여러 일자의 FocusData 배열
// - category: 분석할 카테고리명
// 반환: slope(기울기), intercept(절편)
func Regression(data []common.FocusData, category string) (slope, intercept float64) {
	xs := make([]float64, len(data)) // x축: 일자 인덱스
	ys := make([]float64, len(data)) // y축: 해당 카테고리 점수
	for i, d := range data {
		xs[i] = float64(i)
		ys[i] = float64(d.Categories[category])
	}
	if len(xs) < 2 {
		return 0, 0 // 데이터 2개 미만이면 회귀 불가
	}
	slope, intercept = stat.LinearRegression(xs, ys, nil, false)
	return
}
