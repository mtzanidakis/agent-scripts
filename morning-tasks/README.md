# morning-tasks

A CLI tool that aggregates daily morning routine information into a single, compact output. Designed to be invoked by an AI agent skill, minimizing token usage through concise, structured text output.

## Tasks

| Task | Description |
|------|-------------|
| **weather** | Current conditions and daily forecast via MeteorSource API |
| **namedays** | Greek Orthodox namedays from embedded data (recurring + Easter-relative) |
| **news** | Clustered headlines from Miniflux RSS feeds with gossip filtering and deduplication |
| **offers** | Recent deals from Lagonika.gr (last 24h, auto-marked as read) |

## Usage

```
morning-tasks [-task names] [-date YYYY-MM-DD] [-list]
```

- `-task weather,news` — run only the listed tasks (default: all)
- `-date 2026-02-24` — override the current date
- `-list` — print available task names and exit

## Configuration

Environment variables (loaded via `.env` / direnv):

| Variable | Required | Default |
|----------|----------|---------|
| `METEOSOURCE_API_KEY` | yes | — |
| `MINIFLUX_API_URL` | no | `https://feeder.mtzanidakis.com` |
| `MINIFLUX_API_KEY` | yes | — |
| `WEATHER_LOCATION` | no | `cholargos` |

## Build

```
make build   # static binary, ~6 MB
make test
```

## Agent integration

See [SKILL.md](SKILL.md) for the scheduled morning routine skill definition. The binary is expected to be in `PATH`; the agent runs it, preserves the `===` section headers, and delivers the result as a single message.
