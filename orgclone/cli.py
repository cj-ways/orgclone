import typer
from pathlib import Path
from typing import Optional
from rich.console import Console
from rich.table import Table
from rich import print as rprint

from . import config, cloner
from .providers import github, gitlab

app = typer.Typer(
    name="orgclone",
    help="Clone entire GitHub orgs or GitLab groups with one command.",
    add_completion=False,
)
console = Console()


def _resolve_clone_url(repo: dict, token: str | None, use_ssh: bool) -> str:
    """
    Pick the right clone URL.
    - If user passes --ssh or has SSH keys set up, use SSH URL.
    - If token provided, use HTTPS with token embedded (works without SSH keys).
    - Otherwise plain HTTPS (works for public repos, will prompt for creds on private).
    """
    url: str = repo["clone_url"]

    # Embed token into HTTPS URL so git doesn't prompt
    if token and not use_ssh and url.startswith("https://"):
        # https://token@github.com/org/repo.git
        url = url.replace("https://", f"https://oauth2:{token}@", 1)

    return url


@app.command()
def clone(
    platform: str = typer.Argument(..., help="Platform: github or gitlab"),
    name: str = typer.Argument(..., help="Org name (GitHub) or group path (GitLab)"),
    token: Optional[str] = typer.Option(None, "--token", "-t", help="API token (or set GITHUB_TOKEN / GITLAB_TOKEN env var)"),
    dest: Optional[Path] = typer.Option(None, "--dest", "-d", help="Destination folder (default: ~/Desktop/<name>)"),
    exclude: Optional[str] = typer.Option(None, "--exclude", "-e", help="Comma-separated repo names to exclude"),
    skip_archived: bool = typer.Option(False, "--skip-archived", help="Skip archived repositories"),
    ssh: bool = typer.Option(False, "--ssh", help="Force SSH URLs (requires SSH key set up)"),
    gitlab_url: Optional[str] = typer.Option(None, "--gitlab-url", help="Self-hosted GitLab URL (e.g. https://gitlab.mycompany.com)"),
    dry_run: bool = typer.Option(False, "--dry-run", help="List repos without cloning"),
):
    """Clone all repos from a GitHub org or GitLab group."""
    platform = platform.lower()
    if platform not in ("github", "gitlab"):
        rprint("[red]Platform must be 'github' or 'gitlab'[/red]")
        raise typer.Exit(1)

    # Resolve token (CLI > env var > config file)
    resolved_token = token or config.get_token(platform)

    # Resolve destination
    base_dest = dest or (config.get_default_dest() / name)
    base_dest = base_dest.expanduser()

    # Resolve exclusions (CLI --exclude + config file)
    exclude_set = set(config.get_exclusions(name))
    if exclude:
        exclude_set.update(x.strip() for x in exclude.split(","))

    # Fetch repo list
    console.print(f"\n[bold]Fetching repos from {platform}:[/] [cyan]{name}[/]\n")
    try:
        if platform == "github":
            repos = list(github.list_repos(name, token=resolved_token))
        else:
            gl_url = gitlab_url or config.get_gitlab_url()
            repos = list(gitlab.list_repos(name, token=resolved_token, base_url=gl_url))
    except Exception as e:
        rprint(f"[red]Failed to fetch repos:[/] {e}")
        raise typer.Exit(1)

    # Filter
    filtered = []
    for repo in repos:
        if repo["name"] in exclude_set:
            continue
        if skip_archived and repo["archived"]:
            continue
        filtered.append(repo)

    if not filtered:
        rprint("[yellow]No repos found (or all filtered out).[/yellow]")
        raise typer.Exit(0)

    if dry_run:
        table = Table(title=f"{name} ({len(filtered)} repos)", show_lines=False)
        table.add_column("Repo", style="cyan")
        table.add_column("Archived", style="yellow")
        table.add_column("Description")
        for repo in filtered:
            table.add_row(repo["name"], "yes" if repo["archived"] else "", repo["description"][:60])
        console.print(table)
        raise typer.Exit(0)

    # Clone
    base_dest.mkdir(parents=True, exist_ok=True)
    console.print(f"[dim]Destination:[/] {base_dest}\n")

    results = {"cloned": 0, "pulled": 0, "up-to-date": 0, "failed": 0}

    for repo in filtered:
        clone_url = _resolve_clone_url(repo, resolved_token, ssh)
        status = cloner.clone_or_pull(repo["name"], clone_url, base_dest)

        if status == "cloned":
            rprint(f"  [green]✓[/] [cyan]{repo['name']}[/] cloned")
            results["cloned"] += 1
        elif status == "pulled":
            rprint(f"  [blue]↑[/] [cyan]{repo['name']}[/] updated")
            results["pulled"] += 1
        elif status == "up-to-date":
            rprint(f"  [dim]–[/] [dim]{repo['name']}[/] already up to date")
            results["up-to-date"] += 1
        else:
            rprint(f"  [red]✗[/] [cyan]{repo['name']}[/] {status}")
            results["failed"] += 1

    console.print(
        f"\n[bold]Done.[/] "
        f"[green]{results['cloned']} cloned[/]  "
        f"[blue]{results['pulled']} updated[/]  "
        f"[dim]{results['up-to-date']} up-to-date[/]  "
        f"[red]{results['failed']} failed[/]"
    )


@app.command()
def init():
    """Create a sample ~/.orgclone.yml config file."""
    cfg_path = Path.home() / ".orgclone.yml"
    if cfg_path.exists():
        rprint(f"[yellow]Config already exists:[/] {cfg_path}")
        raise typer.Exit(0)

    cfg_path.write_text("""\
# orgclone configuration
# Place this file at ~/.orgclone.yml

default_dest: ~/Desktop

github:
  token: ""   # or set GITHUB_TOKEN env var

gitlab:
  token: ""   # or set GITLAB_TOKEN env var
  url: https://gitlab.com  # change for self-hosted

orgs:
  my-org:
    exclude:
      - old-repo
      - test-project
""")
    rprint(f"[green]Created config:[/] {cfg_path}")


if __name__ == "__main__":
    app()
