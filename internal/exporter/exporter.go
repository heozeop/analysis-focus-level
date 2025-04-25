package exporter

import (
	"encoding/json"
	"os"
	"os/exec"
	"fmt"

	"github.com/crispy/focus-time-tracker/internal/analyzer"
)

// ExportToJSON: FocusData를 JSON 파일로 저장 (assets/data/YYYY-MM-DD.json)
func ExportToJSON(data analyzer.FocusData, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(data)
}

// ExportAndPR: JSON 저장 후 GitBook repo에 PR 생성
func ExportAndPR(data analyzer.FocusData, path, repoPath, branch, prTitle string) error {
	if err := ExportToJSON(data, path); err != nil {
		return err
	}
	// git add/commit/push
	cmds := [][]string{
		{"git", "-C", repoPath, "checkout", "-b", branch},
		{"git", "-C", repoPath, "add", path},
		{"git", "-C", repoPath, "commit", "-m", prTitle},
		{"git", "-C", repoPath, "push", "origin", branch},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("%v: %s", err, string(out))
		}
	}
	// gh CLI로 PR 생성
	prCmd := exec.Command("gh", "pr", "create", "--repo", repoPath, "--head", branch, "--title", prTitle, "--body", "자동 생성 PR")
	if out, err := prCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("PR 생성 실패: %v: %s", err, string(out))
	}
	return nil
}

// TODO: PRD 2.2 - GitBook repo에 PR 생성 함수 (os/exec 또는 go-git) 