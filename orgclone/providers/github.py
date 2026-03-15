import httpx
from typing import Iterator


def list_repos(org: str, token: str | None = None) -> Iterator[dict]:
    """Yield all repos for a GitHub org or user."""
    headers = {"Accept": "application/vnd.github+json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"

    page = 1
    while True:
        url = f"https://api.github.com/orgs/{org}/repos"
        resp = httpx.get(url, headers=headers, params={"per_page": 100, "page": page})

        if resp.status_code == 404:
            # Try as a user instead of org
            url = f"https://api.github.com/users/{org}/repos"
            resp = httpx.get(url, headers=headers, params={"per_page": 100, "page": page})

        resp.raise_for_status()
        repos = resp.json()

        if not repos:
            break

        for repo in repos:
            yield {
                "name": repo["name"],
                "clone_url": repo["ssh_url"] if token else repo["clone_url"],
                "archived": repo.get("archived", False),
                "description": repo.get("description") or "",
            }

        page += 1
