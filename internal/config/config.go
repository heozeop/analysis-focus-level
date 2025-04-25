package config

import "os"

// GetEnv: 환경변수 로딩 (기본값 지원)
func GetEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// TODO: 필요시 config 파일/구조체 등 확장 