package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/crispy/focus-time-tracker/internal/config"
	"github.com/crispy/focus-time-tracker/internal/sheets"
)

func main() {
	// 기본: 진단 모드
	config.LoadEnv()
	fmt.Println("[진단] Focus Time Tracker & Analyzer - Google Sheets 진단 모드")

	// 1. 환경변수 체크
	creds := config.Envs.GSheetsCredentialsJSON
	if creds == "" {
		fmt.Println("[에러] 환경변수 GSHEETS_CREDENTIALS_JSON이 비어 있습니다. .env 또는 환경변수 설정을 확인하세요.")
		os.Exit(1)
	}
	fmt.Println("[OK] 환경변수 GSHEETS_CREDENTIALS_JSON이 설정되어 있습니다.")

	// 2. 서비스 계정 이메일 추출
	type sa struct {
		ClientEmail string `json:"client_email"`
	}
	var s sa
	err := json.Unmarshal([]byte(creds), &s)
	if err != nil {
		fmt.Println("[에러] 서비스 계정 이메일 파싱 실패:", err)
	} else {
		fmt.Println("[OK] 서비스 계정 이메일:", s.ClientEmail)
	}

	// 3. Google Sheets API 인증 및 파일 생성 시도
	ctx := context.Background()
	sheetsSrv, driveSrv, err := sheets.NewService(ctx)
	if err != nil {
		fmt.Println("[에러] Google Sheets API 인증 실패:", err)
		os.Exit(1)
	}
	fmt.Println("[OK] Google Sheets API 인증 성공")

	// 4. 파일 생성 시도
	title := "진단용 테스트시트"
	year := time.Now().Year()
	spreadsheetID, err := sheets.CreateYearlySheet(sheetsSrv, driveSrv, title, year)
	if err != nil {
		fmt.Println("[에러] Google Sheets 파일 생성 실패:", err)
		os.Exit(1)
	}
	fmt.Println("[OK] Google Sheets 파일 생성 성공! Spreadsheet ID:", spreadsheetID)
	fmt.Println("[참고] https://docs.google.com/spreadsheets/d/" + spreadsheetID)
}
