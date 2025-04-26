package common

// Categories: 집중도 분석에 사용되는 공통 카테고리
// - 업무, 학습, 취미, 수면, 이동 순서로 고정
var Categories = []string{"업무", "학습", "취미", "수면", "이동"}

// CategoryColors: 카테고리별 Google Sheets 색상(RGB 0~1)
// - 카테고리 추가 시 이 맵에 색상도 같이 추가할 것
// - 예시: CategoryColors["업무"]
var CategoryColors = map[string][3]float32{
	"업무":  {0.8, 0.9, 1.0},
	"학습":  {0.8, 1.0, 0.8},
	"취미":  {1.0, 0.9, 0.8},
	"수면":  {0.9, 0.8, 1.0},
	"이동":  {0.95, 0.95, 0.95},
} 