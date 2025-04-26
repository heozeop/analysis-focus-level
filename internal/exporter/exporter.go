package exporter

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/crispy/focus-time-tracker/internal/sheets"
	drivev3 "google.golang.org/api/drive/v3"
	sheetsv4 "google.golang.org/api/sheets/v4"
)

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
	if err := SaveJSON(data, jsonRelPath); err != nil {
		return err
	}

	// 6. 최근 7일치 데이터 로드
	allData, err := LoadRecentFocusData(filepath.Join("dailydata", "raw"), 7)
	if err != nil {
		return err
	}

	// 7. 그래프 이미지 생성 (gitbook, dailydata)
	if len(allData) > 0 {
		graphGitbook := filepath.Join(repoPath, ".gitbook", "assets", "graph.png")
		graphDaily := filepath.Join("dailydata", "images", dateStr+".png")
		if err := GenerateGraphFile(allData, graphGitbook, graphDaily); err != nil {
			return err
		}
		// 일자별 시간대별 몰입 그래프 저장
		timeslotGitbook := filepath.Join(repoPath, ".gitbook", "assets", "timeslot-images.png")
		timeslotDaily := filepath.Join("dailydata", "timeslot-images", dateStr+".png")
		if err := SaveTimeSlotGraphs(allData, timeslotGitbook, timeslotDaily); err != nil {
			return err
		}
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
