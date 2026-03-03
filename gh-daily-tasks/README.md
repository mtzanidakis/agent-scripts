# gh-daily-tasks

GitHub daily dashboard CLI tool. Fetches open PRs, issues, dependabot alerts, and star/fork changes for your personal repos. Outputs JSON for consumption by AI agents.

Only personal (non-org) repos are included. Archived repos and forks are skipped.

## Requirements

Requires the [GitHub CLI](https://cli.github.com/) (`gh`) to be installed and authenticated:

```sh
gh auth login
```

### Token scopes

The tool uses the `gh` CLI's authentication. Your token needs these scopes:

| Scope | Grants | Required for |
|-------|--------|-------------|
| `public_repo` | Access public repositories | PRs, issues, stars/forks on public repos |
| `repo` | Full control of private repositories | PRs, issues, stars/forks on **private** repos (superset of `public_repo`) |
| `security_events` | Read and write security events | Dependabot alerts |

Use `public_repo` if you only have public repos. Use `repo` if you also need private repo data. Note that GitHub's classic OAuth scopes don't offer read-only access for private repos — `repo` is the minimum scope that covers reading PRs and issues from private repos.

To check your current scopes:

```sh
gh auth status
```

To add missing scopes:

```sh
# public repos only
gh auth refresh -s public_repo,security_events

# include private repos
gh auth refresh -s repo,security_events
```

> **Note:** If `security_events` is missing, dependabot alerts will be silently skipped (403/404 responses are ignored).

## Build

```sh
make build
```

Produces a statically linked binary `gh-daily-tasks`.

## Usage

```sh
./gh-daily-tasks
```

JSON output is printed to stdout. Warnings go to stderr.

## Environment variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GH_DAILY_DB` | Path to SQLite database file | `~/.gh-daily-tasks.db` |

The database stores star/fork snapshots for computing deltas between runs. On first run, all deltas are 0.

## Testing

```sh
make test
```
