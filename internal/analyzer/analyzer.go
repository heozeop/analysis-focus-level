package analyzer

import (
	"fmt"
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

// PlotFocusTrendsAndRegression: 여러 일자 데이터를 받아 카테고리별 추이+회귀선 그래프를 그리고 평가 텍스트를 포함해 PNG로 저장
func PlotFocusTrendsAndRegression(data []FocusData, outPath string) error {
	if len(data) == 0 {
		return fmt.Errorf("분석할 데이터가 없습니다.")
	}
	p := plot.New()
	p.Title.Text = "최근 집중도 추이 및 회귀분석"
	p.X.Label.Text = "날짜"
	p.Y.Label.Text = "점수"

	categories := []string{"업무", "학습", "취미", "수면", "이동"}
	colors := plotutil.SoftColors
	for i, cat := range categories {
		pts := make(plotter.XYs, len(data))
		for j, d := range data {
			pts[j].X = float64(j)
			pts[j].Y = float64(d.Categories[cat])
		}
		l, err := plotter.NewLine(pts)
		if err != nil {
			return err
		}
		l.Color = colors[i%len(colors)]
		l.Width = vg.Points(2)
		p.Add(l)
		p.Legend.Add(cat, l)

		// 회귀선
		var regPts plotter.XYs
		if len(data) == 1 {
			// 데이터가 1개면 x축에 평행한 선
			regPts = plotter.XYs{{X: 0, Y: pts[0].Y}, {X: 1, Y: pts[0].Y}}
		} else {
			slope, intercept := Regression(data, cat)
			regPts = make(plotter.XYs, len(data))
			for j := range data {
				regPts[j].X = float64(j)
				regPts[j].Y = slope*float64(j) + intercept
			}
		}
		r, err := plotter.NewLine(regPts)
		if err != nil {
			return err
		}
		r.Color = colors[i%len(colors)]
		r.Dashes = []vg.Length{vg.Points(4), vg.Points(4)}
		r.Width = vg.Points(1)
		p.Add(r)
	}

	// 평가 텍스트(예: slope 해석)
	eval := ""
	for _, cat := range categories {
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
	// plotter.Text로 그래프 하단에 텍스트 추가 (평가)
	label, err := plotter.NewLabels(plotter.XYLabels{
		XYs:    []plotter.XY{{X: float64(len(data))/2 - 0.5, Y: -10}},
		Labels: []string{eval},
	})
	if err == nil {
		p.Add(label)
	}

	// 오늘 날짜 워터마크 추가
	now := time.Now().Format("2006-01-02 15:04:05")
	wm, err := plotter.NewLabels(plotter.XYLabels{
		XYs:    []plotter.XY{{X: float64(len(data))/2 - 0.5, Y: -20}},
		Labels: []string{now},
	})
	if err == nil {
		p.Add(wm)
	}

	if err := p.Save(8*vg.Inch, 4*vg.Inch, outPath); err != nil {
		return err
	}
	return nil
} 