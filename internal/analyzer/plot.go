package analyzer

import (
	"bytes"
	"fmt"
	"time"

	"github.com/crispy/focus-time-tracker/internal/common"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

// PlotFocusTrendsAndRegressionPNG: 이미 계산된 포인트/회귀선/텍스트를 받아서 그래프를 그림
func PlotFocusTrendsAndRegressionPNG(
	points map[string]plotter.XYs,
	regressionLines map[string]plotter.XYs,
	evalText string,
	watermark string,
) ([]byte, error) {
	p := plot.New()
	p.Title.Text = "최근 집중도 추이 및 회귀분석"
	p.X.Label.Text = "날짜"
	p.Y.Label.Text = "점수"

	cats := make([]string, 0, len(points))
	for cat := range points {
		cats = append(cats, cat)
	}
	for i, cat := range cats {
		l, err := plotter.NewLine(points[cat])
		if err != nil {
			return nil, err
		}
		l.Color = plotutil.SoftColors[i%len(plotutil.SoftColors)]
		l.Width = vg.Points(2)
		p.Add(l)
		p.Legend.Add(cat, l)

		r, err := plotter.NewLine(regressionLines[cat])
		if err != nil {
			return nil, err
		}
		r.Color = plotutil.SoftColors[i%len(plotutil.SoftColors)]
		r.Dashes = []vg.Length{vg.Points(4), vg.Points(4)}
		r.Width = vg.Points(1)
		p.Add(r)
	}
	if evalText != "" {
		if label, err := plotter.NewLabels(plotter.XYLabels{
			XYs:    []plotter.XY{{X: float64(len(points[cats[0]])/2) - 0.5, Y: -10}},
			Labels: []string{evalText},
		}); err == nil {
			p.Add(label)
		}
	}
	if watermark != "" {
		if wm, err := plotter.NewLabels(plotter.XYLabels{
			XYs:    []plotter.XY{{X: float64(len(points[cats[0]])/2) - 0.5, Y: -20}},
			Labels: []string{watermark},
		}); err == nil {
			p.Add(wm)
		}
	}
	buf := &bytes.Buffer{}
	w, err := p.WriterTo(8*vg.Inch, 4*vg.Inch, "png")
	if err != nil {
		return nil, err
	}
	_, err = w.WriteTo(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// 카테고리별 데이터 포인트 생성
func makeCategoryPoints(data []common.FocusData, category string) plotter.XYs {
	pts := make(plotter.XYs, len(data))
	for i, d := range data {
		pts[i].X = float64(i)
		pts[i].Y = float64(d.Categories[category])
	}
	return pts
}

// 카테고리별 회귀선 포인트 생성
func makeRegressionPoints(data []common.FocusData, category string) plotter.XYs {
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
func makeEvalText(data []common.FocusData) string {
	eval := ""
	for _, cat := range common.Categories {
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
