package sheets

import (
	"context"
	"log"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// NewService: Google Sheets API 서비스 생성
func NewService(ctx context.Context) (*sheets.Service, error) {
	// credentials.json: GCP OAuth2 서비스 계정 키 파일
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.JWTConfigFromJSON(b, sheets.SpreadsheetsScope)
	if err != nil {
		return nil, err
	}

	ts := config.TokenSource(ctx)
	return sheets.NewService(ctx, option.WithTokenSource(ts))
}

// TODO: PRD 2.1 - 시트 생성 함수 (연도별, 12개월 탭, 10분 단위 row 등)
// TODO: PRD 2.2 - 시트 데이터 파싱 함수 (10분 단위, 라벨/집중도) 

// CreateYearlySheet: 연도별 시트 생성 (12개월 탭, 10분 단위 row, 날짜별 2컬럼)
func CreateYearlySheet(srv *sheets.Service, title string, year int) (string, error) {
	// TODO: 실제 Google Sheets 생성 및 구조화 구현
	// 1. 새 스프레드시트 생성 (title: "YYYY Focus Log")
	// 2. 12개 탭(1월~12월), 각 탭에 144개 row(00:00~23:50, 10분 단위)
	// 3. 각 날짜별 2컬럼(Label, Focus) 생성
	return "", nil
}

// ParseDailyData: 특정 날짜의 10분 단위 라벨/집중도 데이터 파싱
func ParseDailyData(srv *sheets.Service, spreadsheetId, sheetName string, dateCol int) (labels []string, scores []int, err error) {
	// TODO: Google Sheets에서 해당 날짜의 10분 단위 라벨/집중도 데이터 읽기
	// 1. sheetName(월)에서 dateCol(Label), dateCol+1(Focus) 컬럼 읽기
	// 2. 144개 row(00:00~23:50) 순서대로 파싱
	return nil, nil, nil
} 