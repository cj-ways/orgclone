"""
Config file: ~/.orgclone.yml

Example:
  default_dest: ~/Desktop

  github:
    token: ghp_xxx

  gitlab:
    token: glpat_xxx
    url: https://gitlab.com   # or your self-hosted URL

  orgs:
    playtime:
      exclude:
        - old-repo
        - test-project
    my-gitlab-group:
      exclude:
        - archived-thing
"""

import os
import yaml
from pathlib import Path
from typing import Any


CONFIG_PATH = Path.home() / ".orgclone.yml"


def load() -> dict[str, Any]:
    if not CONFIG_PATH.exists():
        return {}
    with open(CONFIG_PATH) as f:
        return yaml.safe_load(f) or {}


def get_token(platform: str) -> str | None:
    cfg = load()
    # Check env vars first
    env_map = {"github": "GITHUB_TOKEN", "gitlab": "GITLAB_TOKEN"}
    env_val = os.environ.get(env_map.get(platform, ""))
    if env_val:
        return env_val
    return cfg.get(platform, {}).get("token")


def get_gitlab_url() -> str:
    cfg = load()
    return cfg.get("gitlab", {}).get("url", "https://gitlab.com")


def get_default_dest() -> Path:
    cfg = load()
    dest = cfg.get("default_dest", "~/Desktop")
    return Path(dest).expanduser()


def get_exclusions(name: str) -> list[str]:
    cfg = load()
    return cfg.get("orgs", {}).get(name, {}).get("exclude", [])
