package sheets

import (
	"fmt"
	"strconv"

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

// ParseDailyData: 각 날짜, 각 카테고리별 [Label, Focus] 컬럼을 읽고, 값이 있는 데이터만 반환
func ParseDailyData(srv *sheets.Service, spreadsheetID, sheetName string, dateCol int) (labels []string, scores []int, err error) {
	rowCount := 144
	// dateCol: 1일=1, 2일=2, ...
	// 1일의 Label 컬럼 인덱스: 2 + (dateCol-1)*2
	startCol := 2 + (dateCol-1)*2
	endCol := startCol + 1
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
		label := ""
		score := 0
		if len(row) > 0 {
			label = fmt.Sprintf("%v", row[0])
		}
		if len(row) > 1 {
			if v, err := strconv.Atoi(fmt.Sprintf("%v", row[1])); err == nil {
				score = v
			}
		}
		if label != "" {
			labels = append(labels, label)
			scores = append(scores, score)
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