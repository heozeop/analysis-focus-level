package sheets

import (
	"testing"
	"context"
	"github.com/crispy/focus-time-tracker/internal/config"
)

func TestCreateYearlySheet_Integration(t *testing.T) {
	config.LoadEnv()
	if config.Envs.GoogleSheetTest != "1" {
		t.Skip("Set GOOGLE_SHEET_TEST=1 to run this integration test (needs credentials.json and Google Sheets API access)")
	}
	ctx := context.Background()
	sheetsSrv, driveSrv, err := NewService(ctx)
	if err != nil {
		t.Fatalf("NewService error: %v", err)
	}
	year := 2024
	title := "테스트시트"
	spreadsheetID, err := CreateYearlySheet(sheetsSrv, driveSrv, title, year)
	if err != nil {
		t.Fatalf("CreateYearlySheet error: %v", err)
	}
	t.Logf("생성된 시트 ID: %s", spreadsheetID)
	// 실제 시트가 생성되었는지, 2월 헤더가 29일(윤년)까지 생성됐는지 확인
	labels, _, err := ParseDailyData(sheetsSrv, spreadsheetID, "2월", 29) // 2월 29일
	if err != nil {
		t.Fatalf("ParseDailyData error: %v", err)
	}
	if len(labels) != 144 {
		t.Errorf("2월 29일 row 개수 = %d, want 144", len(labels))
	}
	// (테스트 후 시트 삭제는 수동 또는 API로 별도 처리) 
}
