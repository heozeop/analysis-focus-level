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

// ParseDailyData: 특정 날짜의 10분 단위 라벨/집중도 데이터 파싱
func ParseDailyData(srv *sheets.Service, spreadsheetID, sheetName string, dateCol int) (labels []string, scores []int, err error) {
	// dateCol: 1일(Label)이 2, 1일(Focus)이 3, 2일(Label)이 4, ...
	labelCol := colIdxToName(2 + (dateCol-1)*2) // B, D, F, ...
	focusCol := colIdxToName(3 + (dateCol-1)*2) // C, E, G, ...
	labelRange := fmt.Sprintf("'%s'!%s2:%s145", sheetName, labelCol, labelCol)
	focusRange := fmt.Sprintf("'%s'!%s2:%s145", sheetName, focusCol, focusCol)
	labelResp, err := srv.Spreadsheets.Values.Get(spreadsheetID, labelRange).Do()
	if err != nil {
		return nil, nil, err
	}
	focusResp, err := srv.Spreadsheets.Values.Get(spreadsheetID, focusRange).Do()
	if err != nil {
		return nil, nil, err
	}
	for i := 0; i < 144; i++ {
		label := ""
		if i < len(labelResp.Values) && len(labelResp.Values[i]) > 0 {
			label = fmt.Sprintf("%v", labelResp.Values[i][0])
		}
		score := 0
		if i < len(focusResp.Values) && len(focusResp.Values[i]) > 0 {
			if v, err := strconv.Atoi(fmt.Sprintf("%v", focusResp.Values[i][0])); err == nil {
				score = v
			}
		}
		labels = append(labels, label)
		scores = append(scores, score)
	}
	// 만약 데이터가 완전히 비어있으면 144개 0점으로 채움
	if len(labelResp.Values) == 0 && len(focusResp.Values) == 0 {
		labels = make([]string, 144)
		scores = make([]int, 144)
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