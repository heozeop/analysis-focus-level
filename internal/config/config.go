package config

import (
	"encoding/base64"
	"os"

	"github.com/joho/godotenv"
)

type Env struct {
	GSheetsCredentialsJSON string
	GSheetsParentFolderID  string
	GoogleSheetTest        string
	GH_TOKEN               string
	RepoDownloadPath       string
	GitbookRepoPath        string
	// 필요한 항목 추가 가능
}

var Envs Env

func LoadEnv() {
	_ = godotenv.Load("../../.env")
	_ = godotenv.Load("../.env")
	_ = godotenv.Load(".env")
	b64 := os.Getenv("GSHEETS_CREDENTIALS_JSON")
	var decoded string
	if b64 != "" {
		if b, err := base64.StdEncoding.DecodeString(b64); err == nil {
			decoded = string(b)
		}
	}
	Envs = Env{
		GSheetsCredentialsJSON: decoded,
		GSheetsParentFolderID:  os.Getenv("GSHEETS_PARENT_FOLDER_ID"),
		GoogleSheetTest:        os.Getenv("GOOGLE_SHEET_TEST"),
		GH_TOKEN:               os.Getenv("GH_TOKEN"),
		RepoDownloadPath:       os.Getenv("REPO_DOWNLOAD_PATH"),
		GitbookRepoPath:        os.Getenv("GITBOOK_REPO_PATH"),
	}
}
