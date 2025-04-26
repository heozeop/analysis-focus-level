package sheets

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/api/drive/v3"
)

// FakeSheetsService는 sheets.Service의 최소 mock 구조체
// 실제 API 호출 없이 함수 시그니처만 검증
// 필요한 메서드만 임시로 구현

type FakeSheetsService struct{}
type FakeDriveService struct{}

type MockSheetsAPI struct {
	values [][]interface{}
	err    error
}
func (m *MockSheetsAPI) GetValues(spreadsheetID, readRange string) ([][]interface{}, error) {
	return m.values, m.err
}

type MockDriveAPI struct {
	files []*drive.File
	err   error
}
func (m *MockDriveAPI) FindFiles(ctx context.Context, query string) ([]*drive.File, error) {
	return m.files, m.err
}

func TestNewService_Fake(t *testing.T) {
	// 환경변수 없이도 panic 없이 함수 호출만 되는지 확인
	ctx := context.Background()
	// 실제 NewService는 환경변수 없으면 에러 반환
	_, _, err := NewService(ctx)
	if err == nil {
		t.Errorf("NewService는 환경변수 없으면 에러를 반환해야 함")
	}
}

func TestFindSpreadsheetIDByYearAPI_Property(t *testing.T) {
	ctx := context.Background()
	// 성공 케이스
	driveAPI := &MockDriveAPI{files: []*drive.File{{Id: "abc123", Name: "2024 Focus Log"}}}
	id, err := FindSpreadsheetIDByYearAPI(ctx, driveAPI, "folder", 2024)
	assert.NoError(t, err)
	assert.Equal(t, "abc123", id)

	// 실패 케이스: 파일 없음
	driveAPI = &MockDriveAPI{files: []*drive.File{}}
	_, err = FindSpreadsheetIDByYearAPI(ctx, driveAPI, "folder", 2024)
	assert.Error(t, err)

	// 실패 케이스: API 에러
	driveAPI = &MockDriveAPI{err: errors.New("api error")}
	_, err = FindSpreadsheetIDByYearAPI(ctx, driveAPI, "folder", 2024)
	assert.Error(t, err)
}

func TestExtractDailyFocusDataAPI_Property(t *testing.T) {
	// 성공 케이스: 정상 데이터
	sheetsAPI := &MockSheetsAPI{values: [][]interface{}{{"업무", 10}, {"학습", 20}}}
	data, dateStr, err := ExtractDailyFocusDataAPI(sheetsAPI, "spreadsheetID", 2024, 6, 1)
	assert.NoError(t, err)
	assert.Equal(t, "2024-06-01", dateStr)
	assert.Equal(t, 2, len(data.Categories))
	assert.Equal(t, 10, data.Categories["업무"])
	assert.Equal(t, 20, data.Categories["학습"])

	// 실패 케이스: API 에러
	sheetsAPI = &MockSheetsAPI{err: errors.New("api error")}
	_, _, err = ExtractDailyFocusDataAPI(sheetsAPI, "spreadsheetID", 2024, 6, 1)
	assert.Error(t, err)

	// 엣지 케이스: 빈 데이터
	sheetsAPI = &MockSheetsAPI{values: [][]interface{}{}}
	data, _, err = ExtractDailyFocusDataAPI(sheetsAPI, "spreadsheetID", 2024, 6, 1)
	assert.NoError(t, err)
	assert.Equal(t, 0, data.TotalFocus)
}
