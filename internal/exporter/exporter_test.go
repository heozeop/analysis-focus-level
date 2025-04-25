package exporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/crispy/focus-time-tracker/internal/analyzer"
)

func TestSaveDailyJSON(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "test.json")
	data := analyzer.FocusData{
		Date:       "2024-06-01",
		TotalFocus: 120,
		Categories: map[string]int{"업무": 60, "학습": 60, "취미": 0, "수면": 0, "이동": 0},
	}
	err := SaveDailyJSON(data, testPath)
	if err != nil {
		t.Fatalf("SaveDailyJSON failed: %v", err)
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

func TestSaveJSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "data.json")
	data := analyzer.FocusData{Date: "2024-06-02", TotalFocus: 100, Categories: map[string]int{"업무": 100}}
	err := SaveJSON(data, path)
	if err != nil {
		t.Fatalf("SaveJSON failed: %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("File not created: %v", err)
	}
	var got analyzer.FocusData
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.Date != data.Date {
		t.Errorf("Date mismatch: got %s, want %s", got.Date, data.Date)
	}
}

func TestEnsureDir(t *testing.T) {
	tmpDir := t.TempDir()
	dir := filepath.Join(tmpDir, "a", "b")
	if err := EnsureDir(dir); err != nil {
		t.Fatalf("EnsureDir failed: %v", err)
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("Directory not created: %s", dir)
	}
}

func TestWriteFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "file.txt")
	content := []byte("hello")
	if err := WriteFile(path, content); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil || string(b) != "hello" {
		t.Errorf("File content mismatch: got %s", string(b))
	}
}

func TestListRecentFiles(t *testing.T) {
	tmpDir := t.TempDir()
	for i := 0; i < 5; i++ {
		name := filepath.Join(tmpDir, fmt.Sprintf("%d.json", i))
		os.WriteFile(name, []byte("{}"), 0644)
	}
	files, err := ListRecentFiles(tmpDir, 3)
	if err != nil {
		t.Fatalf("ListRecentFiles failed: %v", err)
	}
	if len(files) != 3 {
		t.Errorf("Expected 3 files, got %d", len(files))
	}
}

func TestReadFocusDataFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "focus.json")
	data := analyzer.FocusData{Date: "2024-06-03", TotalFocus: 50, Categories: map[string]int{"업무": 50}}
	b, _ := json.Marshal(data)
	os.WriteFile(path, b, 0644)
	got, err := ReadFocusDataFile(path)
	if err != nil {
		t.Fatalf("ReadFocusDataFile failed: %v", err)
	}
	if got.Date != data.Date {
		t.Errorf("Date mismatch: got %s, want %s", got.Date, data.Date)
	}
}

func TestEnsureGraphFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "graph.png")
	if err := EnsureGraphFile(path); err != nil {
		t.Fatalf("EnsureGraphFile failed: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Graph file not created: %s", path)
	}
}

func TestGenerateGraphFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "graph.png")
	data := []analyzer.FocusData{
		{Date: "2024-06-01", Categories: map[string]int{"업무": 10, "학습": 20, "취미": 0, "수면": 0, "이동": 0}},
		{Date: "2024-06-02", Categories: map[string]int{"업무": 20, "학습": 10, "취미": 0, "수면": 0, "이동": 0}},
	}
	err := GenerateGraphFile(data, path)
	if err != nil {
		t.Fatalf("GenerateGraphFile failed: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Graph file not created: %s", path)
	}
}

func TestGitRun(t *testing.T) {
	err := GitRun("not-a-real-git-command")
	if err == nil {
		t.Errorf("Expected error for invalid git command, got nil")
	}
} 