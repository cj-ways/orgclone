# orgclone

Clone entire GitHub organizations or GitLab groups with one command.

## Install

```bash
pip install orgclone
# or
pipx install orgclone
```

## Usage

```bash
# GitHub org — public (no token needed for public repos)
orgclone clone github my-org

# GitHub org — with token (required for private repos, higher rate limits)
orgclone clone github my-org --token ghp_xxx
# or set env var: export GITHUB_TOKEN=ghp_xxx

# GitLab group
orgclone clone gitlab my-group --token glpat_xxx

# Self-hosted GitLab
orgclone clone gitlab my-group --token glpat_xxx --gitlab-url https://gitlab.mycompany.com

# Custom destination folder
orgclone clone github my-org --dest ~/projects/my-org

# Skip archived repos
orgclone clone github my-org --skip-archived

# Exclude specific repos
orgclone clone github my-org --exclude old-repo,test-project

# Preview without cloning
orgclone clone github my-org --dry-run

# Force SSH (requires SSH keys set up with GitHub/GitLab)
orgclone clone github my-org --ssh
```

Running again on an already-cloned folder will `git pull` all repos to keep them up to date.

## Token-free usage

For **public repos**, no token is needed. orgclone uses HTTPS clone URLs and git handles authentication the same way any `git clone` would — using your system credential manager, SSH keys, or `.netrc`.

For **private repos**, provide a token. orgclone embeds it into the HTTPS URL so git doesn't prompt you for credentials.

## Config file

Run `orgclone init` to create `~/.orgclone.yml`:

```yaml
default_dest: ~/Desktop

github:
  token: ghp_xxx   # optional — or use GITHUB_TOKEN env var

gitlab:
  token: glpat_xxx
  url: https://gitlab.com

orgs:
  my-org:
    exclude:
      - old-repo
      - legacy-stuff
```

## Defaults

| Setting | Default |
|---|---|
| Destination | `~/Desktop/<org-name>` |
| Platform | — (you must specify `github` or `gitlab`) |
| Token | from config or env var |
| GitLab URL | `https://gitlab.com` |
