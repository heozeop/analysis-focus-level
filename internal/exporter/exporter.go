package exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/crispy/focus-time-tracker/internal/analyzer"
	"github.com/crispy/focus-time-tracker/internal/sheets"
	drivev3 "google.golang.org/api/drive/v3"
	sheetsv4 "google.golang.org/api/sheets/v4"
)

// ExportToJSON: FocusData를 JSON 파일로 저장 (assets/data/YYYY-MM-DD.json)
func ExportToJSON(data analyzer.FocusData, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(data)
}

// ExtractAndPush: 어제 날짜의 집중도 데이터를 추출해 JSON으로 저장하고, gitbook repo에 push한다.
func ExtractAndPush(ctx context.Context, sheetsSrv *sheetsv4.Service, driveSrv *drivev3.Service, folderID, repoPath string, now time.Time) error {
	// 한국 시간으로 변환
	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		return fmt.Errorf("Asia/Seoul 타임존 로드 실패: %v", err)
	}
	now = now.In(loc)
	yesterday := now.AddDate(0, 0, -1)
	year, month, day := yesterday.Date()

	spreadsheetID, err := sheets.FindSpreadsheetIDByYear(ctx, driveSrv, folderID, year)
	if err != nil {
		return fmt.Errorf("스프레드시트 ID 검색 실패: %v", err)
	}
	data, dateStr, err := sheets.ExtractDailyFocusData(sheetsSrv, spreadsheetID, year, int(month), day)
	if err != nil {
		return fmt.Errorf("시트 데이터 파싱 실패: %v", err)
	}

	// 1. JSON을 dailydata/raw/YYYY-MM-DD.json에 저장
	jsonRelPath := filepath.Join("dailydata", "raw", dateStr+".json")
	commitMsg := "자동 집중도 데이터: " + dateStr

	fmt.Println("[ExtractAndPush] repoPath:", repoPath)
	fmt.Println("[ExtractAndPush] jsonRelPath:", jsonRelPath)

	if err := ExportToJSON(data, jsonRelPath); err != nil {
		return fmt.Errorf("ExportToJSON 실패: %v", err)
	}

	// 2. 그래프 및 회귀분석 이미지 생성 (최근 7일 데이터)
	var allData []analyzer.FocusData
	files, err := filepath.Glob(filepath.Join("dailydata", "raw", "*.json"))
	if err == nil {
		// 최근 7일만 추출
		if len(files) > 7 {
			files = files[len(files)-7:]
		}
		for _, f := range files {
			b, err := os.ReadFile(f)
			if err != nil { continue }
			var d analyzer.FocusData
			if err := json.Unmarshal(b, &d); err != nil { continue }
			allData = append(allData, d)
		}
	}
	if len(allData) > 0 {
		// .gitbook/assets/graph.png (항상 같은 이름)
		graphGitbook := filepath.Join(repoPath, ".gitbook", "assets", "graph.png")
		// 파일이 없으면 빈 파일 생성
		if _, err := os.Stat(graphGitbook); os.IsNotExist(err) {
			os.MkdirAll(filepath.Dir(graphGitbook), 0755)
			os.WriteFile(graphGitbook, []byte{}, 0644)
		}
		analyzer.PlotFocusTrendsAndRegression(allData, graphGitbook)
		// dailydata/images/YYYY-MM-DD.png (일자별)
		imgDir := filepath.Join("dailydata", "images")
		os.MkdirAll(imgDir, 0755)
		graphDaily := filepath.Join(imgDir, dateStr+".png")
		analyzer.PlotFocusTrendsAndRegression(allData, graphDaily)
	}

	// 커밋/푸시 전에 main 브랜치로 checkout
	exec.Command("git", "-C", repoPath, "checkout", "main").Run()

	// 1. gitbook(submodule) push
	fmt.Println("[ExtractAndPush] === gitbook(submodule) push 시작 ===")
	gitbookCmds := [][]string{
		{"git", "-C", repoPath, "add", ".gitbook/assets/graph.png"},
		{"git", "-C", repoPath, "commit", "-m", commitMsg},
		{"git", "-C", repoPath, "push", "origin", "HEAD:main"},
	}
	for _, args := range gitbookCmds {
		cmd := exec.Command(args[0], args[1:]...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("[gitbook push 단계] %v: %s", err, string(out))
		}
	}
	fmt.Println("[ExtractAndPush] === gitbook(submodule) push 끝 ===")

	// 2. main push
	fmt.Println("[ExtractAndPush] === main push 시작 ===")
	mainCmds := [][]string{
		{"git", "add", filepath.Join("dailydata", "images", dateStr+".png")},
		{"git", "add", jsonRelPath},
		{"git", "commit", "-m", commitMsg},
		{"git", "push", "origin", "HEAD:main"},
	}
	for _, args := range mainCmds {
		cmd := exec.Command(args[0], args[1:]...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("[main push 단계] %v: %s", err, string(out))
		}
	}
	fmt.Println("[ExtractAndPush] === main push 끝 ===")

	return nil
}
