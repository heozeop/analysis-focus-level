package analyzer

import (
	"fmt"
	"log"
	"time"

	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

// FocusData: 1일치 집중도/라벨 데이터 구조
// (PRD 예시 JSON 구조와 일치)
type FocusData struct {
	Date        string            `json:"date"`
	TotalFocus  int               `json:"totalFocus"`
	Categories  map[string]int    `json:"categories"`
}

var defaultColors = plotutil.SoftColors

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

// 카테고리별 데이터 포인트 생성
func makeCategoryPoints(data []FocusData, category string) plotter.XYs {
	pts := make(plotter.XYs, len(data))
	for i, d := range data {
		pts[i].X = float64(i)
		pts[i].Y = float64(d.Categories[category])
	}
	return pts
}

// 카테고리별 회귀선 포인트 생성
func makeRegressionPoints(data []FocusData, category string) plotter.XYs {
	if len(data) == 1 {
		v := float64(data[0].Categories[category])
		return plotter.XYs{{X: 0, Y: v}, {X: 1, Y: v}}
	}
	slope, intercept := Regression(data, category)
	regPts := make(plotter.XYs, len(data))
	for i := range data {
		regPts[i].X = float64(i)
		regPts[i].Y = slope*float64(i) + intercept
	}
	return regPts
}

// 평가 텍스트 생성 (카테고리별 slope 해석)
func makeEvalText(data []FocusData) string {
	eval := ""
	for _, cat := range Categories {
		slope, _ := Regression(data, cat)
		trend := ""
		if slope > 1 {
			trend = "상승"
		} else if slope < -1 {
			trend = "감소"
		} else {
			trend = "유지"
		}
		eval += fmt.Sprintf("%s: %.2f (%s)  ", cat, slope, trend)
	}
	return eval
}

// 워터마크(오늘 날짜) 텍스트 생성
func makeWatermark() string {
	loc, err := time.LoadLocation("Asia/Seoul")
	if err == nil {
		return time.Now().In(loc).Format("2006-01-02 15:04:05")
	}
	return time.Now().Format("2006-01-02 15:04:05")
}

// plot 객체 생성 및 스타일 지정
func newFocusPlot() *plot.Plot {
	p := plot.New()
	p.Title.Text = "최근 집중도 추이 및 회귀분석"
	p.X.Label.Text = "날짜"
	p.Y.Label.Text = "점수"
	return p
}

// PlotFocusTrendsAndRegression: 여러 일자 데이터를 받아 카테고리별 추이+회귀선 그래프를 그리고 평가 텍스트를 포함해 PNG로 저장
// data: 여러 일자의 FocusData 배열
// outPath: 저장할 PNG 파일 경로
// - 각 카테고리별 점수 추이와 회귀선을 그려 시각화
// - 하단에 카테고리별 회귀 해석 텍스트, 워터마크(날짜) 추가
// - 분석할 데이터가 없으면 에러 반환
func PlotFocusTrendsAndRegression(data []FocusData, outPath string) error {
	log.Printf("[PlotFocusTrendsAndRegression] 데이터 개수: %d, 저장 경로: %s", len(data), outPath)
	if len(data) == 0 {
		log.Printf("[PlotFocusTrendsAndRegression] 분석할 데이터가 없습니다.")
		return fmt.Errorf("분석할 데이터가 없습니다.")
	}
	p := newFocusPlot()
	log.Printf("[PlotFocusTrendsAndRegression] plot 객체 생성 완료")
	for i, cat := range Categories {
		l, err := plotter.NewLine(makeCategoryPoints(data, cat))
		if err != nil {
			log.Printf("[PlotFocusTrendsAndRegression] 카테고리 %s 라인 생성 실패: %v", cat, err)
			return err
		}
		l.Color = defaultColors[i%len(defaultColors)]
		l.Width = vg.Points(2)
		p.Add(l)
		p.Legend.Add(cat, l)

		r, err := plotter.NewLine(makeRegressionPoints(data, cat))
		if err != nil {
			log.Printf("[PlotFocusTrendsAndRegression] 카테고리 %s 회귀선 생성 실패: %v", cat, err)
			return err
		}
		r.Color = defaultColors[i%len(defaultColors)]
		r.Dashes = []vg.Length{vg.Points(4), vg.Points(4)}
		r.Width = vg.Points(1)
		p.Add(r)
		log.Printf("[PlotFocusTrendsAndRegression] 카테고리 %s 라인/회귀선 추가 완료", cat)
	}
	if label, err := plotter.NewLabels(plotter.XYLabels{
		XYs:    []plotter.XY{{X: float64(len(data))/2 - 0.5, Y: -10}},
		Labels: []string{makeEvalText(data)},
	}); err == nil {
		p.Add(label)
		log.Printf("[PlotFocusTrendsAndRegression] 평가 텍스트 추가 완료")
	}
	if wm, err := plotter.NewLabels(plotter.XYLabels{
		XYs:    []plotter.XY{{X: float64(len(data))/2 - 0.5, Y: -20}},
		Labels: []string{makeWatermark()},
	}); err == nil {
		p.Add(wm)
		log.Printf("[PlotFocusTrendsAndRegression] 워터마크 추가 완료")
	}
	err := p.Save(8*vg.Inch, 4*vg.Inch, outPath)
	if err != nil {
		log.Printf("[PlotFocusTrendsAndRegression] 파일 저장 실패: %v", err)
		return err
	}
	log.Printf("[PlotFocusTrendsAndRegression] 그래프 저장 성공: %s", outPath)
	return nil
} 