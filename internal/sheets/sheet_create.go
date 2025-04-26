package sheets

import (
	"fmt"

	"github.com/crispy/focus-time-tracker/internal/config"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/sheets/v4"
)

// createSpreadsheet: 연도별 Google Spreadsheet 생성 및 폴더 이동
// - sheetsSrv: Google Sheets API 서비스
// - driveSrv: Google Drive API 서비스
// - title: 시트 제목
// - year: 연도
// 반환: 생성된 스프레드시트 ID, 에러
func createSpreadsheet(sheetsSrv *sheets.Service, driveSrv *drive.Service, title string, year int) (string, error) {
	spreadsheet := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: fmt.Sprintf("%d Focus Log", year),
		},
	}
	sheetTitles := []string{"1월", "2월", "3월", "4월", "5월", "6월", "7월", "8월", "9월", "10월", "11월", "12월"}
	for _, name := range sheetTitles {
		spreadsheet.Sheets = append(spreadsheet.Sheets, &sheets.Sheet{
			Properties: &sheets.SheetProperties{Title: name},
		})
	}
	created, err := sheetsSrv.Spreadsheets.Create(spreadsheet).Do()
	if err != nil {
		return "", err
	}
	spreadsheetID := created.SpreadsheetId

	// 폴더 이동 처리
	parentFolderID := config.Envs.GSheetsParentFolderID
	if parentFolderID != "" {
		_, err := driveSrv.Files.Update(spreadsheetID, nil).
			AddParents(parentFolderID).
			Do()
		if err != nil {
			return spreadsheetID, fmt.Errorf("시트 생성 후 폴더 이동 실패: %v", err)
		}
	}
	return spreadsheetID, nil
}

// CreateYearlySheet: 연도별 시트 생성 후 월별 데이터 초기화 및 스타일 적용
// - sheetsSrv: Google Sheets API 서비스
// - driveSrv: Google Drive API 서비스
// - title: 시트 제목
// - year: 연도
// 반환: 생성된 스프레드시트 ID, 에러
func CreateYearlySheet(sheetsSrv *sheets.Service, driveSrv *drive.Service, title string, year int) (string, error) {
	// 1. 스프레드시트 생성 및 폴더 이동
	spreadsheetID, err := createSpreadsheet(sheetsSrv, driveSrv, title, year)
	if err != nil {
		return "", err
	}
	sheetTitles := []string{"1월", "2월", "3월", "4월", "5월", "6월", "7월", "8월", "9월", "10월", "11월", "12월"}
	for monthIdx, name := range sheetTitles {
		// 2. 월별 시트 데이터 초기화
		if err := initSheetData(sheetsSrv, spreadsheetID, name, year, monthIdx+1); err != nil {
			return spreadsheetID, err
		}
		// 3. 월별 시트 스타일/유효성/조건부서식 적용
		if err := applySheetStyles(sheetsSrv, spreadsheetID, name); err != nil {
			return spreadsheetID, err
		}
	}
	return spreadsheetID, nil
}
