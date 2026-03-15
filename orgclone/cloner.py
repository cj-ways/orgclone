import subprocess
from pathlib import Path
from rich.console import Console

console = Console()


def clone_or_pull(name: str, clone_url: str, dest_dir: Path) -> str:
    """Clone repo if not present, pull if it already exists. Returns status string."""
    repo_path = dest_dir / name

    if repo_path.exists():
        result = subprocess.run(
            ["git", "-C", str(repo_path), "pull", "--ff-only"],
            capture_output=True,
            text=True,
        )
        if result.returncode == 0:
            msg = result.stdout.strip()
            return "up-to-date" if "Already up to date" in msg else "pulled"
        else:
            return f"pull-failed: {result.stderr.strip()}"
    else:
        result = subprocess.run(
            ["git", "clone", clone_url, str(repo_path)],
            capture_output=True,
            text=True,
        )
        if result.returncode == 0:
            return "cloned"
        else:
            return f"failed: {result.stderr.strip()}"
