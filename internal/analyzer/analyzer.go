package analyzer

import (
	"fmt"

	"github.com/crispy/focus-time-tracker/internal/common"
	"gonum.org/v1/plot/plotter"
)

// PreparePlotData: plotting에 필요한 데이터(카테고리별 점, 회귀선, 평가 텍스트, 워터마크) 생성
func PreparePlotData(data []common.FocusData) (map[string]plotter.XYs, map[string]plotter.XYs, string, string) {
	points := map[string]plotter.XYs{}
	regressionLines := map[string]plotter.XYs{}
	for _, cat := range common.Categories {
		points[cat] = makeCategoryPoints(data, cat)
		regressionLines[cat] = makeRegressionPoints(data, cat)
	}
	evalText := makeEvalText(data)
	watermark := makeWatermark()
	return points, regressionLines, evalText, watermark
}

// PlotFocusTrendsAndRegression: orchestration 함수. 분석 및 시각화 전체 수행
// data: 여러 일자의 FocusData 배열
// 반환: PNG 이미지 []byte, 에러
func PlotFocusTrendsAndRegression(data []common.FocusData) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("분석할 데이터가 없습니다.")
	}
	return PlotDailyTotalTrendAndRegressionPNG(data)
} 