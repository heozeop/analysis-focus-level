package sheets

import (
	"fmt"
	"time"

	"google.golang.org/api/sheets/v4"
)

// initSheetData: 월별 시트에 시간표/헤더/빈 데이터 초기화 (카테고리 수만큼 동적 확장)
// - sheetsSrv: 구글 시트 서비스 객체
// - spreadsheetID: 대상 스프레드시트 ID
// - sheetName: 초기화할 시트 이름
// - year, month: 연/월
// 반환: 에러 (없으면 nil)
func initSheetData(sheetsSrv *sheets.Service, spreadsheetID, sheetName string, year, month int) error {
	const maxDays = 31 // 한 달 최대 일수
	maxCols := 1 + maxDays*2 // 시간 열 + (일수*2: Label, Focus)
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC) // 해당 월 1일
	lastDay := firstDay.AddDate(0, 1, -1).Day() // 해당 월 마지막 일

	// --- 헤더 행 생성 ---
	// 첫 번째 열: 시간 범위 표시
	row := []interface{}{fmt.Sprintf("시간 (%02d:00 ~ %02d:50)", 0, 23)}
	// 각 날짜별로 Label, Focus 열 추가
	for d := 1; d <= lastDay; d++ {
		date := time.Date(year, time.Month(month), d, 0, 0, 0, 0, time.UTC)
		weekday := []rune("일월화수목금토")[date.Weekday()]
		row = append(row,
			fmt.Sprintf("%d일(%c) Label", d, weekday), // 예: 1일(월) Label
			fmt.Sprintf("%d일(%c) Focus", d, weekday), // 예: 1일(월) Focus
		)
	}
	// 남는 열은 빈칸으로 채움 (최대 열수 맞춤)
	for len(row) < maxCols {
		row = append(row, "")
	}

	var rows [][]interface{} // 전체 시트 데이터(2차원 배열)
	rows = append(rows, row) // 첫 행: 헤더

	// --- 시간표/빈 데이터 행 생성 ---
	// 10분 단위(24*6=144) 시간 행 생성
	for t := 0; t < 24*6; t++ {
		hour := t / 6
		min := (t % 6) * 10
		timeStr := fmt.Sprintf("%02d:%02d", hour, min) // 예: 09:30
		row := []interface{}{timeStr} // 첫 열: 시간
		// 각 날짜별로 Label, Focus 빈칸 추가
		for d := 1; d <= lastDay; d++ {
			row = append(row, "", "") // Label, Focus
		}
		// 남는 열은 빈칸으로 채움
		for len(row) < maxCols {
			row = append(row, "")
		}
		rows = append(rows, row)
	}

	// --- ValueRange 객체 생성 (시트에 쓸 데이터) ---
	vr := &sheets.ValueRange{
		Range:  fmt.Sprintf("'%s'!A1", sheetName), // 시작 셀
		Values: rows, // 전체 데이터
	}

	// --- 시트에 데이터 업데이트 ---
	_, err := sheetsSrv.Spreadsheets.Values.Update(spreadsheetID, vr.Range, vr).ValueInputOption("RAW").Do()
	if err != nil {
		return fmt.Errorf("시트 %s 데이터 초기화 실패: %v", sheetName, err)
	}
	return nil
}

// initSheetDataFrom: 월별 시트에 시간표/헤더/빈 데이터 초기화 (startDay부터)
// - isStartMonth: true면 startDay 이전 날짜는 nil로 둬서 기존 데이터 보존
func initSheetDataFrom(sheetsSrv *sheets.Service, spreadsheetID, sheetName string, year, month, startDay int, isStartMonth bool) error {
	const maxDays = 31
	maxCols := 1 + maxDays*2
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, -1).Day()

	// --- 헤더 행 생성 ---
	row := []interface{}{fmt.Sprintf("시간 (%02d:00 ~ %02d:50)", 0, 23)}
	for d := 1; d <= lastDay; d++ {
		date := time.Date(year, time.Month(month), d, 0, 0, 0, 0, time.UTC)
		weekday := []rune("일월화수목금토")[date.Weekday()]
		row = append(row,
			fmt.Sprintf("%d일(%c) Label", d, weekday),
			fmt.Sprintf("%d일(%c) Focus", d, weekday),
		)
	}
	for len(row) < maxCols {
		row = append(row, "")
	}

	var rows [][]interface{}
	rows = append(rows, row)

	// --- 시간표/빈 데이터 행 생성 ---
	for t := 0; t < 24*6; t++ {
		hour := t / 6
		min := (t % 6) * 10
		timeStr := fmt.Sprintf("%02d:%02d", hour, min)
		row := []interface{}{timeStr}
		for d := 1; d <= lastDay; d++ {
			if isStartMonth && d < startDay {
				row = append(row, nil, nil) // 기존 데이터 보존 (nil)
			} else {
				row = append(row, "", "") // Label, Focus
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
