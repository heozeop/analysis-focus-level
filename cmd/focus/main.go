package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/crispy/focus-time-tracker/internal/config"
	"github.com/crispy/focus-time-tracker/internal/sheets"
	"github.com/crispy/focus-time-tracker/internal/exporter"
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
	repoPath := os.Getenv("GITBOOK_REPO_PATH")
	if folderID == "" || repoPath == "" {
		log.Fatal("GSHEETS_PARENT_FOLDER_ID, GITBOOK_REPO_PATH 환경변수를 설정하세요.")
	}
	if err := exporter.ExtractAndPush(ctx, sheetsSrv, driveSrv, folderID, repoPath, time.Now()); err != nil {
		log.Fatalf("ExtractAndPush 실패: %v", err)
	}
	fmt.Println("완료! Push 완료.")
}
