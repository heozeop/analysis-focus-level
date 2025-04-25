package exporter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/crispy/focus-time-tracker/internal/analyzer"
)

func TestExportToJSON(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "test.json")
	data := analyzer.FocusData{
		Date:       "2024-06-01",
		TotalFocus: 120,
		Categories: map[string]int{"업무": 60, "학습": 60, "취미": 0, "수면": 0, "이동": 0},
	}
	err := ExportToJSON(data, testPath)
	if err != nil {
		t.Fatalf("ExportToJSON failed: %v", err)
	}
	// 파일이 생성되었는지 확인
	f, err := os.Open(testPath)
	if err != nil {
		t.Fatalf("JSON file not created: %v", err)
	}
	defer f.Close()
	// JSON 내용 검증
	var got analyzer.FocusData
	if err := json.NewDecoder(f).Decode(&got); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}
	if got.Date != data.Date || got.TotalFocus != data.TotalFocus {
		t.Errorf("Exported data mismatch: got %+v, want %+v", got, data)
	}
	for k, v := range data.Categories {
		if got.Categories[k] != v {
			t.Errorf("Category %s mismatch: got %d, want %d", k, got.Categories[k], v)
		}
	}
} 