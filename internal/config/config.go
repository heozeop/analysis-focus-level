package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Env struct {
	GSheetsCredentialsJSON string
	GSheetsParentFolderID  string
	GoogleSheetTest        string
	GH_TOKEN               string
	// 필요한 항목 추가 가능
}

var Envs Env

func LoadEnv() {
	_ = godotenv.Load("../../.env")
	_ = godotenv.Load("../.env")
	_ = godotenv.Load(".env")
	Envs = Env{
		GSheetsCredentialsJSON: os.Getenv("GSHEETS_CREDENTIALS_JSON"),
		GSheetsParentFolderID:  os.Getenv("GSHEETS_PARENT_FOLDER_ID"),
		GoogleSheetTest:        os.Getenv("GOOGLE_SHEET_TEST"),
		GH_TOKEN:               os.Getenv("GH_TOKEN"),
	}
} 