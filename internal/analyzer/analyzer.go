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
	points, regressionLines, evalText, watermark := PreparePlotData(data)
	return DrawFocusTrends(points, regressionLines, evalText, watermark)
}
