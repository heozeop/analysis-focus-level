package analyzer

import (
	"fmt"

	"github.com/crispy/focus-time-tracker/internal/common"
	"gonum.org/v1/plot/plotter"
)

// PreparePlotData: plotting에 필요한 데이터(카테고리별 점, 회귀선, 평가 텍스트, 워터마크) 생성
// - data: 여러 일자의 FocusData 배열
// 반환: 카테고리별 점(points), 회귀선(regressionLines), 평가 텍스트, 워터마크 문자열
func PreparePlotData(data []common.FocusData) (map[string]plotter.XYs, map[string]plotter.XYs, string, string) {
	points := map[string]plotter.XYs{} // 카테고리별 실제 점 데이터
	regressionLines := map[string]plotter.XYs{} // 카테고리별 회귀선 데이터
	for _, cat := range common.Categories {
		points[cat] = makeCategoryPoints(data, cat) // 실제 점 생성
		regressionLines[cat] = makeRegressionPoints(data, cat) // 회귀선 생성
	}
	evalText := makeEvalText(data) // 카테고리별 트렌드 평가 텍스트
	watermark := makeWatermark() // 워터마크(날짜/시간)
	return points, regressionLines, evalText, watermark
}

// PlotFocusTrendsAndRegression: 분석 및 시각화 전체 orchestration 함수
// - data: 여러 일자의 FocusData 배열
// 반환: PNG 이미지 []byte, 에러
// 1. plot용 데이터 준비(점, 회귀선, 텍스트)
// 2. plot.go의 DrawFocusTrends로 그림 생성
func PlotFocusTrendsAndRegression(data []common.FocusData) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("분석할 데이터가 없습니다")
	}

	// 1. 모든 데이터에서 등장한 카테고리 동적 추출
	categorySet := make(map[string]struct{})
	for _, d := range data {
		for cat := range d.Categories {
			categorySet[cat] = struct{}{}
		}
	}
	categories := make([]string, 0, len(categorySet))
	for cat := range categorySet {
		categories = append(categories, cat)
	}

	// 2. 정규화된 데이터 준비 (카테고리별 MaxScore[cat] > 0인 날만)
	normData := make([]common.FocusData, 0, len(data))
	for _, d := range data {
		norm := common.FocusData{
			Date:       d.Date,
			TotalFocus: d.TotalFocus,
			MaxScore:   d.MaxScore,
			Categories: map[string]int{},
			TimeSlots:  d.TimeSlots,
		}
		for k, v := range d.Categories {
			max, ok := d.MaxScore[k]
			if ok && max > 0 {
				norm.Categories[k] = int(float64(v) / float64(max) * 100.0)
			} else {
				norm.Categories[k] = 0
			}
		}
		normData = append(normData, norm)
	}

	// 3. points, regressionLines 동적 카테고리로 생성
	points := map[string]plotter.XYs{}
	regressionLines := map[string]plotter.XYs{}
	for _, cat := range categories {
		points[cat] = makeCategoryPoints(normData, cat)
		regressionLines[cat] = makeRegressionPoints(normData, cat)
	}
	evalText := makeEvalText(normData)
	watermark := makeWatermark()

	// 4. aggregateLine 계산: 동적 카테고리별로 모든 일자의 평균 (0점 제외), MaxScore로 비율화
	totalAverages := make([]float64, len(categories))
	totalFocus := 0.0
	for i, cat := range categories {
		sum := 0.0
		count := 0.0
		for _, d := range data {
			if d.Categories == nil || d.MaxScore == nil {
				continue
			}
			max, ok := d.MaxScore[cat]
			if !ok || max == 0 {
				continue
			}
			v, ok := d.Categories[cat]
			if !ok || v == 0 {
				continue
			}
			sum += float64(v) / float64(max) * 100.0
			count++
		}
		avg := 0.0
		if count > 0 {
			avg = sum / count
		}
		totalAverages[i] = avg
		totalFocus += avg
	}

	aggregateLine := make(plotter.XYs, len(categories))
	for i, avg := range totalAverages {
		ratio := 0.0
		if totalFocus > 0 {
			ratio = avg / totalFocus * 100.0
		}
		aggregateLine[i].X = float64(i)
		aggregateLine[i].Y = ratio
	}

	// 5. DrawFocusTrends에 동적 카테고리 전달
	return DrawFocusTrends(points, regressionLines, evalText, watermark, aggregateLine, normData, categories)
}

// PlotTimeSlotAverageFocusAggregatePNG: 전체 데이터를 합산하여 단일 평균 라인 그래프를 그림
// - data: 여러 일자의 FocusData 배열
// 반환: PNG 이미지 []byte, 에러
func PlotTimeSlotAverageFocusAggregatePNG(data []common.FocusData) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("분석할 데이터가 없습니다")
	}
	// 시간대별 합산 및 평균 계산 (0점 제외)
	timeSlotSum := map[string]int{}
	timeSlotCount := map[string]int{}
	for _, d := range data {
		for t, v := range d.TimeSlots {
			if v == 0 {
				continue
			}
			timeSlotSum[t] += v
			timeSlotCount[t]++
		}
	}
	agg := common.FocusData{
		Date:      "Aggregate",
		TimeSlots: map[string]int{},
	}
	for t, sum := range timeSlotSum {
		if timeSlotCount[t] > 0 {
			agg.TimeSlots[t] = int(float64(sum) / float64(timeSlotCount[t]) + 0.5) // 반올림
		}
	}
	
	return PlotTimeSlotAverageFocusPNG([]common.FocusData{agg})
}
