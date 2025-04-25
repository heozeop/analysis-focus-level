package exporter

import (
	"fmt"
	"log"
	"os/exec"
)

// GitRun runs a git command and logs output.
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