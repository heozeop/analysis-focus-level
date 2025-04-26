package exporter

import (
	"fmt"
	"log"
	"os/exec"
)

// GitRun: git 명령 실행 및 결과 로그 출력
// - args: git 명령 인자 (예: ["add", "."])
// 반환: 에러 (실패 시)
func GitRun(args ...string) error {
	log.Printf("[git] 실행: %v", args)
	cmd := exec.Command("git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[git] 에러: %s", string(out))
		return fmt.Errorf("git %v: %w", args, err)
	}
	log.Printf("[git] 결과: %s", string(out))
	return nil
}

// PushGitbookAssets: gitbook repo에 그래프 push
// - repoPath: gitbook 저장소 경로
// - commitMsg: 커밋 메시지
// 반환: 에러 (없으면 nil)
func PushGitbookAssets(repoPath, commitMsg string) error {
	log.Println("[PushGitbookAssets] === gitbook(submodule) push 시작 ===")
	cmds := [][]string{
		{"-C", repoPath, "add", ".gitbook/assets/graph.png"},
		{"-C", repoPath, "add", ".gitbook/assets/timeslot-images.png"},
		{"-C", repoPath, "commit", "-m", commitMsg},
		{"-C", repoPath, "pull", "--rebase", "origin", "main"},
		{"-C", repoPath, "push", "--no-verify", "origin", "HEAD:main"},
	}
	for _, args := range cmds {
		if err := GitRun(args...); err != nil {
			return fmt.Errorf("[gitbook push 단계] %w", err)
		}
	}
	log.Println("[PushGitbookAssets] === gitbook(submodule) push 끝 ===")
	return nil
}

// PushMainAssets: main repo에 데이터/이미지 push
// - dateStr: 날짜 문자열 (예: "2024-06-01")
// - jsonRelPath: JSON 상대 경로
// - commitMsg: 커밋 메시지
// 반환: 에러 (없으면 nil)
func PushMainAssets(dateStr, jsonRelPath, commitMsg string) error {
	log.Println("[PushMainAssets] === main push 시작 ===")
	cmds := [][]string{
		{"add", "."},
		{"commit", "-m", commitMsg},
		{"pull", "--rebase", "origin", "main"},
		{"push", "--no-verify", "origin", "HEAD:main"},
	}
	for _, args := range cmds {
		if err := GitRun(args...); err != nil {
			return fmt.Errorf("[main push 단계] %w", err)
		}
	}
	log.Println("[PushMainAssets] === main push 끝 ===")
	return nil
}
