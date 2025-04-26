package exporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/crispy/focus-time-tracker/internal/common"
)

func TestSaveJSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "data.json")
	data := common.FocusData{Date: "2024-06-02", TotalFocus: 100, Categories: map[string]int{"업무": 100}}
	err := SaveJSON(data, path)
	if err != nil {
		t.Fatalf("SaveJSON failed: %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("File not created: %v", err)
	}
	var got common.FocusData
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
	data := common.FocusData{Date: "2024-06-03", TotalFocus: 50, Categories: map[string]int{"업무": 50}}
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
	data := []common.FocusData{
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

func TestLoadRecentFocusData(t *testing.T) {
	tmpDir := t.TempDir()
	// 5일치 데이터 생성
	for i := 0; i < 5; i++ {
		path := filepath.Join(tmpDir, fmt.Sprintf("%d.json", i))
		data := common.FocusData{Date: fmt.Sprintf("2024-06-0%d", i+1), TotalFocus: i * 10, Categories: map[string]int{"업무": i * 10}}
		b, _ := json.Marshal(data)
		os.WriteFile(path, b, 0644)
	}
	// 최근 3개만 로드
	all, err := LoadRecentFocusData(tmpDir, 3)
	if err != nil {
		t.Fatalf("LoadRecentFocusData failed: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("Expected 3, got %d", len(all))
	}
	if all[0].Date != "2024-06-03" || all[2].Date != "2024-06-05" {
		t.Errorf("Wrong data order: %+v", all)
	}
} 