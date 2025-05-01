package analyzer

import (
	"bytes"
	"fmt"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/crispy/focus-time-tracker/internal/common"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

// FindKoreanFont attempts to find a Korean font on the system
func FindKoreanFont() string {
	// Check if KOREAN_FONT_PATH is set in environment
	if fontPath := os.Getenv("KOREAN_FONT_PATH"); fontPath != "" {
		if _, err := os.Stat(fontPath); err == nil {
			return fontPath
		}
	}

	// Check the fonts directory first (relative to current directory)
	fontPaths := []string{
		"fonts/D2Coding/D2Coding-Ver1.3.2-20180524.ttf",
		"fonts/D2CodingLigature/D2Coding-Ver1.3.2-20180524-ligature.ttf",
	}
	
	// Get absolute path to project root
	cwd, err := os.Getwd()
	if err == nil {
		for _, path := range fontPaths {
			absPath := filepath.Join(cwd, path)
			if _, err := os.Stat(absPath); err == nil {
				fmt.Printf("Found Korean font: %s\n", absPath)
				return absPath
			}
		}
	}
	
	// If we didn't find any custom fonts, try system fonts
	systemFontPaths := []string{
		// Linux
	}
	
	// Try user's home directory
	home, err := os.UserHomeDir()
	if err == nil {
		// Linux user fonts
		systemFontPaths = append(systemFontPaths, filepath.Join(home, ".fonts/NanumGothic.ttf"))
	}
	
	for _, path := range systemFontPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	
	return ""
}

// makeCategoryPoints: 카테고리별 데이터 포인트 생성
// - data: 여러 일자의 FocusData 배열
// - category: 카테고리명
// 반환: plotter.XYs (x: 인덱스, y: 점수)
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

// makeRegressionPoints: 카테고리별 회귀선 포인트 생성
// - data: 여러 일자의 FocusData 배열
// - category: 카테고리명
// 반환: plotter.XYs (회귀선)
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

// makeEvalText: 평가 텍스트 생성 (카테고리별 slope 해석)
// - data: 여러 일자의 FocusData 배열
// 반환: 카테고리별 트렌드(상승/감소/유지) 텍스트
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

// makeWatermark: 워터마크(오늘 날짜/시간) 텍스트 생성
func makeWatermark() string {
	loc, err := time.LoadLocation("Asia/Seoul")
	if err == nil {
		return time.Now().In(loc).Format("2006-01-02 15:04:05")
	}
	return time.Now().Format("2006-01-02 15:04:05")
}

// PlotTimeSlotAverageFocusPNG: 시간대별 일자별 평균 몰입 점수 그래프를 PNG로 저장
// - data: 여러 일자의 FocusData 배열
// 반환: PNG 이미지 []byte, 에러
func PlotTimeSlotAverageFocusPNG(data []common.FocusData) ([]byte, error) {
	// Initialize Korean font
	if err := InitKoreanFont(); err != nil {
		fmt.Printf("Warning: failed to initialize Korean font: %v\n", err)
		fmt.Println("Korean characters may not display correctly.")
	}

	p := plot.New()
	p.Title.Text = "시간대별 일자별 평균 몰입 점수"
	// 제목과 라벨 사이에 여백 늘리기
	p.Title.Padding = vg.Points(10)
	
	p.X.Label.Text = "시간"
	p.Y.Label.Text = "평균 몰입 점수"
	
	// 라벨과 축 사이 여백 늘리기
	p.X.Label.Padding = vg.Points(10)
	p.Y.Label.Padding = vg.Points(10)
	
	// 축과 데이터 사이 여백 늘리기
	p.X.Padding = vg.Points(5)
	p.Y.Padding = vg.Points(5)
	
	// 폰트 경로 출력 (디버그용)
	fontPath := FindKoreanFont()
	if fontPath != "" {
		fmt.Printf("Using Korean font: %s\n", fontPath)
	} else {
		fmt.Println("Warning: Korean font not found. Korean characters may not display correctly.")
	}

	// X축 눈금 간격 설정
	p.X.Tick.Marker = plot.ConstantTicks([]plot.Tick{
		{Value: 0, Label: "0"}, {Value: 6, Label: "6"}, {Value: 12, Label: "12"}, {Value: 18, Label: "18"}, {Value: 24, Label: "24"},
	})
	p.X.LineStyle.Width = vg.Points(1)
	p.X.LineStyle.Color = color.Gray{Y: 128}
	
	// 눈금 레이블 크기 조정
	p.X.Tick.Label.Font.Size = vg.Points(10)
	p.Y.Tick.Label.Font.Size = vg.Points(10)

	// 워터마크 추가 (오른쪽 하단에 보이도록 위치 조정)
	labels, err := plotter.NewLabels(plotter.XYLabels{
		XYs:    []plotter.XY{{X: 23, Y: 5}},
		Labels: []string{makeWatermark()},
	})
	if err == nil && labels != nil && len(labels.Labels) > 0 && len(labels.XYs) > 0 {
		// 워터마크 글꼴 크기 조정
		labels.TextStyle[0].Font.Size = vg.Points(8)
		p.Add(labels)
	}

	colors := plotutil.SoftColors
	for idx, d := range data {
		// 시간대별 점수 집계 (0점 제외)
		timeSlotSum := map[string]int{}
		timeSlotCount := map[string]int{}
		for t, v := range d.TimeSlots {
			if v == 0 {
				continue
			}
			timeSlotSum[t] += v
			timeSlotCount[t]++
		}
		// 시간대 정렬
		times := make([]string, 0, len(timeSlotSum))
		for t := range timeSlotSum {
			times = append(times, t)
		}
		type timeSlot struct {
			h, m int
			s    string
		}
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
			avg := 0.0
			if timeSlotCount[obj.s] > 0 {
				avg = float64(timeSlotSum[obj.s]) / float64(timeSlotCount[obj.s])
			}
			pts[i].X = float64(obj.h) + float64(obj.m)/60.0
			pts[i].Y = avg
		}
		if len(pts) == 0 {
			continue
		}
		l, err := plotter.NewLine(pts)
		if err != nil {
			return nil, err
		}
		l.Color = colors[idx%len(colors)]
		l.Width = vg.Points(2)
		p.Add(l)
		p.Legend.Add(d.Date, l)
	}
	
	// 범례 설정
	p.Legend.Top = true
	p.Legend.Left = false
	p.Legend.XOffs = vg.Points(-10) 
	p.Legend.YOffs = vg.Points(10)
	p.Legend.Padding = vg.Points(8)
	p.Legend.TextStyle.Font.Size = vg.Points(10)
	p.Legend.ThumbnailWidth = vg.Points(30)
	
	p.X.Min = 0
	p.X.Max = 24
	p.Y.Min = 0
	
	buf := &bytes.Buffer{}
	// 이미지 크기를 키워서 여백을 더 확보
	w, err := p.WriterTo(vg.Points(1280), vg.Points(640), "png")
	if err != nil {
		return nil, err
	}
	_, err = w.WriteTo(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DrawFocusTrends: 준비된 데이터(points, regressionLines, evalText, watermark, aggregateLine, categories)로 그림만 그림
// - points: 카테고리별 실제 점 데이터
// - regressionLines: 카테고리별 회귀선 데이터
// - evalText: 평가 텍스트
// - watermark: 워터마크(날짜/시간)
// - aggregateLine: 전체 평균 라인 (없으면 nil)
// - data: 추가 데이터 배열
// - categories: 동적으로 추출된 카테고리 목록
// 반환: PNG 이미지 []byte, 에러
func DrawFocusTrends(points, regressionLines map[string]plotter.XYs, evalText, watermark string, aggregateLine plotter.XYs, data []common.FocusData, categories []string) ([]byte, error) {
	// Initialize Korean font
	if err := InitKoreanFont(); err != nil {
		fmt.Printf("Warning: failed to initialize Korean font: %v\n", err)
		fmt.Println("Korean characters may not display correctly.")
	}
	
	p := plot.New()
	p.Title.Text = "카테고리별 트렌드 및 회귀선"
	// 제목과 라벨 사이에 여백 늘리기
	p.Title.Padding = vg.Points(10)
	
	p.X.Label.Text = "일자"
	p.Y.Label.Text = "점수"
	
	// 라벨과 축 사이 여백 늘리기
	p.X.Label.Padding = vg.Points(10)
	p.Y.Label.Padding = vg.Points(10)
	
	// 축과 데이터 사이 여백 늘리기
	p.X.Padding = vg.Points(5)
	p.Y.Padding = vg.Points(5)
	
	// 폰트 경로 출력 (디버그용)
	fontPath := FindKoreanFont()
	if fontPath != "" {
		fmt.Printf("Using Korean font: %s\n", fontPath)
	} else {
		fmt.Println("Warning: Korean font not found. Korean characters may not display correctly.")
	}

	// 오늘 기준 ±6일 x축 생성
	today := time.Now().Truncate(24 * time.Hour)
	dates := make([]string, 0, 13)
	dateToX := map[string]float64{}
	for i := -6; i <= 6; i++ {
		d := today.AddDate(0, 0, i)
		dstr := d.Format("2006-01-02")
		dates = append(dates, dstr)
		dateToX[dstr] = float64(i + 6) // x=0~12
	}

	// x축 눈금 설정 - 레이블 간격 더 넓게 (반절만 표시)
	ticks := make([]plot.Tick, 0, 13)
	for i, d := range dates {
		// 표시 간격 넓히기 (짝수 인덱스만 레이블 표시)
		if i%2 == 0 {
			label := d
			if i == 6 {
				label += " (오늘)"
			}
			ticks = append(ticks, plot.Tick{Value: float64(i), Label: label})
		} else {
			// 레이블 없이 눈금만 표시
			ticks = append(ticks, plot.Tick{Value: float64(i), Label: ""})
		}
	}
	p.X.Tick.Marker = plot.ConstantTicks(ticks)
	
	// X축 레이블 회전 (가독성 향상)
	p.X.Tick.Label.Rotation = math.Pi / 6
	p.X.Tick.Label.YAlign = draw.YCenter
	p.X.Tick.Label.XAlign = draw.XRight
	
	// 레이블 글꼴 크기 조정
	p.X.Tick.Label.Font.Size = vg.Points(10)
	p.Y.Tick.Label.Font.Size = vg.Points(10)
	
	p.X.Min = 0
	p.X.Max = 12

	colors := plotutil.SoftColors
	colorIdx := 0
	for _, cat := range categories {
		pts := points[cat]
		// pts의 X를 실제 일자 기반으로 재설정 (data[i].Date 사용)
		newPts := make(plotter.XYs, 0, len(pts))
		for i, pt := range pts {
			if i >= len(data) {
				break
			}
			dateStr := data[i].Date
			if x, ok := dateToX[dateStr]; ok {
				// 미래(오늘 이후)는 점을 그리지 않음
				dateTime, _ := time.Parse("2006-01-02", dateStr)
				if !dateTime.After(today) {
					newPts = append(newPts, plotter.XY{X: x, Y: pt.Y})
				}
			}
		}
		if len(newPts) > 0 {
			l, err := plotter.NewLine(newPts)
			if err != nil {
				return nil, err
			}
			l.Color = colors[colorIdx%len(colors)]
			l.Width = vg.Points(2)
			p.Add(l)
			p.Legend.Add(cat+"(실제)", l)
		}

		if regPts, ok := regressionLines[cat]; ok {
			newRegPts := make(plotter.XYs, 0, len(regPts))
			for i, pt := range regPts {
				if i >= len(data) {
					break
				}
				dateStr := data[i].Date
				if x, ok := dateToX[dateStr]; ok {
					newRegPts = append(newRegPts, plotter.XY{X: x, Y: pt.Y})
				}
			}
			// 미래 구간(오늘 이후) 회귀선 예측 추가 (기울기 2배 반영)
			lastY := 0.0
			var lastDelta float64
			if len(regPts) > 1 {
				lastDelta = regPts[len(regPts)-1].Y - regPts[len(regPts)-2].Y
				lastY = regPts[len(regPts)-1].Y
			} else if len(regPts) > 0 {
				lastY = regPts[len(regPts)-1].Y
			}
			for i := 7; i < 13; i++ { // x=7~12: 미래
				dateStr := dates[i]
				if x, ok := dateToX[dateStr]; ok {
					// 마지막 기울기를 2배로 반영
					predY := lastY + lastDelta*2*float64(i-len(regPts)+1)
					newRegPts = append(newRegPts, plotter.XY{X: x, Y: predY})
				}
			}
			if len(newRegPts) > 0 {
				rl, err := plotter.NewLine(newRegPts)
				if err != nil {
					return nil, err
				}
				rl.Color = colors[colorIdx%len(colors)]
				rl.Dashes = []vg.Length{vg.Points(4), vg.Points(4)}
				rl.Width = vg.Points(2)
				p.Add(rl)
				p.Legend.Add(cat+"(회귀)", rl)
			}
		}
		colorIdx++
	}

	// aggregateLine이 있으면 굵은 검정색 선으로 항상 추가
	if aggregateLine != nil && len(aggregateLine) > 0 {
		newAgg := make(plotter.XYs, 0, len(aggregateLine))
		for i, pt := range aggregateLine {
			if i >= len(data) {
				break
			}
			dateStr := data[i].Date
			if x, ok := dateToX[dateStr]; ok {
				// 미래(오늘 이후)는 점을 그리지 않음
				dateTime, _ := time.Parse("2006-01-02", dateStr)
				if !dateTime.After(today) {
					newAgg = append(newAgg, plotter.XY{X: x, Y: pt.Y})
				}
			}
		}
		// 미래 구간(오늘 이후) 회귀선 예측: 이전 aggregate Y값의 평균 사용
		if len(newAgg) > 0 {
			var meanY float64
			for _, pt := range newAgg {
				meanY += pt.Y
			}
			meanY = meanY / float64(len(newAgg))
			for i := 7; i < 13; i++ { // x=7~12: 미래
				if x, ok := dateToX[dates[i]]; ok {
					newAgg = append(newAgg, plotter.XY{X: x, Y: meanY})
				}
			}
			aggLine, err := plotter.NewLine(newAgg)
			if err != nil {
				return nil, err
			}
			aggLine.Color = color.Black
			aggLine.Width = vg.Points(4)
			p.Add(aggLine)
			p.Legend.Add("전체 평균", aggLine)
		}
	}

	// 범례 설정
	p.Legend.Top = true
	p.Legend.XOffs = vg.Points(-10)
	p.Legend.YOffs = vg.Points(10)
	p.Legend.Padding = vg.Points(8)   // 범례 내부 여백
	p.Legend.TextStyle.Font.Size = vg.Points(10)
	p.Legend.ThumbnailWidth = vg.Points(30)  // 범례 썸네일 크기

	// 평가 텍스트 추가
	if evalText != "" {
		labels, err := plotter.NewLabels(plotter.XYLabels{
			XYs:    []plotter.XY{{X: 6, Y: -10}},
			Labels: []string{evalText},
		})
		if err == nil {
			labels.TextStyle[0].Font.Size = vg.Points(10)
			p.Add(labels)
		}
	}
	// 워터마크 추가
	if watermark != "" {
		labels, err := plotter.NewLabels(plotter.XYLabels{
			XYs:    []plotter.XY{{X: 6, Y: -20}},
			Labels: []string{watermark},
		})
		if err == nil {
			labels.TextStyle[0].Font.Size = vg.Points(8)
			p.Add(labels)
		}
	}

	// 색상 설명 추가 - 별도 라벨로 나누기
	legendDescs := []string{
		"빨간색(업무)",
		"파란색(학습)",
		"초록색(취미)",
		"보라색(수면)",
		"주황색(이동)",
		"검정(전체 평균)",
	}
	
	for i, desc := range legendDescs {
		posX := float64(i*2)  // 간격을 두고 배치
		labels, _ := plotter.NewLabels(plotter.XYLabels{
			XYs:    []plotter.XY{{X: posX, Y: 5}},
			Labels: []string{desc},
		})
		if labels != nil {
			labels.TextStyle[0].Font.Size = vg.Points(9)
			p.Add(labels)
		}
	}

	p.Y.Min = 0
	p.Y.Max = 100

	buf := &bytes.Buffer{}
	// 이미지 크기를 키워서 여백을 더 확보
	w, err := p.WriterTo(vg.Points(1280), vg.Points(640), "png")
	if err != nil {
		return nil, err
	}
	_, err = w.WriteTo(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
