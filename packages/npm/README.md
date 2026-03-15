# orgclone

Clone entire GitHub organizations or GitLab groups with one command.

[![npm](https://img.shields.io/npm/v/@cj-ways/orgclone)](https://www.npmjs.com/package/@cj-ways/orgclone)
[![license](https://img.shields.io/github/license/cj-ways/orgclone)](https://github.com/cj-ways/orgclone/blob/master/LICENSE)

```bash
npm install -g @cj-ways/orgclone
```

---

## Quick Start

```bash
# Clone a GitHub org (GitHub is the default)
orgclone clone my-org

# Interactively pick which repos to clone
orgclone clone my-org --pick

# Clone a GitLab group
orgclone clone my-group --gitlab

# Preview what would be cloned
orgclone clone my-org --dry-run

# Change defaults permanently
orgclone default platform gitlab
orgclone default dest ~/projects
```

Repos land in `~/Desktop/my-org/` by default. Run it again on the same folder and it `git pull`s everything — no re-cloning.

---

## Install

| Method | Command |
|--------|---------|
| **npm** | `npm install -g @cj-ways/orgclone` |
| **pip** | `pip install orgclone` |
| **Go** | `go install github.com/cj-ways/orgclone@latest` |
| **Binary** | [Download from Releases](https://github.com/cj-ways/orgclone/releases) |

---

## All Options

```
orgclone clone <name> [options]

  name                  org name (GitHub) or group path (GitLab)

  --gitlab              Use GitLab instead of the default platform
  --pick                Interactively select which repos to clone
  -t, --token           API token (or set GITHUB_TOKEN / GITLAB_TOKEN)
  -d, --dest            Destination folder (default: ~/Desktop/<name>)
  -e, --exclude         Comma-separated repo names to skip
      --skip-archived   Skip archived repositories
      --gitlab-url      Self-hosted GitLab URL (default: https://gitlab.com)
      --dry-run         List repos without cloning

orgclone default <setting> <value>

  platform    github or gitlab
  dest        path to clone into (e.g. ~/projects)

  Examples:
    orgclone default platform gitlab
    orgclone default dest ~/projects
```

---

## No token? No problem

For **public repos**, no token needed — orgclone calls `git clone` the same way you would manually. If you already have SSH keys or a credential manager set up (you can `git push` without a password), **private repos work too** with no extra config.

A token only adds value when you want auth to work automatically on a fresh machine or in CI.

---

## Why orgclone over alternatives?

| Feature | orgclone | others |
|---------|----------|--------|
| GitHub support | ✓ | ✓ |
| GitLab support | ✓ | ✗ |
| Self-hosted GitLab | ✓ | ✗ |
| Auto-update on re-run | ✓ | some |
| Config file | ✓ | ✗ |
| Dry run | ✓ | ✗ |
| Skip archived | ✓ | ✗ |
| npm + pip + Go install | ✓ | ✗ |
| No runtime dependency | ✓ (single binary) | ✗ |

---

## Config file

Run `orgclone init` to generate `~/.orgclone.yml`:

```yaml
default_dest: ~/Desktop
default_platform: github   # or gitlab

github:
  token: ghp_xxx        # or: export GITHUB_TOKEN=ghp_xxx

gitlab:
  token: glpat_xxx
  url: https://gitlab.com   # change for self-hosted

orgs:
  my-org:
    exclude:
      - old-repo
      - scratch
```

CLI flags always override config file values.

---

## Links

- [GitHub Repository](https://github.com/cj-ways/orgclone)
- [Report an Issue](https://github.com/cj-ways/orgclone/issues)
- [Releases & Binaries](https://github.com/cj-ways/orgclone/releases)

## License

MIT
