package sheets

import (
	"fmt"

	"github.com/crispy/focus-time-tracker/internal/common"
	"google.golang.org/api/sheets/v4"
)

// applySheetStyles: 월별 시트에 스타일, 유효성, 조건부 서식 적용 (카테고리 수만큼 동적 확장)
// - sheetsSrv: Google Sheets API 서비스
// - spreadsheetID: 스프레드시트 ID
// - sheetName: 시트 이름
// 반환: 에러 (없으면 nil)
func applySheetStyles(sheetsSrv *sheets.Service, spreadsheetID, sheetName string) error {
	const maxDays = 31
	const maxCols = 1 + maxDays*2 // 시간 + (31일*2)
	categories := common.Categories
	labelColors := map[string]*sheets.Color{} // 카테고리별 색상 매핑
	for _, cat := range categories {
		rgb := common.CategoryColors[cat]
		labelColors[cat] = &sheets.Color{Red: float64(rgb[0]), Green: float64(rgb[1]), Blue: float64(rgb[2])}
	}
	gray := &sheets.Color{Red: 0.95, Green: 0.95, Blue: 0.95}
	black := &sheets.Color{Red: 0, Green: 0, Blue: 0}

	sheet, err := getSheetByName(sheetsSrv, spreadsheetID, sheetName)
	if err != nil {
		return fmt.Errorf("시트 ID 조회 실패: %v", err)
	}
	sheetID := sheet.Properties.SheetId
	requests := []*sheets.Request{}
	// 1. 1행(헤더) 전체, A열 전체에 옅은 회색
	requests = append(requests,
		&sheets.Request{
			RepeatCell: &sheets.RepeatCellRequest{
				Range:  &sheets.GridRange{SheetId: sheetID, StartRowIndex: 0, EndRowIndex: 1},
				Cell:   &sheets.CellData{UserEnteredFormat: &sheets.CellFormat{BackgroundColor: gray}},
				Fields: "userEnteredFormat.backgroundColor",
			},
		},
		&sheets.Request{
			RepeatCell: &sheets.RepeatCellRequest{
				Range:  &sheets.GridRange{SheetId: sheetID, StartColumnIndex: 0, EndColumnIndex: 1},
				Cell:   &sheets.CellData{UserEnteredFormat: &sheets.CellFormat{BackgroundColor: gray}},
				Fields: "userEnteredFormat.backgroundColor",
			},
		},
		// A열 고정
		&sheets.Request{
			UpdateSheetProperties: &sheets.UpdateSheetPropertiesRequest{
				Properties: &sheets.SheetProperties{
					SheetId: sheetID,
					GridProperties: &sheets.GridProperties{
						FrozenColumnCount: 1,
					},
				},
				Fields: "gridProperties.frozenColumnCount",
			},
		},
	)
	// 2. 모든 셀에 검은 border
	requests = append(requests, &sheets.Request{
		UpdateBorders: &sheets.UpdateBordersRequest{
			Range:           &sheets.GridRange{SheetId: sheetID},
			Top:             &sheets.Border{Style: "SOLID", Color: black},
			Bottom:          &sheets.Border{Style: "SOLID", Color: black},
			Left:            &sheets.Border{Style: "SOLID", Color: black},
			Right:           &sheets.Border{Style: "SOLID", Color: black},
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
	// 3. 각 날짜별로 Label/Focus 컬럼에 유효성/조건부서식 적용 (카테고리 드롭다운은 Label에만)
	for d := 0; d < maxDays; d++ {
		labelCol := 1 + d*2
		focusCol := labelCol + 1
		if labelCol >= maxCols || focusCol >= maxCols {
			break
		}
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
						Values: toConditionValues(categories),
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
								Type:   "TEXT_EQ",
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
			return fmt.Errorf("시트 %s 스타일/유효성/조건부서식 적용 실패: %v", sheetName, err)
		}
	}
	return nil
}
 