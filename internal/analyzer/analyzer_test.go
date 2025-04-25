package analyzer

import (
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