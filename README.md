# agent-scripts

A collection of scripts and commands designed to be consumed by AI agents. For personal use.

## Scripts

| Script | Description |
|--------|-------------|
| [gh-daily-tasks](gh-daily-tasks/) | GitHub daily dashboard CLI. Fetches open PRs, issues, dependabot alerts, and star/fork changes for personal repos. Outputs JSON. |
| [morning-tasks](morning-tasks/) | Morning routine aggregator CLI. Fetches weather, Greek namedays, clustered news headlines, and deals into compact, agent-friendly output. |
| [youtube-summarizer](youtube-summarizer/) | Claude Code skill that summarizes YouTube videos from their transcript using yt-dlp. |

The two Go CLIs live in a single module (`github.com/mtzanidakis/agent-scripts`); each tool is a `main` package under its own directory.

## Development

The toolchain is pinned with [mise](https://mise.jdx.dev/) (`mise.toml`: Go 1.26.4, golangci-lint 2.12.2):

```bash
mise install            # install the pinned Go + golangci-lint
```

Common tasks run through the root `Makefile` (both binaries build into `bin/`):

```bash
make build              # -> bin/gh-daily-tasks, bin/morning-tasks
make test               # go test ./...
make vet                # go vet ./...
make fmt                # gofmt -w .
make lint               # golangci-lint run
make tidy               # go mod tidy
make clean
```

## Configuration

Secrets and per-tool environment variables are read from the process environment. Locally they live in `mise.local.toml`, which is gitignored and loaded automatically by mise:

```toml
[env]
METEOSOURCE_API_KEY = "..."   # morning-tasks: weather
MINIFLUX_API_URL = "..."      # morning-tasks: news + offers feeds
MINIFLUX_API_KEY = "..."
```

`gh-daily-tasks` authenticates through the GitHub CLI (`gh auth`) and stores its snapshot database at `~/.gh-daily-tasks.db` (override with `GH_DAILY_DB`). See each tool's own README for details.

## Releases

Tagged versions (`v*`) trigger a [GoReleaser](https://goreleaser.com/) build on GitHub Actions, producing per-binary archives for Linux/macOS/Windows × amd64/arm64. Push a tag to cut a release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

`go test` and golangci-lint also run on every push and pull request to `main`.
