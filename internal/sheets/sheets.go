package sheets

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"github.com/crispy/focus-time-tracker/internal/config"
)

// NewService: Google Sheets API + Drive API 서비스 생성 (환경변수 GSHEETS_CREDENTIALS_JSON 사용)
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

// TODO: PRD 2.1 - 시트 생성 함수 (연도별, 12개월 탭, 10분 단위 row 등)
// TODO: PRD 2.2 - 시트 데이터 파싱 함수 (10분 단위, 라벨/집중도) 

// CreateYearlySheet: 연도별 시트 생성 후 폴더로 이동 (환경변수 GSHEETS_PARENT_FOLDER_ID 사용)
func CreateYearlySheet(sheetsSrv *sheets.Service, driveSrv *drive.Service, title string, year int) (string, error) {
	spreadsheet := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: fmt.Sprintf("%d Focus Log", year),
		},
	}
	sheetTitles := []string{"1월", "2월", "3월", "4월", "5월", "6월", "7월", "8월", "9월", "10월", "11월", "12월"}
	for _, name := range sheetTitles {
		spreadsheet.Sheets = append(spreadsheet.Sheets, &sheets.Sheet{
			Properties: &sheets.SheetProperties{Title: name},
		})
	}
	created, err := sheetsSrv.Spreadsheets.Create(spreadsheet).Do()
	if err != nil {
		return "", err
	}
	spreadsheetID := created.SpreadsheetId

	// 폴더 이동 처리
	parentFolderID := config.Envs.GSheetsParentFolderID
	if parentFolderID != "" {
		_, err := driveSrv.Files.Update(spreadsheetID, nil).
			AddParents(parentFolderID).
			Do()
		if err != nil {
			return spreadsheetID, fmt.Errorf("시트 생성 후 폴더 이동 실패: %v", err)
		}
	}

	// 색상/옵션 상수 정의
	var (
		labelOptions = []string{"업무", "학습", "취미", "수면", "이동"}
		labelColors = map[string]*sheets.Color{
			"업무":  {Red: 0.8, Green: 0.9, Blue: 1.0},
			"학습":  {Red: 0.8, Green: 1.0, Blue: 0.8},
			"취미":  {Red: 1.0, Green: 0.9, Blue: 0.8},
			"수면":  {Red: 0.9, Green: 0.8, Blue: 1.0},
			"이동":  {Red: 0.95, Green: 0.95, Blue: 0.95},
		}
		gray  = &sheets.Color{Red: 0.95, Green: 0.95, Blue: 0.95}
		black = &sheets.Color{Red: 0, Green: 0, Blue: 0}
	)

	for monthIdx, name := range sheetTitles {
		var rows [][]interface{}
		month := monthIdx + 1
		firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		lastDay := firstDay.AddDate(0, 1, -1).Day()
		row := []interface{}{fmt.Sprintf("시간 (%02d:00 ~ %02d:50)", 0, 23)}
		for d := 1; d <= lastDay; d++ {
			date := time.Date(year, time.Month(month), d, 0, 0, 0, 0, time.UTC)
			weekday := "일월화수목금토"[date.Weekday()]
			row = append(row, fmt.Sprintf("%d일(%c) Label", d, weekday), fmt.Sprintf("%d일(%c) Focus", d, weekday))
		}
		rows = append(rows, row)
		for t := 0; t < 24*6; t++ {
			hour := t / 6
			min := (t % 6) * 10
			timeStr := fmt.Sprintf("%02d:%02d", hour, min)
			row := []interface{}{timeStr}
			for d := 1; d <= lastDay*2; d++ {
				row = append(row, "")
			}
			rows = append(rows, row)
		}
		vr := &sheets.ValueRange{
			Range:  fmt.Sprintf("'%s'!A1", name),
			Values: rows,
		}
		_, err := sheetsSrv.Spreadsheets.Values.Update(spreadsheetID, vr.Range, vr).ValueInputOption("RAW").Do()
		if err != nil {
			return spreadsheetID, fmt.Errorf("시트 %s 데이터 초기화 실패: %v", name, err)
		}

		// 스타일/유효성/조건부 서식 BatchUpdate
		sheet, err := getSheetByName(sheetsSrv, spreadsheetID, name)
		if err != nil {
			return spreadsheetID, fmt.Errorf("시트 ID 조회 실패: %v", err)
		}
		sheetID := sheet.Properties.SheetId
		requests := []*sheets.Request{}
		// 1. 1행(헤더) 전체, A열 전체에 옅은 회색
		requests = append(requests,
			&sheets.Request{
				RepeatCell: &sheets.RepeatCellRequest{
					Range: &sheets.GridRange{SheetId: sheetID, StartRowIndex: 0, EndRowIndex: 1},
					Cell: &sheets.CellData{UserEnteredFormat: &sheets.CellFormat{BackgroundColor: gray}},
					Fields: "userEnteredFormat.backgroundColor",
				},
			},
			&sheets.Request{
				RepeatCell: &sheets.RepeatCellRequest{
					Range: &sheets.GridRange{SheetId: sheetID, StartColumnIndex: 0, EndColumnIndex: 1},
					Cell: &sheets.CellData{UserEnteredFormat: &sheets.CellFormat{BackgroundColor: gray}},
					Fields: "userEnteredFormat.backgroundColor",
				},
			},
		)
		// 2. 모든 셀에 검은 border
		requests = append(requests, &sheets.Request{
			UpdateBorders: &sheets.UpdateBordersRequest{
				Range: &sheets.GridRange{SheetId: sheetID},
				Top:    &sheets.Border{Style: "SOLID", Color: black},
				Bottom: &sheets.Border{Style: "SOLID", Color: black},
				Left:   &sheets.Border{Style: "SOLID", Color: black},
				Right:  &sheets.Border{Style: "SOLID", Color: black},
				InnerHorizontal: &sheets.Border{Style: "SOLID", Color: black},
				InnerVertical:   &sheets.Border{Style: "SOLID", Color: black},
			},
		})
		// 2-2. 1시간(6행)마다 굵은 border
		for h := 1; h < 24; h++ {
			requests = append(requests, &sheets.Request{
				UpdateBorders: &sheets.UpdateBordersRequest{
					Range: &sheets.GridRange{
						SheetId:       sheetID,
						StartRowIndex: int64(1 + h*6 - 1),
						EndRowIndex:   int64(1 + h*6),
					},
					Bottom: &sheets.Border{Style: "SOLID_MEDIUM", Color: black},
				},
			})
		}
		// 3. Label 컬럼 드롭다운, Focus 컬럼 숫자 제한
		for d := 0; d < lastDay; d++ {
			labelCol := 1 + d*2
			focusCol := 2 + d*2
			// Label 드롭다운
			requests = append(requests, &sheets.Request{
				SetDataValidation: &sheets.SetDataValidationRequest{
					Range: &sheets.GridRange{
						SheetId:          sheetID,
						StartRowIndex:    1,
						EndRowIndex:      145,
						StartColumnIndex: int64(labelCol),
						EndColumnIndex:   int64(labelCol + 1),
					},
					Rule: &sheets.DataValidationRule{
						Condition: &sheets.BooleanCondition{
							Type:   "ONE_OF_LIST",
							Values: toConditionValues(labelOptions),
						},
						Strict: true,
					},
				},
			})
			// Focus 숫자만
			requests = append(requests, &sheets.Request{
				SetDataValidation: &sheets.SetDataValidationRequest{
					Range: &sheets.GridRange{
						SheetId:          sheetID,
						StartRowIndex:    1,
						EndRowIndex:      145,
						StartColumnIndex: int64(focusCol),
						EndColumnIndex:   int64(focusCol + 1),
					},
					Rule: &sheets.DataValidationRule{
						Condition: &sheets.BooleanCondition{
							Type:   "NUMBER_BETWEEN",
							Values: []*sheets.ConditionValue{{UserEnteredValue: "0"}, {UserEnteredValue: "100"}},
						},
						Strict: true,
					},
				},
			})
			// Label 조건부 색상
			for label, color := range labelColors {
				requests = append(requests, &sheets.Request{
					AddConditionalFormatRule: &sheets.AddConditionalFormatRuleRequest{
						Rule: &sheets.ConditionalFormatRule{
							Ranges: []*sheets.GridRange{{
								SheetId:          sheetID,
								StartRowIndex:    1,
								EndRowIndex:      145,
								StartColumnIndex: int64(labelCol),
								EndColumnIndex:   int64(labelCol + 1),
							}},
							BooleanRule: &sheets.BooleanRule{
								Condition: &sheets.BooleanCondition{
									Type:  "TEXT_EQ",
									Values: []*sheets.ConditionValue{{UserEnteredValue: label}},
								},
								Format: &sheets.CellFormat{BackgroundColor: color},
							},
						},
						Index: 0,
					},
				})
			}
		}
		if len(requests) > 0 {
			_, err := sheetsSrv.Spreadsheets.BatchUpdate(spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
				Requests: requests,
			}).Do()
			if err != nil {
				return spreadsheetID, fmt.Errorf("시트 %s 스타일/유효성/조건부서식 적용 실패: %v", name, err)
			}
		}
	}
	return spreadsheetID, nil
}

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
	labelCol := colIdxToName(1 + (dateCol-1)*2) // B, D, F, ...
	focusCol := colIdxToName(2 + (dateCol-1)*2) // C, E, G, ...
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
	return labels, scores, nil
}

// 보조 함수: 시트 이름으로 시트 정보 조회
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

// 보조 함수: []string → []*sheets.ConditionValue 변환
func toConditionValues(opts []string) []*sheets.ConditionValue {
	var out []*sheets.ConditionValue
	for _, o := range opts {
		out = append(out, &sheets.ConditionValue{UserEnteredValue: o})
	}
	return out
} 