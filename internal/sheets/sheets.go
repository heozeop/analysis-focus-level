package sheets

import (
	"context"
	"fmt"
	"time"

	"github.com/crispy/focus-time-tracker/internal/analyzer"
	"github.com/crispy/focus-time-tracker/internal/common"
	"github.com/crispy/focus-time-tracker/internal/config"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// NewService: Google Sheets API + Drive API 서비스 생성 (환경변수 GSHEETS_CREDENTIALS_JSON 사용)
// - ctx: context.Context
// 반환: sheets.Service, drive.Service, 에러
func NewService(ctx context.Context) (*sheets.Service, *drive.Service, error) {
	creds := config.Envs.GSheetsCredentialsJSON
	if creds == "" {
		return nil, nil, fmt.Errorf("환경변수 GSHEETS_CREDENTIALS_JSON이 비어 있습니다")
	}
	config, err := google.JWTConfigFromJSON([]byte(creds), sheets.SpreadsheetsScope, drive.DriveScope)
	if err != nil {
		return nil, nil, err
	}
	ts := config.TokenSource(ctx)
	sheetsSrv, err := sheets.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, nil, err
	}
	driveSrv, err := drive.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, nil, err
	}
	return sheetsSrv, driveSrv, nil
}

// FindSpreadsheetIDByYear: 폴더 내에서 연도별 규칙적 파일명으로 Google Sheets ID 검색
// - ctx: context.Context
// - driveSrv: Google Drive API 서비스
// - folderID: 폴더 ID
// - year: 연도
// 반환: 스프레드시트 ID, 에러
func FindSpreadsheetIDByYear(ctx context.Context, driveSrv *drive.Service, folderID string, year int) (string, error) {
	name := fmt.Sprintf("%d Focus Log", year)
	q := fmt.Sprintf("name = '%s' and '%s' in parents and mimeType = 'application/vnd.google-apps.spreadsheet'", name, folderID)
	files, err := driveSrv.Files.List().Q(q).Fields("files(id, name)").Do()
	if err != nil {
		return "", err
	}
	if len(files.Files) == 0 {
		return "", fmt.Errorf("해당 연도의 시트를 찾을 수 없음: %s", name)
	}
	return files.Files[0].Id, nil
}

// ExtractDailyFocusData: 특정 연/월/일의 시트 데이터(라벨, 집중도) 추출 및 FocusData 집계
// - sheetsSrv: Google Sheets API 서비스
// - spreadsheetID: 스프레드시트 ID
// - year, month, day: 연/월/일
// 반환: FocusData, 날짜 문자열(YYYY-MM-DD), 에러
func ExtractDailyFocusData(sheetsSrv *sheets.Service, spreadsheetID string, year, month, day int) (common.FocusData, string, error) {
	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		return common.FocusData{}, "", err
	}
	sheetName := fmt.Sprintf("%d월", month)
	dateCol := day // 1일=1, 2일=2, ...
	labels, scores, err := ParseDailyData(sheetsSrv, spreadsheetID, sheetName, dateCol)
	if err != nil {
		return common.FocusData{}, "", err
	}
	data := analyzer.AnalyzeFocus(labels, scores)
	// 날짜 문자열 생성 (YYYY-MM-DD, 한국시간)
	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc)
	dateStr := date.Format("2006-01-02")
	data.Date = dateStr
	return data, dateStr, nil
}
