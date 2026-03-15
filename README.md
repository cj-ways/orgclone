# orgclone

Clone entire GitHub organizations or GitLab groups with one command.
Works on Linux, macOS, and Windows.

## Install

**npm (Node.js users)**
```bash
npm install -g orgclone
```

**pip (Python users)**
```bash
pip install orgclone
# or
pipx install orgclone
```

**Go users**
```bash
go install github.com/000Janela000/orgclone@latest
```

**Direct binary download** — grab the right binary from [Releases](https://github.com/000Janela000/orgclone/releases) and put it on your PATH.

---

## Usage

```bash
# GitHub org — public repos (no token needed)
orgclone clone github my-org

# GitHub org — private repos or higher rate limits
orgclone clone github my-org --token ghp_xxx
# or: export GITHUB_TOKEN=ghp_xxx

# GitLab group
orgclone clone gitlab my-group --token glpat_xxx

# Self-hosted GitLab
orgclone clone gitlab my-group --token glpat_xxx --gitlab-url https://gitlab.mycompany.com

# Custom destination
orgclone clone github my-org --dest ~/projects/my-org

# Skip archived repos
orgclone clone github my-org --skip-archived

# Exclude specific repos
orgclone clone github my-org --exclude old-repo,test-project

# Preview without cloning
orgclone clone github my-org --dry-run

# Use SSH URLs (requires SSH key set up with GitHub/GitLab)
orgclone clone github my-org --ssh
```

Running it again on an already-cloned folder will `git pull` all repos.

---

## Token-free usage

For **public repos**, no token is required. orgclone uses plain `git clone` under the hood —
if you're already authenticated via SSH keys or your system credential manager, private repos
will work too. Providing a token just embeds it in the HTTPS URL so git never prompts you.

---

## Config file

Run `orgclone init` to create `~/.orgclone.yml`:

```yaml
default_dest: ~/Desktop

github:
  token: ghp_xxx   # or use GITHUB_TOKEN env var

gitlab:
  token: glpat_xxx
  url: https://gitlab.com  # or your self-hosted URL

orgs:
  my-org:
    exclude:
      - old-repo
      - legacy-stuff
```

---

## Defaults

| Setting       | Default                    |
|---------------|----------------------------|
| Destination   | `~/Desktop/<org-name>`     |
| GitLab URL    | `https://gitlab.com`       |
| Token         | from config or env var     |
