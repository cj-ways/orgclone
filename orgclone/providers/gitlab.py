import httpx
from typing import Iterator


def list_repos(group: str, token: str | None = None, base_url: str = "https://gitlab.com") -> Iterator[dict]:
    """Yield all projects in a GitLab group (handles subgroups recursively)."""
    base_url = base_url.rstrip("/")
    headers = {}
    if token:
        headers["PRIVATE-TOKEN"] = token

    # First resolve group ID
    resp = httpx.get(f"{base_url}/api/v4/groups/{group}", headers=headers)
    resp.raise_for_status()
    group_id = resp.json()["id"]

    page = 1
    while True:
        resp = httpx.get(
            f"{base_url}/api/v4/groups/{group_id}/projects",
            headers=headers,
            params={
                "per_page": 100,
                "page": page,
                "include_subgroups": True,
                "with_shared": False,
            },
        )
        resp.raise_for_status()
        projects = resp.json()

        if not projects:
            break

        for proj in projects:
            yield {
                "name": proj["path"],
                "clone_url": proj["ssh_url_to_repo"] if token else proj["http_url_to_repo"],
                "archived": proj.get("archived", False),
                "description": proj.get("description") or "",
            }

        page += 1
