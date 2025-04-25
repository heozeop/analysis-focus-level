package exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"log"

	"github.com/crispy/focus-time-tracker/internal/analyzer"
	"github.com/crispy/focus-time-tracker/internal/sheets"
	drivev3 "google.golang.org/api/drive/v3"
	sheetsv4 "google.golang.org/api/sheets/v4"
)

// SaveDailyJSON: 하루치 데이터를 JSON으로 저장
func SaveDailyJSON(data analyzer.FocusData, jsonRelPath string) error {
	return SaveJSON(data, jsonRelPath)
}

// LoadRecentFocusData: 최근 7일치 FocusData를 로드
func LoadRecentFocusData(rawDir string, days int) ([]analyzer.FocusData, error) {
	files, err := filepath.Glob(filepath.Join(rawDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("파일 glob 실패: %w", err)
	}
	if len(files) > days {
		files = files[len(files)-days:]
	}
	var allData []analyzer.FocusData
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			log.Printf("[LoadRecentFocusData] 파일 읽기 실패: %s (%v)", f, err)
			continue
		}
		var d analyzer.FocusData
		if err := json.Unmarshal(b, &d); err != nil {
			log.Printf("[LoadRecentFocusData] JSON 파싱 실패: %s (%v)", f, err)
			continue
		}
		allData = append(allData, d)
	}
	return allData, nil
}

// GenerateGraphs: 그래프 및 회귀분석 이미지 생성
func GenerateGraphs(allData []analyzer.FocusData, graphGitbook, graphDaily string) error {
	if len(allData) == 0 {
		return nil
	}
	if _, err := os.Stat(graphGitbook); os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(graphGitbook), 0755)
		os.WriteFile(graphGitbook, []byte{}, 0644)
	}
	log.Printf("[GenerateGraphs] 그래프 생성: %s", graphGitbook)
	analyzer.PlotFocusTrendsAndRegression(allData, graphGitbook)
	os.MkdirAll(filepath.Dir(graphDaily), 0755)
	log.Printf("[GenerateGraphs] 그래프 생성: %s", graphDaily)
	analyzer.PlotFocusTrendsAndRegression(allData, graphDaily)
	return nil
}

// PushGitbookAssets: gitbook repo에 그래프 push
func PushGitbookAssets(repoPath, commitMsg string) error {
	log.Println("[PushGitbookAssets] === gitbook(submodule) push 시작 ===")
	cmds := [][]string{
		{"-C", repoPath, "add", ".gitbook/assets/graph.png"},
		{"-C", repoPath, "commit", "-m", commitMsg},
		{"-C", repoPath, "push", "origin", "HEAD:main"},
	}
	for _, args := range cmds {
		if err := GitRun(args...); err != nil {
			return fmt.Errorf("[gitbook push 단계] %w", err)
		}
	}
	log.Println("[PushGitbookAssets] === gitbook(submodule) push 끝 ===")
	return nil
}

// PushMainAssets: main repo에 데이터/이미지 push
func PushMainAssets(dateStr, jsonRelPath, commitMsg string) error {
	log.Println("[PushMainAssets] === main push 시작 ===")
	cmds := [][]string{
		{"add", filepath.Join("dailydata", "images", dateStr+".png")},
		{"add", jsonRelPath},
		{"commit", "-m", commitMsg},
		{"push", "origin", "HEAD:main"},
	}
	for _, args := range cmds {
		if err := GitRun(args...); err != nil {
			return fmt.Errorf("[main push 단계] %w", err)
		}
	}
	log.Println("[PushMainAssets] === main push 끝 ===")
	return nil
}

// ExtractAndPush orchestrates the full export process
func ExtractAndPush(ctx context.Context, sheetsSrv *sheetsv4.Service, driveSrv *drivev3.Service, folderID, repoPath string, now time.Time) error {
	// 1. 한국 시간으로 변환
	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		return fmt.Errorf("Asia/Seoul 타임존 로드 실패: %w", err)
	}
	now = now.In(loc)

	// 2. 어제 날짜 계산
	yesterday := now.AddDate(0, 0, -1)
	year, month, day := yesterday.Date()

	// 3. Google Sheets에서 해당 연도 스프레드시트 ID 찾기
	spreadsheetID, err := sheets.FindSpreadsheetIDByYear(ctx, driveSrv, folderID, year)
	if err != nil {
		return fmt.Errorf("스프레드시트 ID 검색 실패: %w", err)
	}

	// 4. 어제 날짜의 집중도 데이터 추출
	data, dateStr, err := sheets.ExtractDailyFocusData(sheetsSrv, spreadsheetID, year, int(month), day)
	if err != nil {
		return fmt.Errorf("시트 데이터 파싱 실패: %w", err)
	}

	// 5. JSON 파일로 저장
	jsonRelPath := filepath.Join("dailydata", "raw", dateStr+".json")
	commitMsg := "자동 집중도 데이터: " + dateStr
	if err := SaveDailyJSON(data, jsonRelPath); err != nil {
		return err
	}

	// 6. 최근 7일치 데이터 로드
	allData, err := LoadRecentFocusData(filepath.Join("dailydata", "raw"), 7)
	if err != nil {
		return err
	}

	// 7. 그래프 이미지 생성 (gitbook, dailydata)
	graphGitbook := filepath.Join(repoPath, ".gitbook", "assets", "graph.png")
	graphDaily := filepath.Join("dailydata", "images", dateStr+".png")
	if err := GenerateGraphFile(allData, graphGitbook); err != nil {
		return err
	}
	if err := GenerateGraphFile(allData, graphDaily); err != nil {
		return err
	}

	// 8. gitbook repo main 브랜치로 checkout
	exec.Command("git", "-C", repoPath, "checkout", "main").Run()

	// 9. gitbook repo에 그래프 push
	if err := PushGitbookAssets(repoPath, commitMsg); err != nil {
		return err
	}

	// 10. main repo에 데이터/이미지 push
	if err := PushMainAssets(dateStr, jsonRelPath, commitMsg); err != nil {
		return err
	}
	return nil
}
