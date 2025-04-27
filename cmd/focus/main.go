package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/crispy/focus-time-tracker/internal/config"
	"github.com/crispy/focus-time-tracker/internal/exporter"
	"github.com/crispy/focus-time-tracker/internal/sheets"
)

func main() {
	config.LoadEnv()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "extract":
			extract()
			return
		case "push":
			if len(os.Args) < 5 {
				fmt.Println("Usage: focus push <dateStr> <jsonRelPath> <commitMsg>")
				return
			}
			repoPath := config.Envs.GitbookRepoPath
			dateStr := os.Args[2]
			jsonRelPath := os.Args[3]
			commitMsg := os.Args[4]
			if err := exporter.Push(repoPath, dateStr, jsonRelPath, commitMsg); err != nil {
				log.Fatalf("Push 실패: %v", err)
			}
			fmt.Println("Push 완료!")
			return
		}
	}
	fmt.Println("Usage: focus extract | push <dateStr> <jsonRelPath> <commitMsg>")
}

func extract() {
	ctx := context.Background()
	sheetsSrv, driveSrv, err := sheets.NewService(ctx)
	if err != nil {
		log.Fatalf("Google Sheets API 인증 실패: %v", err)
	}
	folderID := config.Envs.GSheetsParentFolderID
	repoPath := config.Envs.GitbookRepoPath
	repoDownloadPath := config.Envs.RepoDownloadPath

	if folderID == "" || repoPath == "" || repoDownloadPath == "" {
		log.Fatal("GSHEETS_PARENT_FOLDER_ID, REPO_DOWNLOAD_PATH 환경변수를 설정하세요.")
	}
	dateStr, jsonRelPath, commitMsg, err := exporter.Extract(ctx, sheetsSrv, driveSrv, folderID, repoPath, repoDownloadPath, time.Now())
	if err != nil {
		log.Fatalf("Extract 실패: %v", err)
	}
	fmt.Printf("추출 완료! dateStr: %s, jsonRelPath: %s, commitMsg: %s\n", dateStr, jsonRelPath, commitMsg)

	// 파일 저장 대신 표준 출력으로 결과만 출력 (CI/CD 연동)
	fmt.Printf("%s|%s|%s\n", dateStr, jsonRelPath, commitMsg)
}
