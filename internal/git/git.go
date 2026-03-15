package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Status string

const (
	Cloned    Status = "cloned"
	Pulled    Status = "pulled"
	UpToDate  Status = "up-to-date"
	Failed    Status = "failed"
)

type Result struct {
	Status  Status
	Message string
}

func CloneOrPull(name, cloneURL, destDir string) Result {
	repoPath := filepath.Join(destDir, name)

	if _, err := os.Stat(repoPath); err == nil {
		cmd := exec.Command("git", "-C", repoPath, "pull", "--ff-only")
		out, err := cmd.CombinedOutput()
		msg := strings.TrimSpace(string(out))
		if err != nil {
			return Result{Failed, msg}
		}
		if strings.Contains(msg, "Already up to date") {
			return Result{UpToDate, ""}
		}
		return Result{Pulled, msg}
	}

	cmd := exec.Command("git", "clone", cloneURL, repoPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Failed, strings.TrimSpace(string(out))}
	}
	return Result{Cloned, ""}
}
