package sheets

import (
	"fmt"

	"google.golang.org/api/sheets/v4"
)

// UpgradeSheetToNewFormatFrom: 지정한 월/일(startMonth, startDay)부터 연말까지 시트 포맷 업그레이드
// - sheetsSrv: Google Sheets API 서비스
// - spreadsheetID: 스프레드시트 ID
// - year: 연도
// - startMonth: 시작 월
// - startDay: 시작 일
// 반환: 에러 (없으면 nil)
func UpgradeSheetToNewFormatFrom(sheetsSrv *sheets.Service, spreadsheetID string, year, startMonth, startDay int) error {
	sheetTitles := []string{"1월", "2월", "3월", "4월", "5월", "6월", "7월", "8월", "9월", "10월", "11월", "12월"}
	for monthIdx, name := range sheetTitles {
		month := monthIdx + 1
		if month < startMonth {
			continue // 시작 월 이전은 건너뜀
		}
		// 1. 월별 시트 데이터 재초기화 (기존 데이터는 덮어씀)
		if err := initSheetDataFrom(sheetsSrv, spreadsheetID, name, year, month, startDay, month == startMonth); err != nil {
			return fmt.Errorf("%s 시트 데이터 초기화 실패: %v", name, err)
		}
		// 2. 월별 시트 스타일/유효성/조건부서식 재적용
		if err := applySheetStyles(sheetsSrv, spreadsheetID, name); err != nil {
			return fmt.Errorf("%s 시트 스타일/유효성/조건부서식 적용 실패: %v", name, err)
		}
	}
	return nil
}
