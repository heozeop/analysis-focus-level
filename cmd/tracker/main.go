package main

import (
	"context"
	"fmt"

	"github.com/crispy/focus-time-tracker/internal/sheets"
	"github.com/crispy/focus-time-tracker/internal/analyzer"
	"github.com/crispy/focus-time-tracker/internal/exporter"
)

func main() {
	fmt.Println("Focus Time Tracker & Analyzer - Go Edition")

	ctx := context.Background()

	// 1. Google Sheets API 서비스 생성
	srv, err := sheets.NewService(ctx)
	if err != nil {
		panic(err)
	}

	// 2. (예시) 특정 시트/날짜 데이터 파싱
	spreadsheetId := "your-spreadsheet-id"
	sheetName := "4월"
	dateCol := 2 // 예: 4월 25일이 2번째 날짜라면
	labels, scores, err := sheets.ParseDailyData(srv, spreadsheetId, sheetName, dateCol)
	if err != nil {
		panic(err)
	}

	// 3. 집계/분석
	data := analyzer.AnalyzeFocus(labels, scores)
	data.Date = "2025-04-25"

	// 4. JSON 저장 및 PR 생성
	jsonPath := "assets/data/2025-04-25.json"
	repoPath := "../gitbook-repo" // 실제 GitBook repo 경로
	branch := "auto/2025-04-25"
	prTitle := "자동 집중도 데이터: 2025-04-25"
	if err := exporter.ExportAndPR(data, jsonPath, repoPath, branch, prTitle); err != nil {
		panic(err)
	}

	fmt.Println("완료!")
} 