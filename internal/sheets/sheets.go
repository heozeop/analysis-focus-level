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

// SheetsAPI: Google Sheets API 래퍼 인터페이스 (테스트/Mock 용)
type SheetsAPI interface {
	GetValues(spreadsheetID, readRange string) ([][]interface{}, error)
}

// DriveAPI: Google Drive API 래퍼 인터페이스 (테스트/Mock 용)
type DriveAPI interface {
	FindFiles(ctx context.Context, query string) ([]*drive.File, error)
}

// 실제 구현체: Google Sheets API
// (테스트에서는 gomock/mockgen으로 대체)
type RealSheetsAPI struct{ srv *sheets.Service }
func (r *RealSheetsAPI) GetValues(spreadsheetID, readRange string) ([][]interface{}, error) {
	resp, err := r.srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return nil, err
	}
	return resp.Values, nil
}

type RealDriveAPI struct{ srv *drive.Service }
func (r *RealDriveAPI) FindFiles(ctx context.Context, query string) ([]*drive.File, error) {
	files, err := r.srv.Files.List().Q(query).Fields("files(id, name)").Do()
	if err != nil {
		return nil, err
	}
	return files.Files, nil
}

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

// FindSpreadsheetIDByYearAPI: 폴더 내에서 연도별 규칙적 파일명으로 Google Sheets ID 검색 (mockable)
// - driveAPI: DriveAPI 인터페이스
// - ctx, folderID, year: 기존과 동일
// 반환: 스프레드시트 ID, 에러
func FindSpreadsheetIDByYearAPI(ctx context.Context, driveAPI DriveAPI, folderID string, year int) (string, error) {
	name := fmt.Sprintf("%d Focus Log", year)
	q := fmt.Sprintf("name = '%s' and '%s' in parents and mimeType = 'application/vnd.google-apps.spreadsheet'", name, folderID)
	files, err := driveAPI.FindFiles(ctx, q)
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", fmt.Errorf("해당 연도의 시트를 찾을 수 없음: %s", name)
	}
	return files[0].Id, nil
}

// ExtractDailyFocusDataAPI: 특정 연/월/일의 시트 데이터(라벨, 집중도) 추출 및 FocusData 집계 (mockable)
// - sheetsAPI: SheetsAPI 인터페이스
// - spreadsheetID, year, month, day: 기존과 동일
// 반환: FocusData, 날짜 문자열(YYYY-MM-DD), 에러
func ExtractDailyFocusDataAPI(sheetsAPI SheetsAPI, spreadsheetID string, year, month, day int) (common.FocusData, string, error) {
	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		return common.FocusData{}, "", err
	}
	sheetName := fmt.Sprintf("%d월", month)
	dateCol := day // 1일=1, 2일=2, ...
	startCol := 2 + (dateCol-1)*2
	endCol := startCol + 1
	startColName := colIdxToName(startCol)
	endColName := colIdxToName(endCol)
	rangeStr := fmt.Sprintf("'%s'!%s2:%s145", sheetName, startColName, endColName)
	values, err := sheetsAPI.GetValues(spreadsheetID, rangeStr)
	if err != nil {
		return common.FocusData{}, "", err
	}
	labels := []string{}
	scores := []int{}
	for i := 0; i < 144; i++ {
		row := []interface{}{}
		if i < len(values) {
			row = values[i]
		}
		label := ""
		score := 0
		if len(row) > 0 {
			label = fmt.Sprintf("%v", row[0])
		}
		if len(row) > 1 {
			if v, err := fmt.Sscanf(fmt.Sprintf("%v", row[1]), "%d", &score); v == 1 && err == nil {
				// score already set
			} else {
				score = 0
			}
		}
		if label != "" {
			labels = append(labels, label)
			scores = append(scores, score)
		}
	}
	data := analyzer.AnalyzeFocus(labels, scores)
	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc)
	dateStr := date.Format("2006-01-02")
	data.Date = dateStr
	return data, dateStr, nil
}

// (기존 함수는 deprecated, 테스트/실제 코드에서 위 API 기반 함수 사용 권장)
// FindSpreadsheetIDByYear: deprecated, 테스트에서는 FindSpreadsheetIDByYearAPI 사용
func FindSpreadsheetIDByYear(ctx context.Context, driveSrv *drive.Service, folderID string, year int) (string, error) {
	driveAPI := &RealDriveAPI{srv: driveSrv}
	return FindSpreadsheetIDByYearAPI(ctx, driveAPI, folderID, year)
}
// ExtractDailyFocusData: deprecated, 테스트에서는 ExtractDailyFocusDataAPI 사용
func ExtractDailyFocusData(sheetsSrv *sheets.Service, spreadsheetID string, year, month, day int) (common.FocusData, string, error) {
	sheetsAPI := &RealSheetsAPI{srv: sheetsSrv}
	return ExtractDailyFocusDataAPI(sheetsAPI, spreadsheetID, year, month, day)
}
