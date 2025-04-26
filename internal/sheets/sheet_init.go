package sheets

import (
	"fmt"
	"time"

	"google.golang.org/api/sheets/v4"
)

// initSheetData: 월별 시트에 시간표/헤더/빈 데이터 초기화
func initSheetData(sheetsSrv *sheets.Service, spreadsheetID, sheetName string, year, month int) error {
	const maxCols = 1 + 31*2 // 시간 + (31일*2)
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, -1).Day()
	row := []interface{}{fmt.Sprintf("시간 (%02d:00 ~ %02d:50)", 0, 23)}
	for d := 1; d <= lastDay; d++ {
		date := time.Date(year, time.Month(month), d, 0, 0, 0, 0, time.UTC)
		weekday := "일월화수목금토"[date.Weekday()]
		row = append(row, fmt.Sprintf("%d일(%c) Label", d, weekday), fmt.Sprintf("%d일(%c) Focus", d, weekday))
	}
	// 남은 컬럼은 빈 값으로 채움
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
		for d := 1; d <= lastDay*2; d++ {
			row = append(row, "")
		}
		// 남은 컬럼은 빈 값으로 채움
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