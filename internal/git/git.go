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
	Skipped  Status = "skipped"
	Failed   Status = "failed"
)

type Result struct {
	Status  Status
	Message string // populated on failure or skip
}

// Exists checks whether a repo directory already exists at the destination.
func Exists(name, destDir string) bool {
	info, err := os.Stat(filepath.Join(destDir, name))
	return err == nil && info.IsDir()
}

// CloneOrPull clones a repo if it doesn't exist locally, or pulls if it does.
func CloneOrPull(name, cloneURL, destDir string) Result {
	repoPath := filepath.Join(destDir, name)

	info, statErr := os.Stat(repoPath)

	// Path exists as a file (not a directory) — cannot clone here
	if statErr == nil && !info.IsDir() {
		return Result{Skipped, "path exists as a file, not a directory"}
	}

	// Directory exists
	if statErr == nil && info.IsDir() {
		gitDir := filepath.Join(repoPath, ".git")
		if _, err := os.Stat(gitDir); err != nil {
			// Directory exists but is not a git repo — don't overwrite
			return Result{Skipped, "directory exists but is not a git repo"}
		}
		return pull(repoPath)
	}

	return clone(cloneURL, repoPath)
}

func clone(cloneURL, dest string) Result {
	out, err := exec.Command("git", "clone", "--quiet", cloneURL, dest).CombinedOutput()
	if err != nil {
		return Result{Failed, strings.TrimSpace(string(out))}
	}
	return Result{Status: Cloned}
}

func pull(repoPath string) Result {
	// Check for detached HEAD — don't try to pull
	headCmd := exec.Command("git", "-C", repoPath, "symbolic-ref", "--quiet", "HEAD")
	if err := headCmd.Run(); err != nil {
		return Result{Skipped, "detached HEAD"}
	}

	// Check for dirty working tree — don't risk losing uncommitted work
	dirtyCmd := exec.Command("git", "-C", repoPath, "status", "--porcelain")
	dirtyOutput, err := dirtyCmd.CombinedOutput()
	if err == nil && len(strings.TrimSpace(string(dirtyOutput))) > 0 {
		return Result{Skipped, "has local changes"}
	}

	fetchCmd := exec.Command("git", "-C", repoPath, "fetch", "--quiet")
	if output, err := fetchCmd.CombinedOutput(); err != nil {
		return Result{Failed, "git fetch: " + strings.TrimSpace(string(output))}
	}

	statusCmd := exec.Command("git", "-C", repoPath, "status", "--porcelain", "--branch")
	output, err := statusCmd.CombinedOutput()
	if err != nil {
		return Result{Status: UpToDate}
	}

	statusStr := string(output)
	if !strings.Contains(statusStr, "behind") {
		return Result{Status: UpToDate}
	}

	pullOut, err := exec.Command("git", "-C", repoPath, "pull", "--quiet", "--ff-only").CombinedOutput()
	if err != nil {
		return Result{Failed, "git pull: " + strings.TrimSpace(string(pullOut))}
	}
	return Result{Status: Pulled}
}
