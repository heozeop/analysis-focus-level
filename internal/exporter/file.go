package exporter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/crispy/focus-time-tracker/internal/analyzer"
	"github.com/crispy/focus-time-tracker/internal/common"
)

// SaveJSON saves FocusData as JSON to the given path.
func SaveJSON(data common.FocusData, path string) error {
	log.Printf("[SaveJSON] 저장 경로: %s", path)
	if err := EnsureDir(filepath.Dir(path)); err != nil {
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

// EnsureDir: 디렉토리 생성 (없으면)
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// WriteFile: 파일 저장 (경로 자동 생성)
func WriteFile(path string, data []byte) error {
	if err := EnsureDir(filepath.Dir(path)); err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}

// EnsureGraphFile: 그래프 파일 저장 전 디렉토리만 생성
func EnsureGraphFile(path string) error {
	return EnsureDir(filepath.Dir(path))
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

func ReadFocusDataFile(path string) (common.FocusData, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return common.FocusData{}, err
	}
	var d common.FocusData
	if err := json.Unmarshal(b, &d); err != nil {
		return common.FocusData{}, err
	}
	return d, nil
}

// GenerateGraphFile: 동일 이미지를 여러 경로에 저장
func GenerateGraphFile(data []common.FocusData, paths ...string) error {
	b, err := analyzer.PlotFocusTrendsAndRegression(data)
	if err != nil {
		return err
	}
	for _, path := range paths {
		if err := EnsureGraphFile(path); err != nil {
			return err
		}
		if err := WriteFile(path, b); err != nil {
			return err
		}
	}
	return nil
}

// LoadRecentFocusData: 최근 N일치 FocusData를 로드
func LoadRecentFocusData(rawDir string, days int) ([]common.FocusData, error) {
	files, err := filepath.Glob(filepath.Join(rawDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("파일 glob 실패: %w", err)
	}
	if len(files) > days {
		files = files[len(files)-days:]
	}
	var allData []common.FocusData
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			log.Printf("[LoadRecentFocusData] 파일 읽기 실패: %s (%v)", f, err)
			continue
		}
		var d common.FocusData
		if err := json.Unmarshal(b, &d); err != nil {
			log.Printf("[LoadRecentFocusData] JSON 파싱 실패: %s (%v)", f, err)
			continue
		}
		allData = append(allData, d)
	}
	return allData, nil
} 