package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/crispy/focus-time-tracker/internal/config"
	"github.com/crispy/focus-time-tracker/internal/sheets"
)

func main() {
	var (
		spreadsheetID string
		year          int
		folderID      string
	)
	config.LoadEnv()
	currentYear := time.Now().In(time.FixedZone("KST", 9*60*60)).Year()
	flag.StringVar(&spreadsheetID, "id", "", "업그레이드할 Google Spreadsheet ID (없으면 자동 검색)")
	flag.IntVar(&year, "year", currentYear, "업그레이드할 연도 (기본: 올해)")
	flag.StringVar(&folderID, "folder", config.Envs.GSheetsParentFolderID, "Google Drive 폴더 ID (기본: config.Envs.GSheetsParentFolderID)")
	flag.Parse()

	ctx := context.Background()
	sheetsSrv, driveSrv, err := sheets.NewService(ctx)
	if err != nil {
		fmt.Printf("Google Sheets API 인증 실패: %v\n", err)
		os.Exit(1)
	}

	if spreadsheetID == "" {
		if folderID == "" {
			fmt.Println("Google Drive 폴더 ID가 필요합니다. -folder 플래그 또는 환경변수 GSHEETS_FOLDER_ID를 설정하세요.")
			os.Exit(1)
		}
		spreadsheetID, err = sheets.FindSpreadsheetIDByYear(ctx, driveSrv, folderID, year)
		if err != nil {
			fmt.Printf("스프레드시트 자동 검색 실패: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("자동으로 찾은 스프레드시트 ID: %s\n", spreadsheetID)
	}

	// 오늘 날짜(한국시간) 기준으로 내일부터 연말까지 업그레이드
	loc := time.FixedZone("KST", 9*60*60)
	today := time.Now().In(loc)
	nextDay := today.AddDate(0, 0, 1)
	startMonth := int(nextDay.Month())
	startDay := nextDay.Day()

	fmt.Printf("[%d년] %d월 %d일부터 연말까지 스프레드시트(%s) 업그레이드 시작...\n", year, startMonth, startDay, spreadsheetID)
	err = sheets.UpgradeSheetToNewFormatFrom(sheetsSrv, spreadsheetID, year, startMonth, startDay)
	if err != nil {
		fmt.Printf("업그레이드 실패: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("업그레이드 완료!")
}
