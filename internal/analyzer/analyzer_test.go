package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeFocus(t *testing.T) {
	labels := []string{"업무", "학습", "취미", "이동", "업무", "수면"}
	scores := []int{50, 30, 20, 40, 100, 0}
	result := AnalyzeFocus(labels, scores)

	if result.TotalFocus != 200 {
		t.Errorf("TotalFocus = %d, want 200", result.TotalFocus)
	}
	if result.Categories["업무"] != 150 {
		t.Errorf("업무 = %d, want 150", result.Categories["업무"])
	}
	if result.Categories["이동"] != 40 {
		t.Errorf("이동 = %d, want 40", result.Categories["이동"])
	}
}

func TestRegression(t *testing.T) {
	data := []FocusData{
		{Categories: map[string]int{"업무": 10}},
		{Categories: map[string]int{"업무": 20}},
		{Categories: map[string]int{"업무": 30}},
	}
	slope, intercept := Regression(data, "업무")
	if slope < 9.9 || slope > 10.1 {
		t.Errorf("slope = %f, want ~10", slope)
	}
	if intercept < 9.9 || intercept > 10.1 {
		t.Errorf("intercept = %f, want ~10", intercept)
	}
}

func TestMakeCategoryPoints(t *testing.T) {
	data := []FocusData{
		{Categories: map[string]int{"업무": 10}},
		{Categories: map[string]int{"업무": 20}},
	}
	pts := makeCategoryPoints(data, "업무")
	if len(pts) != 2 || pts[0].Y != 10 || pts[1].Y != 20 {
		t.Errorf("makeCategoryPoints 결과 이상: %+v", pts)
	}
}

func TestMakeRegressionPoints(t *testing.T) {
	data := []FocusData{
		{Categories: map[string]int{"업무": 10}},
		{Categories: map[string]int{"업무": 20}},
	}
	pts := makeRegressionPoints(data, "업무")
	if len(pts) != 2 {
		t.Errorf("makeRegressionPoints 길이 이상: %d", len(pts))
	}
	if pts[0].Y == pts[1].Y {
		t.Errorf("회귀선이 평행선만 나옴: %+v", pts)
	}
}

func TestMakeEvalText(t *testing.T) {
	data := []FocusData{
		{Categories: map[string]int{"업무": 10, "학습": 20, "취미": 0, "수면": 0, "이동": 0}},
		{Categories: map[string]int{"업무": 20, "학습": 10, "취미": 0, "수면": 0, "이동": 0}},
	}
	eval := makeEvalText(data)
	if len(eval) == 0 || eval == "" {
		t.Errorf("makeEvalText 결과 없음")
	}
}

func TestMakeWatermark(t *testing.T) {
	wm := makeWatermark()
	if len(wm) < 10 {
		t.Errorf("makeWatermark 결과 이상: %s", wm)
	}
}

func TestNewFocusPlot(t *testing.T) {
	p := newFocusPlot()
	if p == nil || p.Title.Text == "" {
		t.Errorf("newFocusPlot 생성 실패")
	}
}

func TestPlotFocusTrendsAndRegression(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "graph.png")
	data := []FocusData{
		{Date: "2024-06-01", Categories: map[string]int{"업무": 10, "학습": 20, "취미": 0, "수면": 0, "이동": 0}},
		{Date: "2024-06-02", Categories: map[string]int{"업무": 20, "학습": 10, "취미": 0, "수면": 0, "이동": 0}},
	}
	err := PlotFocusTrendsAndRegression(data, path)
	if err != nil {
		t.Fatalf("PlotFocusTrendsAndRegression 실패: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("그래프 파일 생성 안됨: %s", path)
	}
} 