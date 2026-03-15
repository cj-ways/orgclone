// Package git wraps git clone and git pull for local repo management.
package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Status string

const (
	Cloned   Status = "cloned"
	Pulled   Status = "pulled"
	UpToDate Status = "up-to-date"
	Failed   Status = "failed"
)

type Result struct {
	Status  Status
	Message string // populated on failure
}

// CloneOrPull clones a repo if it doesn't exist locally, or pulls if it does.
func CloneOrPull(name, cloneURL, destDir string) Result {
	repoPath := filepath.Join(destDir, name)

	if exists(repoPath) {
		return pull(repoPath)
	}
	return clone(cloneURL, repoPath)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func clone(cloneURL, dest string) Result {
	out, err := exec.Command("git", "clone", cloneURL, dest).CombinedOutput()
	if err != nil {
		return Result{Failed, strings.TrimSpace(string(out))}
	}
	return Result{Status: Cloned}
}

func pull(repoPath string) Result {
	out, err := exec.Command("git", "-C", repoPath, "pull", "--ff-only").CombinedOutput()
	msg := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Failed, msg}
	}
	if strings.Contains(msg, "Already up to date") {
		return Result{Status: UpToDate}
	}
	return Result{Status: Pulled}
}
