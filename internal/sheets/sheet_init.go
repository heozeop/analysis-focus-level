package sheets

import (
	"fmt"
	"time"

	"github.com/crispy/focus-time-tracker/internal/common"
	"google.golang.org/api/sheets/v4"
)

// initSheetData: 월별 시트에 시간표/헤더/빈 데이터 초기화 (카테고리 수만큼 동적 확장)
func initSheetData(sheetsSrv *sheets.Service, spreadsheetID, sheetName string, year, month int) error {
	categories := common.Categories
	catCount := len(categories)
	const maxDays = 31
	maxCols := 1 + maxDays*catCount*3 // 시간 + (31일*카테고리수*3)
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, -1).Day()
	// 헤더 생성
	row := []interface{}{fmt.Sprintf("시간 (%02d:00 ~ %02d:50)", 0, 23)}
	for d := 1; d <= lastDay; d++ {
		for _, cat := range categories {
			date := time.Date(year, time.Month(month), d, 0, 0, 0, 0, time.UTC)
			weekday := "일월화수목금토"[date.Weekday()]
			row = append(row,
				fmt.Sprintf("%d일(%c) %s Label", d, weekday, cat),
				fmt.Sprintf("%d일(%c) %s Focus", d, weekday, cat),
				fmt.Sprintf("%d일(%c) %s 확인", d, weekday, cat),
			)
		}
	}
	for len(row) < maxCols {
		row = append(row, "")
	}
	var rows [][]interface{}
	rows = append(rows, row)
	for t := 0; t < 24*6; t++ {
		hour := t / 6
		min := (t % 6) * 10
		timeStr := fmt.Sprintf("%02d:%02d", hour, min)
		row := []interface{}{timeStr}
		for d := 1; d <= lastDay; d++ {
			for range categories {
				row = append(row, "", "", "") // Label, Focus, 확인
			}
		}
		for len(row) < maxCols {
			row = append(row, "")
		}
		rows = append(rows, row)
	}
	vr := &sheets.ValueRange{
		Range:  fmt.Sprintf("'%s'!A1", sheetName),
		Values: rows,
	}
	_, err := sheetsSrv.Spreadsheets.Values.Update(spreadsheetID, vr.Range, vr).ValueInputOption("RAW").Do()
	if err != nil {
		return fmt.Errorf("시트 %s 데이터 초기화 실패: %v", sheetName, err)
	}
	return nil
} 