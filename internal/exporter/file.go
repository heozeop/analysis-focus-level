package exporter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/crispy/focus-time-tracker/internal/analyzer"
)

// SaveJSON saves FocusData as JSON to the given path.
func SaveJSON(data analyzer.FocusData, path string) error {
	log.Printf("[SaveJSON] 저장 경로: %s", path)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("디렉토리 생성 실패: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("파일 생성 실패: %w", err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(data); err != nil {
		return fmt.Errorf("JSON 인코딩 실패: %w", err)
	}
	log.Printf("[SaveJSON] 저장 성공")
	return nil
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func WriteFile(path string, data []byte) error {
	return ioutil.WriteFile(path, data, 0644)
}

func ListRecentFiles(dir string, n int) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}
	if len(files) > n {
		files = files[len(files)-n:]
	}
	return files, nil
}

func ReadFocusDataFile(path string) (analyzer.FocusData, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return analyzer.FocusData{}, err
	}
	var d analyzer.FocusData
	if err := json.Unmarshal(b, &d); err != nil {
		return analyzer.FocusData{}, err
	}
	return d, nil
}

func EnsureGraphFile(path string) error {
	dir := filepath.Dir(path)
	if err := EnsureDir(dir); err != nil {
		return err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return WriteFile(path, []byte{})
	}
	return nil
}

func GenerateGraphFile(data []analyzer.FocusData, path string) error {
	if err := EnsureGraphFile(path); err != nil {
		return err
	}
	return analyzer.PlotFocusTrendsAndRegression(data, path)
} 