package analyzer

import (
	"testing"

	"github.com/crispy/focus-time-tracker/internal/common"
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
	data := []common.FocusData{
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
	data := []common.FocusData{
		{Categories: map[string]int{"업무": 10}},
		{Categories: map[string]int{"업무": 20}},
	}
	pts := makeCategoryPoints(data, "업무")
	if len(pts) != 2 || pts[0].Y != 10 || pts[1].Y != 20 {
		t.Errorf("makeCategoryPoints 결과 이상: %+v", pts)
	}
}

func TestMakeRegressionPoints(t *testing.T) {
	data := []common.FocusData{
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
	data := []common.FocusData{
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

func TestPlotFocusTrendsAndRegression(t *testing.T) {
	data := []common.FocusData{
		{Date: "2024-06-01", Categories: map[string]int{"업무": 10, "학습": 20, "취미": 0, "수면": 0, "이동": 0}},
		{Date: "2024-06-02", Categories: map[string]int{"업무": 20, "학습": 10, "취미": 0, "수면": 0, "이동": 0}},
	}
	imgBytes, err := PlotFocusTrendsAndRegression(data)
	if err != nil {
		t.Fatalf("PlotFocusTrendsAndRegression 실패: %v", err)
	}
	if len(imgBytes) == 0 {
		t.Errorf("생성된 PNG 바이트가 비어 있음")
	}
	// 실제 파일로 저장해볼 수도 있음 (선택)
	// path := filepath.Join(tmpDir, "graph.png")
	// if err := os.WriteFile(path, imgBytes, 0644); err != nil {
	// 	t.Errorf("PNG 파일 저장 실패: %v", err)
	// }
}
 