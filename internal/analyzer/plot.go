package analyzer

import (
	"bytes"
	"fmt"
	"image/color"
	"sort"
	"time"

	"github.com/crispy/focus-time-tracker/internal/common"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

// 카테고리별 데이터 포인트 생성
func makeCategoryPoints(data []common.FocusData, category string) plotter.XYs {
	pts := make(plotter.XYs, len(data))
	for i, d := range data {
		val := 0
		if d.Categories != nil {
			if v, ok := d.Categories[category]; ok {
				val = v
			}
		}
		pts[i].X = float64(i)
		pts[i].Y = float64(val)
	}
	return pts
}

// 카테고리별 회귀선 포인트 생성
func makeRegressionPoints(data []common.FocusData, category string) plotter.XYs {
	if len(data) == 1 {
		v := 0.0
		if data[0].Categories != nil {
			if vv, ok := data[0].Categories[category]; ok {
				v = float64(vv)
			}
		}
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

// PlotTimeSlotAverageFocusPNG: 시간대별 평균 몰입 점수 그래프를 PNG로 저장
func PlotTimeSlotAverageFocusPNG(data []common.FocusData) ([]byte, error) {
	// 1. 모든 시간대 추출
	timeSlotSum := map[string]int{}
	timeSlotCount := map[string]int{}
	for _, d := range data {
		for t, v := range d.TimeSlots {
			timeSlotSum[t] += v
			timeSlotCount[t]++
		}
	}
	// 2. 시간대 정렬
	times := make([]string, 0, len(timeSlotSum))
	for t := range timeSlotSum {
		times = append(times, t)
	}
	// 시간 오름차순 정렬
	type timeSlot struct{ h, m int; s string }
	timeObjs := make([]timeSlot, 0, len(times))
	for _, t := range times {
		var h, m int
		fmt.Sscanf(t, "%02d:%02d", &h, &m)
		timeObjs = append(timeObjs, timeSlot{h, m, t})
	}
	sort.Slice(timeObjs, func(i, j int) bool {
		if timeObjs[i].h == timeObjs[j].h {
			return timeObjs[i].m < timeObjs[j].m
		}
		return timeObjs[i].h < timeObjs[j].h
	})
	// 3. plotter.XYs 생성
	pts := make(plotter.XYs, len(timeObjs))
	for i, obj := range timeObjs {
		avg := 0.0
		if timeSlotCount[obj.s] > 0 {
			avg = float64(timeSlotSum[obj.s]) / float64(timeSlotCount[obj.s])
		}
		pts[i].X = float64(obj.h) + float64(obj.m)/60.0
		pts[i].Y = avg
	}
	// 4. 그래프 생성
	p := plot.New()
	p.Title.Text = "시간대별 평균 몰입 점수"
	p.X.Label.Text = "시간"
	p.Y.Label.Text = "평균 몰입 점수"

	// x축 눈금/선 명시적 추가
	p.X.Tick.Marker = plot.ConstantTicks([]plot.Tick{
		{Value: 0, Label: "0"}, {Value: 6, Label: "6"}, {Value: 12, Label: "12"}, {Value: 18, Label: "18"}, {Value: 24, Label: "24"},
	})
	p.X.LineStyle.Width = vg.Points(1)
	p.X.LineStyle.Color = color.Gray{Y: 128}

	// 워터마크 추가
	if wm, err := plotter.NewLabels(plotter.XYLabels{
		XYs:    []plotter.XY{{X: 12, Y: p.Y.Min - 5}},
		Labels: []string{makeWatermark()},
	}); err == nil {
		p.Add(wm)
	}

	l, err := plotter.NewLine(pts)
	if err != nil {
		return nil, err
	}
	l.Color = plotutil.SoftColors[0]
	l.Width = vg.Points(2)
	p.Add(l)
	p.X.Min = 0
	p.X.Max = 24
	p.Y.Min = 0
	buf := &bytes.Buffer{}
	w, err := p.WriterTo(vg.Points(1024), vg.Points(512), "png")
	if err != nil {
		return nil, err
	}
	_, err = w.WriteTo(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// PlotTimeSlotAverageFocusPNGPerDay: 일자별 시간대별 평균 몰입 점수 그래프를 PNG로 저장
func PlotTimeSlotAverageFocusPNGPerDay(data []common.FocusData) (map[string][]byte, error) {
	result := make(map[string][]byte)
	for _, d := range data {
		// 1. 시간대별 점수 추출 (해당 일자만)
		times := make([]string, 0, len(d.TimeSlots))
		for t := range d.TimeSlots {
			times = append(times, t)
		}
		type timeSlot struct{ h, m int; s string }
		timeObjs := make([]timeSlot, 0, len(times))
		for _, t := range times {
			var h, m int
			fmt.Sscanf(t, "%02d:%02d", &h, &m)
			timeObjs = append(timeObjs, timeSlot{h, m, t})
		}
		sort.Slice(timeObjs, func(i, j int) bool {
			if timeObjs[i].h == timeObjs[j].h {
				return timeObjs[i].m < timeObjs[j].m
			}
			return timeObjs[i].h < timeObjs[j].h
		})
		pts := make(plotter.XYs, len(timeObjs))
		for i, obj := range timeObjs {
			pts[i].X = float64(obj.h) + float64(obj.m)/60.0
			pts[i].Y = float64(d.TimeSlots[obj.s])
		}
		p := plot.New()
		p.Title.Text = d.Date + " 시간대별 몰입 점수"
		p.X.Label.Text = "시간"
		p.Y.Label.Text = "몰입 점수"

		// x축 눈금/선 명시적 추가
		p.X.Tick.Marker = plot.ConstantTicks([]plot.Tick{
			{Value: 0, Label: "0"}, {Value: 6, Label: "6"}, {Value: 12, Label: "12"}, {Value: 18, Label: "18"}, {Value: 24, Label: "24"},
		})
		p.X.LineStyle.Width = vg.Points(1)
		p.X.LineStyle.Color = color.Gray{Y: 128}

		// 워터마크 추가
		if wm, err := plotter.NewLabels(plotter.XYLabels{
			XYs:    []plotter.XY{{X: 12, Y: p.Y.Min - 5}},
			Labels: []string{makeWatermark()},
		}); err == nil {
			p.Add(wm)
		}

		l, err := plotter.NewLine(pts)
		if err != nil {
			return nil, err
		}
		l.Color = plotutil.SoftColors[0]
		l.Width = vg.Points(2)
		p.Add(l)
		p.X.Min = 0
		p.X.Max = 24
		p.Y.Min = 0
		buf := &bytes.Buffer{}
		w, err := p.WriterTo(vg.Points(1024), vg.Points(512), "png")
		if err != nil {
			return nil, err
		}
		_, err = w.WriteTo(buf)
		if err != nil {
			return nil, err
		}
		result[d.Date] = buf.Bytes()
	}
	return result, nil
}

// PlotDailyTotalTrendAndRegressionPNG: x축=일자, y축=일별 총점(하루의 TotalFocus), 1주는 실제, 2주는 regression 예측
func PlotDailyTotalTrendAndRegressionPNG(data []common.FocusData) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("데이터가 1개 이상 필요합니다.")
	}
	// 1. 날짜 오름차순 정렬
	sorted := make([]common.FocusData, len(data))
	copy(sorted, data)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Date < sorted[j].Date })

	if len(sorted) == 1 {
		// 데이터가 1개일 때: 점 하나만 찍힌 그래프
		p := plot.New()
		p.Title.Text = "일별 총점 트렌드 (1개 데이터)"
		p.X.Label.Text = "일자"
		p.Y.Label.Text = "총점"
		pts := make(plotter.XYs, 1)
		pts[0].X = 0
		pts[0].Y = float64(sorted[0].TotalFocus)
		// x축 눈금: 날짜 1개
		p.X.Tick.Marker = plot.ConstantTicks([]plot.Tick{{Value: 0, Label: sorted[0].Date}})
		p.X.LineStyle.Width = vg.Points(1)
		p.X.LineStyle.Color = color.Gray{Y: 128}
		// 점 추가
		scatter, err := plotter.NewScatter(pts)
		if err != nil {
			return nil, err
		}
		scatter.GlyphStyle.Color = plotutil.SoftColors[0]
		scatter.GlyphStyle.Radius = vg.Points(6)
		p.Add(scatter)
		p.Legend.Add("실제", scatter)
		// 워터마크
		if wm, err := plotter.NewLabels(plotter.XYLabels{
			XYs:    []plotter.XY{{X: 0, Y: pts[0].Y - 10}},
			Labels: []string{makeWatermark()},
		}); err == nil {
			p.Add(wm)
		}
		buf := &bytes.Buffer{}
		w, err := p.WriterTo(vg.Points(1024), vg.Points(512), "png")
		if err != nil {
			return nil, err
		}
		_, err = w.WriteTo(buf)
		if err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}

	// 2. 최근 7일(1주)만 실제 데이터로 사용
	var weekData []common.FocusData
	if len(sorted) > 7 {
		weekData = sorted[len(sorted)-7:]
	} else {
		weekData = sorted
	}

	// 3. x축: 날짜(문자열), y축: TotalFocus
	pts := make(plotter.XYs, len(weekData))
	for i, d := range weekData {
		pts[i].X = float64(i)
		pts[i].Y = float64(d.TotalFocus)
	}

	// 4. 회귀선으로 2주(14일) 예측
	xs := make([]float64, len(weekData))
	ys := make([]float64, len(weekData))
	for i, d := range weekData {
		xs[i] = float64(i)
		ys[i] = float64(d.TotalFocus)
	}
	slope, intercept := 0.0, 0.0
	if len(xs) >= 2 {
		slope, intercept = RegressionTotal(xs, ys)
	}
	regPts := make(plotter.XYs, 14)
	for i := 0; i < 14; i++ {
		regPts[i].X = float64(i)
		regPts[i].Y = slope*float64(i) + intercept
	}

	// 5. 그래프 생성
	p := plot.New()
	p.Title.Text = "일별 총점 트렌드 (1주 실제, 1주 예측)"
	p.X.Label.Text = "일자"
	p.Y.Label.Text = "총점"

	// x축 눈금: 실제 날짜 라벨 + 예측 구간은 +N일
	xticks := []plot.Tick{}
	for i, d := range weekData {
		xticks = append(xticks, plot.Tick{Value: float64(i), Label: d.Date})
	}
	for i := len(weekData); i < 14; i++ {
		xticks = append(xticks, plot.Tick{Value: float64(i), Label: fmt.Sprintf("+%d일", i-len(weekData)+1)})
	}
	p.X.Tick.Marker = plot.ConstantTicks(xticks)
	p.X.LineStyle.Width = vg.Points(1)
	p.X.LineStyle.Color = color.Gray{Y: 128}

	// 실제 데이터 라인
	l, err := plotter.NewLine(pts)
	if err != nil {
		return nil, err
	}
	l.Color = plotutil.SoftColors[0]
	l.Width = vg.Points(2)
	p.Add(l)
	p.Legend.Add("실제", l)

	// 예측(회귀) 라인
	rl, err := plotter.NewLine(regPts)
	if err != nil {
		return nil, err
	}
	rl.Color = plotutil.SoftColors[1]
	rl.Dashes = []vg.Length{vg.Points(4), vg.Points(4)}
	rl.Width = vg.Points(2)
	p.Add(rl)
	p.Legend.Add("예측", rl)

	// 워터마크
	if wm, err := plotter.NewLabels(plotter.XYLabels{
		XYs:    []plotter.XY{{X: 3, Y: -20}},
		Labels: []string{makeWatermark()},
	}); err == nil {
		p.Add(wm)
	}

	buf := &bytes.Buffer{}
	w, err := p.WriterTo(vg.Points(1024), vg.Points(512), "png")
	if err != nil {
		return nil, err
	}
	_, err = w.WriteTo(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RegressionTotal: 단순 선형회귀 (gonum/stat 없이)
func RegressionTotal(xs, ys []float64) (slope, intercept float64) {
	n := float64(len(xs))
	if n < 2 {
		return 0, 0
	}
	sumX, sumY, sumXY, sumXX := 0.0, 0.0, 0.0, 0.0
	for i := 0; i < int(n); i++ {
		sumX += xs[i]
		sumY += ys[i]
		sumXY += xs[i] * ys[i]
		sumXX += xs[i] * xs[i]
	}
	denom := n*sumXX - sumX*sumX
	if denom == 0 {
		return 0, 0
	}
	slope = (n*sumXY - sumX*sumY) / denom
	intercept = (sumY - slope*sumX) / n
	return
}
