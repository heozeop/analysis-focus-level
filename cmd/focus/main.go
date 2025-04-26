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

	if len(os.Args) > 1 && os.Args[1] == "extract" {
		extract()
		return
	}
	fmt.Println("Usage: focus extract | analyze")
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
	if err := exporter.ExtractAndPush(ctx, sheetsSrv, driveSrv, folderID, repoPath, repoDownloadPath, time.Now()); err != nil {
		log.Fatalf("ExtractAndPush 실패: %v", err)
	}
	fmt.Println("완료! Push 완료.")
}
