package sheets

import (
	"fmt"
	"strconv"

	"github.com/crispy/focus-time-tracker/internal/common"
	"google.golang.org/api/sheets/v4"
)

// colIdxToName: 1-based 컬럼 인덱스를 엑셀 컬럼명(B, C, ..., AA, AB...)으로 변환
func colIdxToName(idx int) string {
	name := ""
	for idx > 0 {
		idx--
		name = fmt.Sprintf("%c", 'A'+(idx%26)) + name
		idx /= 26
	}
	return name
}

// ParseDailyData: 각 날짜, 각 카테고리별 [Label, Focus, 확인] 컬럼을 읽고, 확인(체크)된 데이터만 반환
func ParseDailyData(srv *sheets.Service, spreadsheetID, sheetName string, dateCol int) (labels []string, scores []int, err error) {
	categories := common.Categories
	catCount := len(categories)
	rowCount := 144
	// dateCol: 1일=1, 2일=2, ...
	// 1일의 첫 번째 카테고리 Label 컬럼 인덱스: 2 + (dateCol-1)*catCount*3
	startCol := 2 + (dateCol-1)*catCount*3
	// 전체 읽을 범위: Label~확인까지 모든 카테고리
	endCol := startCol + catCount*3 - 1
	startColName := colIdxToName(startCol)
	endColName := colIdxToName(endCol)
	rangeStr := fmt.Sprintf("'%s'!%s2:%s145", sheetName, startColName, endColName)
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, rangeStr).Do()
	if err != nil {
		return nil, nil, err
	}
	for i := 0; i < rowCount; i++ {
		row := []interface{}{}
		if i < len(resp.Values) {
			row = resp.Values[i]
		}
		for catIdx := 0; catIdx < catCount; catIdx++ {
			label := ""
			score := 0
			confirm := ""
			if len(row) > catIdx*3 {
				label = fmt.Sprintf("%v", row[catIdx*3])
			}
			if len(row) > catIdx*3+1 {
				if v, err := strconv.Atoi(fmt.Sprintf("%v", row[catIdx*3+1])); err == nil {
					score = v
				}
			}
			if len(row) > catIdx*3+2 {
				confirm = fmt.Sprintf("%v", row[catIdx*3+2])
			}
			if confirm == "TRUE" || confirm == "true" || confirm == "1" {
				labels = append(labels, label)
				scores = append(scores, score)
			}
		}
	}
	return labels, scores, nil
}

// getSheetByName: 시트 이름으로 시트 정보 조회
func getSheetByName(srv *sheets.Service, spreadsheetID, name string) (*sheets.Sheet, error) {
	ss, err := srv.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return nil, err
	}
	for _, s := range ss.Sheets {
		if s.Properties.Title == name {
			return s, nil
		}
	}
	return nil, fmt.Errorf("시트 %s를 찾을 수 없음", name)
}

// toConditionValues: []string → []*sheets.ConditionValue 변환
func toConditionValues(opts []string) []*sheets.ConditionValue {
	var out []*sheets.ConditionValue
	for _, o := range opts {
		out = append(out, &sheets.ConditionValue{UserEnteredValue: o})
	}
	return out
} 